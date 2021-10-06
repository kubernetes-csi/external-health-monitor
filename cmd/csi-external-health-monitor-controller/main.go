/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kubernetes-csi/external-provisioner/pkg/capacity/topology"
	"net/http"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/connection"
	"github.com/kubernetes-csi/csi-lib-utils/leaderelection"
	"github.com/kubernetes-csi/csi-lib-utils/metrics"
	"github.com/kubernetes-csi/csi-lib-utils/rpc"
	"google.golang.org/grpc"

	monitorcontroller "github.com/kubernetes-csi/external-health-monitor/pkg/controller"
)

const (

	// Default timeout of short CSI calls like GetPluginInfo
	csiTimeout = time.Second
)

// Command line flags
var (
	monitorInterval = flag.Duration("monitor-interval", 1*time.Minute, "Interval for controller to check volumes health condition.")

	kubeconfig               = flag.String("kubeconfig", "", "Absolute path to the kubeconfig file. Required only when running out of cluster.")
	resync                   = flag.Duration("resync", 10*time.Minute, "Resync interval of the controller.")
	csiAddress               = flag.String("csi-address", "/run/csi/socket", "Address of the CSI driver socket.")
	showVersion              = flag.Bool("version", false, "Show version.")
	timeout                  = flag.Duration("timeout", 15*time.Second, "Timeout for waiting for attaching or detaching the volume.")
	listVolumesInterval      = flag.Duration("list-volumes-interval", 5*time.Minute, "Time interval for calling ListVolumes RPC to check volumes' health condition")
	volumeListAndAddInterval = flag.Duration("volume-list-add-interval", 5*time.Minute, "Time interval for listing volumes and add them to queue")
	nodeListAndAddInterval   = flag.Duration("node-list-add-interval", 5*time.Minute, "Time interval for listing nodess and add them to queue")
	workerThreads            = flag.Uint("worker-threads", 10, "Number of pv monitor worker threads")
	enableNodeWatcher        = flag.Bool("enable-node-watcher", false, "Indicates whether the node watcher is enabled or not.")

	enableLeaderElection        = flag.Bool("leader-election", false, "Enable leader election.")
	leaderElectionNamespace     = flag.String("leader-election-namespace", "", "Namespace where the leader election resource lives. Defaults to the pod namespace if not set.")
	leaderElectionLeaseDuration = flag.Duration("leader-election-lease-duration", 15*time.Second, "Duration, in seconds, that non-leader candidates will wait to force acquire leadership. Defaults to 15 seconds.")
	leaderElectionRenewDeadline = flag.Duration("leader-election-renew-deadline", 10*time.Second, "Duration, in seconds, that the acting leader will retry refreshing leadership before giving up. Defaults to 10 seconds.")
	leaderElectionRetryPeriod   = flag.Duration("leader-election-retry-period", 5*time.Second, "Duration, in seconds, the LeaderElector clients should wait between tries of actions. Defaults to 5 seconds.")

	metricsAddress = flag.String("metrics-address", "", "(deprecated) The TCP network address where the prometheus metrics endpoint will listen (example: `:8080`). The default is empty string, which means metrics endpoint is disabled. Only one of `--metrics-address` and `--http-endpoint` can be set.")
	httpEndpoint   = flag.String("http-endpoint", "", "The TCP network address where the HTTP server for diagnostics, including metrics and leader election health check, will listen (example: `:8080`). The default is empty string, which means the server is disabled. Only one of `--metrics-address` and `--http-endpoint` can be set.")
	metricsPath    = flag.String("metrics-path", "/metrics", "The HTTP path where prometheus metrics will be exposed. Default is `/metrics`.")
	enableNodeDeployment = flag.Bool("node-deployment", false, "Enable deploying the sidecar controller together with a CSI driver on nodes to manage snapshots for node-local volumes. Off by default.")
)

var (
	version = "unknown"
)

func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()

	node := os.Getenv("NODE_NAME")
	if *enableNodeDeployment && node == "" {
		klog.Fatal("The NODE_NAME environment variable must be set when using --enable-node-deployment.")
	}

	if *showVersion {
		fmt.Println(os.Args[0], version)
		return
	}
	klog.Infof("Version: %s", version)

	if *metricsAddress != "" && *httpEndpoint != "" {
		klog.Error("only one of `--metrics-address` and `--http-endpoint` can be set.")
		os.Exit(1)
	}
	addr := *metricsAddress
	if addr == "" {
		addr = *httpEndpoint
	}

	// Create the client config. Use kubeconfig if given, otherwise assume in-cluster.
	config, err := buildConfig(*kubeconfig)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	if *workerThreads == 0 {
		klog.Error("option -worker-threads must be greater than zero")
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	factory := informers.NewSharedInformerFactory(clientset, *resync)

	metricsManager := metrics.NewCSIMetricsManager("" /* driverName */)

	// Connect to CSI.
	csiConn, err := connection.Connect(*csiAddress, metricsManager, connection.OnConnectionLoss(connection.ExitOnConnectionLoss()))
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	err = rpc.ProbeForever(csiConn, *timeout)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	// Find driver name.
	ctx, cancel := context.WithTimeout(context.Background(), csiTimeout)
	defer cancel()
	storageDriver, err := rpc.GetDriverName(ctx, csiConn)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}
	klog.V(2).Infof("CSI driver name: %q", storageDriver)
	metricsManager.SetDriverName(storageDriver)

	// Prepare HTTP endpoint for metrics + leader election healthz
	mux := http.NewServeMux()
	if addr != "" {
		metricsManager.RegisterToServer(mux, *metricsPath)
		go func() {
			klog.Infof("ServeMux listening at %q", addr)
			err := http.ListenAndServe(addr, mux)
			if err != nil {
				klog.Fatalf("Failed to start HTTP server at specified address (%q) and metrics path (%q): %s", addr, *metricsPath, err)
			}
		}()
	}

	supportsService, err := supportsPluginControllerService(ctx, csiConn)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}
	if !supportsService {
		klog.V(2).Infof("CSI driver does not support Plugin Controller Service, exiting")
		os.Exit(1)
	}

	supportControllerListVolumes, err := supportControllerListVolumes(ctx, csiConn)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	supportControllerGetVolume, err := supportControllerGetVolume(ctx, csiConn)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	supportControllerVolumeCondition, err := supportControllerVolumeCondition(ctx, csiConn)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	if (!supportControllerListVolumes && !supportControllerGetVolume) || !supportControllerVolumeCondition {
		klog.V(2).Infof("CSI driver does not support Controller ListVolumes and GetVolume service or does not implement VolumeCondition, exiting")
		os.Exit(1)
	}

	option := monitorcontroller.PVMonitorOptions{
		DriverName:        storageDriver,
		ContextTimeout:    *timeout,
		EnableNodeWatcher: *enableNodeWatcher,
		SupportListVolume: supportControllerListVolumes,

		ListVolumesInterval:      *listVolumesInterval,
		PVWorkerExecuteInterval:  *monitorInterval,
		VolumeListAndAddInterval: *volumeListAndAddInterval,

		NodeWorkerExecuteInterval: *monitorInterval,
		NodeListAndAddInterval:    *nodeListAndAddInterval,
	}

	broadcaster := record.NewBroadcaster()
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: clientset.CoreV1().Events(v1.NamespaceAll)})
	eventRecorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: fmt.Sprintf("csi-pv-monitor-controller-%s", option.DriverName)})

	monitorController := monitorcontroller.NewPVMonitorController(clientset, csiConn, factory.Core().V1().PersistentVolumes(),
		factory.Core().V1().PersistentVolumeClaims(), factory.Core().V1().Pods(), factory.Core().V1().Nodes(), factory.Core().V1().Events(), eventRecorder, &option)

	run := func(ctx context.Context) {
		stopCh := ctx.Done()
		factory.Start(stopCh)
		monitorController.Run(int(*workerThreads), stopCh)
	}

	var topologyInformer topology.Informer
	if nodeDeployment == nil {
		topologyInformer = topology.NewNodeTopology(
			provisionerName,
			clientset,
			factory.Core().V1().Nodes(),
			factory.Storage().V1().CSINodes(),
			workqueue.NewNamedRateLimitingQueue(rateLimiter, "csitopology"),
		)
	} else {
		var segment topology.Segment
		if nodeDeployment.NodeInfo.AccessibleTopology != nil {
			for key, value := range nodeDeployment.NodeInfo.AccessibleTopology.Segments {
				segment = append(segment, topology.SegmentEntry{Key: key, Value: value})
			}
		}
		klog.Infof("producing CSIStorageCapacity objects with fixed topology segment %s", segment)
		topologyInformer = topology.NewFixedNodeTopology(&segment)
	}
	go topologyInformer.RunWorker(context.Background())


	if !*enableLeaderElection {
		run(context.TODO())
	} else {
		// Name of config map with leader election lock
		lockName := "external-health-monitor-leader-" + storageDriver
		le := leaderelection.NewLeaderElection(clientset, lockName, run)
		if *httpEndpoint != "" {
			le.PrepareHealthCheck(mux, leaderelection.DefaultHealthCheckTimeout)
		}

		if *leaderElectionNamespace != "" {
			le.WithNamespace(*leaderElectionNamespace)
		}

		le.WithLeaseDuration(*leaderElectionLeaseDuration)
		le.WithRenewDeadline(*leaderElectionRenewDeadline)
		le.WithRetryPeriod(*leaderElectionRetryPeriod)

		if err := le.Run(); err != nil {
			klog.Fatalf("failed to initialize leader election: %v", err)
		}
	}

}

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func supportControllerListVolumes(ctx context.Context, csiConn *grpc.ClientConn) (supportControllerListVolumes bool, err error) {
	caps, err := rpc.GetControllerCapabilities(ctx, csiConn)
	if err != nil {
		return false, fmt.Errorf("failed to get controller capabilities: %v", err)
	}

	return caps[csi.ControllerServiceCapability_RPC_LIST_VOLUMES], nil
}

// TODO: move this to csi-lib-utils
func supportControllerGetVolume(ctx context.Context, csiConn *grpc.ClientConn) (supportControllerGetVolume bool, err error) {
	client := csi.NewControllerClient(csiConn)
	req := csi.ControllerGetCapabilitiesRequest{}
	rsp, err := client.ControllerGetCapabilities(ctx, &req)
	if err != nil {
		return false, err
	}

	for _, cap := range rsp.GetCapabilities() {
		if cap == nil {
			continue
		}
		rpc := cap.GetRpc()
		if rpc == nil {
			continue
		}
		t := rpc.GetType()
		if t == csi.ControllerServiceCapability_RPC_GET_VOLUME {
			return true, nil
		}
	}

	return false, nil
}

// TODO: move this to csi-lib-utils
func supportControllerVolumeCondition(ctx context.Context, csiConn *grpc.ClientConn) (supportControllerVolumeCondition bool, err error) {
	client := csi.NewControllerClient(csiConn)
	req := csi.ControllerGetCapabilitiesRequest{}
	rsp, err := client.ControllerGetCapabilities(ctx, &req)
	if err != nil {
		return false, err
	}

	for _, cap := range rsp.GetCapabilities() {
		if cap == nil {
			continue
		}

		rpc := cap.GetRpc()
		if rpc == nil {
			continue
		}
		t := rpc.GetType()
		if t == csi.ControllerServiceCapability_RPC_VOLUME_CONDITION {
			return true, nil
		}
	}

	return false, nil
}

// TODO: move this to csi-lib-utils
func supportsPluginControllerService(ctx context.Context, csiConn *grpc.ClientConn) (bool, error) {
	client := csi.NewIdentityClient(csiConn)
	req := csi.GetPluginCapabilitiesRequest{}
	rsp, err := client.GetPluginCapabilities(ctx, &req)
	if err != nil {
		return false, err
	}
	for _, cap := range rsp.GetCapabilities() {
		if cap == nil {
			continue
		}
		srv := cap.GetService()
		if srv == nil {
			continue
		}
		t := srv.GetType()
		if t == csi.PluginCapability_Service_CONTROLLER_SERVICE {
			return true, nil
		}
	}

	return false, nil
}
