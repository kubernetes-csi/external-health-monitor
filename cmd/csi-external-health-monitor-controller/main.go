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
	"net/http"
	"os"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apiserver/pkg/server"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/component-base/featuregate"
	"k8s.io/component-base/logs"
	logsapi "k8s.io/component-base/logs/api/v1"
	_ "k8s.io/component-base/logs/json/register"
	"k8s.io/klog/v2"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/connection"
	"github.com/kubernetes-csi/csi-lib-utils/leaderelection"
	"github.com/kubernetes-csi/csi-lib-utils/metrics"
	"github.com/kubernetes-csi/csi-lib-utils/rpc"
	"github.com/kubernetes-csi/csi-lib-utils/standardflags"
	"google.golang.org/grpc"

	monitorcontroller "github.com/kubernetes-csi/external-health-monitor/pkg/controller"
	"github.com/kubernetes-csi/external-health-monitor/pkg/features"
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
)

var (
	version = "unknown"
)

func main() {
	fg := featuregate.NewFeatureGate()
	logsapi.AddFeatureGates(fg)
	c := logsapi.NewLoggingConfiguration()
	logsapi.AddGoFlags(c, flag.CommandLine)
	logs.InitLogs()
	standardflags.AddAutomaxprocs(klog.Infof)
	flag.Parse()
	logger := klog.Background()
	if err := logsapi.ValidateAndApply(c, fg); err != nil {
		logger.Error(err, "LoggingConfiguration is invalid")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	if *showVersion {
		fmt.Println(os.Args[0], version)
		return
	}
	logger.Info("Version", "version", version)

	if *metricsAddress != "" && *httpEndpoint != "" {
		logger.Error(nil, "Only one of `--metrics-address` and `--http-endpoint` can be set.")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	addr := *metricsAddress
	if addr == "" {
		addr = *httpEndpoint
	}

	// Create the client config. Use kubeconfig if given, otherwise assume in-cluster.
	config, err := buildConfig(*kubeconfig)
	if err != nil {
		logger.Error(err, "Failed to build a Kubernetes config")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	if *workerThreads == 0 {
		logger.Error(nil, "Option --worker-threads must be greater than zero")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error(err, "Failed to create a Clientset")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	factory := informers.NewSharedInformerFactory(clientset, *resync)

	metricsManager := metrics.NewCSIMetricsManager("" /* driverName */)

	// Connect to CSI.
	ctx := context.Background()
	csiConn, err := connection.Connect(ctx, *csiAddress, metricsManager, connection.OnConnectionLoss(connection.ExitOnConnectionLoss()))
	if err != nil {
		logger.Error(err, "Failed to connect to the CSI driver")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	err = rpc.ProbeForever(ctx, csiConn, *timeout)
	if err != nil {
		logger.Error(err, "Failed to probe the CSI driver")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	// Find driver name.
	cancelationCtx, cancel := context.WithTimeout(context.Background(), csiTimeout)
	cancelationCtx = klog.NewContext(cancelationCtx, logger)
	defer cancel()
	storageDriver, err := rpc.GetDriverName(cancelationCtx, csiConn)
	if err != nil {
		logger.Error(err, "Failed to get the CSI driver name")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	logger.V(2).Info("CSI driver name", "driver", storageDriver)
	metricsManager.SetDriverName(storageDriver)

	// Prepare HTTP endpoint for metrics + leader election healthz
	mux := http.NewServeMux()
	if addr != "" {
		metricsManager.RegisterToServer(mux, *metricsPath)
		go func() {
			logger.Info("ServeMux listening", "address", addr)
			err := http.ListenAndServe(addr, mux)
			if err != nil {
				logger.Error(err, "Failed to start HTTP server at specified address and metrics path", "address", addr, "path", *metricsPath)
				klog.FlushAndExit(klog.ExitFlushTimeout, 1)
			}
		}()
	}

	supportsService, err := supportsPluginControllerService(cancelationCtx, csiConn)
	if err != nil {
		logger.Error(err, "Failed to check whether the CSI driver supports the Plugin Controller Service")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	if !supportsService {
		logger.V(2).Info("CSI driver does not support Plugin Controller Service, exiting")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	supportControllerListVolumes, err := supportControllerListVolumes(cancelationCtx, csiConn)
	if err != nil {
		logger.Error(err, "Failed to check whether the CSI driver supports the Controller Service ListVolumes")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	supportControllerGetVolume, err := supportControllerGetVolume(cancelationCtx, csiConn)
	if err != nil {
		logger.Error(err, "Failed to check whether the CSI driver supports the Controller Service GetVolume")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	supportControllerVolumeCondition, err := supportControllerVolumeCondition(cancelationCtx, csiConn)
	if err != nil {
		logger.Error(err, "Failed to check whether the CSI driver supports the Controller Service VolumeCondition")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	if (!supportControllerListVolumes && !supportControllerGetVolume) || !supportControllerVolumeCondition {
		logger.V(2).Info("CSI driver does not support Controller ListVolumes and GetVolume service or does not implement VolumeCondition, exiting")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
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

	broadcaster := record.NewBroadcaster(record.WithContext(ctx))
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: clientset.CoreV1().Events(v1.NamespaceAll)})
	eventRecorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: fmt.Sprintf("csi-pv-monitor-controller-%s", option.DriverName)}).WithLogger(logger)

	monitorController := monitorcontroller.NewPVMonitorController(
		logger,
		clientset,
		csiConn,
		factory,
		eventRecorder,
		&option,
	)

	// handle SIGTERM and SIGINT by cancelling the context.

	var (
		terminate       func()          // called when all controllers are finished
		controllerCtx   context.Context // shuts down all controllers on a signal
		shutdownHandler <-chan struct{} // called when the signal is received
	)

	if utilfeature.DefaultFeatureGate.Enabled(features.ReleaseLeaderElectionOnExit) {
		// ctx waits for all controllers to finish, then shuts down the whole process, incl. leader election
		ctx, terminate = context.WithCancel(ctx)
		var cancelControllerCtx context.CancelFunc
		controllerCtx, cancelControllerCtx = context.WithCancel(ctx)
		shutdownHandler = server.SetupSignalHandler()

		defer terminate()

		go func() {
			defer cancelControllerCtx()
			<-shutdownHandler
			logger.Info("Received SIGTERM or SIGINT signal, shutting down controller.")
		}()
	}

	run := func(ctx context.Context) {
		if utilfeature.DefaultFeatureGate.Enabled(features.ReleaseLeaderElectionOnExit) {
			var wg sync.WaitGroup
			stopCh := controllerCtx.Done()
			factory.Start(stopCh)
			monitorController.Run(controllerCtx, int(*workerThreads), &wg)
		} else {
			stopCh := ctx.Done()
			factory.Start(stopCh)
			monitorController.Run(ctx, int(*workerThreads), nil)
		}
	}

	if !*enableLeaderElection {
		run(ctx)
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
		le.WithContext(ctx)
		if utilfeature.DefaultFeatureGate.Enabled(features.ReleaseLeaderElectionOnExit) {
			le.WithReleaseOnCancel(true)
		}

		// TODO: The broadcaster and eventRecorder in the leaderelection package
		// within csi-lib-utils do not support contextual logging.
		// To fully support contextual logging in external-health-monitor,
		// an upgrade of csi-lib-utils version will be necessary
		// after contextual logging support is added to csi-lib-utils.
		// https://github.com/kubernetes-sigs/sig-storage-lib-external-provisioner/pull/171
		if err := le.Run(); err != nil {
			logger.Error(err, "Failed to initialize leader election")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
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
