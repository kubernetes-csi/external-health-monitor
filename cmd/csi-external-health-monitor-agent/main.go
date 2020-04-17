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
	"os"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/connection"
	"github.com/kubernetes-csi/csi-lib-utils/metrics"
	"github.com/kubernetes-csi/csi-lib-utils/rpc"
	"google.golang.org/grpc"

	monitoragent "github.com/kubernetes-csi/external-health-monitor/pkg/agent"
)

const (

	// Default timeout of short CSI calls like GetPluginInfo
	csiTimeout = time.Second
)

// Command line flags
var (
	monitorInterval = flag.Duration("monitor-interval", 1*time.Minute, "Interval for controller to check volumes health condition.")

	kubeconfig      = flag.String("kubeconfig", "", "Absolute path to the kubeconfig file. Required only when running out of cluster.")
	resync          = flag.Duration("resync", 10*time.Minute, "Resync interval of the controller.")
	csiAddress      = flag.String("csi-address", "/run/csi/socket", "Address of the CSI driver socket.")
	showVersion     = flag.Bool("version", false, "Show version.")
	timeout         = flag.Duration("timeout", 15*time.Second, "Timeout for waiting for attaching or detaching the volume.")
	workerThreads   = flag.Uint("worker-threads", 10, "Number of pv monitor worker threads")
	kubeletRootPath = flag.String("kubelet-root-path", "/var/lib/kubelet", "The root path of kubelet.")

	metricsAddress = flag.String("metrics-address", "", "The TCP network address where the prometheus metrics endpoint will listen (example: `:8080`). The default is empty string, which means metrics endpoint is disabled.")
	metricsPath    = flag.String("metrics-path", "/metrics", "The HTTP path where prometheus metrics will be exposed. Default is `/metrics`.")
)

var (
	version = "unknown"
)

func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()

	if *showVersion {
		fmt.Println(os.Args[0], version)
		return
	}
	klog.Infof("Version: %s", version)

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
	metricsManager.StartMetricsEndpoint(*metricsAddress, *metricsPath)

	supportNodeGetVolumeCondition, err := supportNodeGetVolumeCondition(ctx, csiConn)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}
	if !supportNodeGetVolumeCondition {
		klog.V(2).Infof("CSI driver does not support Node VolumeCondition Capability, exiting")
		os.Exit(1)
	}

	supportStageUnstage, err := supportStageUnstage(ctx, csiConn)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	monitorAgent := monitoragent.NewPVMonitorAgent(clientset, storageDriver, csiConn, *timeout, *monitorInterval, factory.Core().V1().PersistentVolumes(),
		factory.Core().V1().PersistentVolumeClaims(), factory.Core().V1().Pods(), supportStageUnstage, *kubeletRootPath)

	run := func(ctx context.Context) {
		stopCh := ctx.Done()
		factory.Start(stopCh)
		monitorAgent.Run(int(*workerThreads), stopCh)
	}

	run(context.TODO())
}

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func supportNodeGetVolumeCondition(ctx context.Context, csiConn *grpc.ClientConn) (supportNodeGetVolumeCondition bool, err error) {
	client := csi.NewNodeClient(csiConn)
	req := csi.NodeGetCapabilitiesRequest{}
	rsp, err := client.NodeGetCapabilities(ctx, &req)
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
		if t == csi.NodeServiceCapability_RPC_VOLUME_CONDITION {
			return true, nil
		}
	}

	return false, nil
}

func supportStageUnstage(ctx context.Context, csiConn *grpc.ClientConn) (supportStageUnstage bool, err error) {
	client := csi.NewNodeClient(csiConn)
	req := csi.NodeGetCapabilitiesRequest{}
	rsp, err := client.NodeGetCapabilities(ctx, &req)
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
		if t == csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME {
			return true, nil
		}
	}

	return false, nil
}
