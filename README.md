# Volume Health Monitor

The Volume Health Monitor is part of Kubernetes implementation of [Container Storage Interface (CSI)](https://github.com/container-storage-interface/spec). It was introduced as an Alpha feature in Kubernetes v1.19. In Kubernetes 1.21, a second Alpha was done due to a design change.

## Overview

The [Volume Health Monitor](https://github.com/kubernetes/enhancements/tree/master/keps/sig-storage/1432-volume-health-monitor) is implemented in two components: `External Health Monitor Controller` and `Kubelet`.

When this feature was first introduced in Kubernetes 1.19, there was an `External Health Monitor Agent` that monitors volume health from the node side. In the Kubernetes 1.21 release, the node side volume health monitoring logic was moved to Kubelet to avoid duplicate CSI RPC calls.

- External Health Monitor Controller:
  - The external health monitor controller will be deployed as a sidecar together with the CSI controller driver, similar to how the external-provisioner sidecar is deployed.
  - Trigger controller RPC to check the health condition of the CSI volumes.
  - The external controller sidecar will also watch for node failure events. This component can be enabled via a flag.

- Kubelet:
  - In addition to existing volume stats collected already, Kubelet will also check volume's mounting conditions collected from the same CSI node RPC and log events to Pods if volume condition is abnormal.

The Volume Health Monitoring feature need to invoke the following CSI interfaces.

- External Health Monitor Controller:
  - ListVolumes (If both `ListVolumes` and `ControllerGetVolume` are supported, `ListVolumes` will be used)
  - ControllerGetVolume
- Kubelet:
  - NodeGetVolumeStats
  - This feature in Kubelet is controlled by an Alpha feature gate `CSIVolumeHealth`.

## Compatibility

This information reflects the head of this branch.

| Compatible with CSI Version                                                                | Container Image             |
| ------------------------------------------------------------------------------------------ | ----------------------------|
| [CSI Spec v1.3.0](https://github.com/container-storage-interface/spec/releases/tag/v1.3.0) | registry.k8s.io/sig-storage.csi-external-health-monitor-controller |

## Driver Support

Currently, the CSI volume health monitoring interfaces are only implemented in the Mock Driver and the CSI Hostpath driver.

## Usage

External Health Monitor Controller needs to be deployed with CSI driver.

Alpha feature gate `CSIVolumeHealth` needs to be enabled for the node side monitoring to take effect.

### Build && Push Image

You can run the command below in the root directory of the project.

```bash
make container GOFLAGS_VENDOR=$( [ -d vendor ] && echo '-mod=vendor' )
```

And then, you can tag and push the csi-external-health-monitor-controller image to your own image repository.

```bash
docker tag csi-external-health-monitor-controller:latest <custom-image-repo-addr>/csi-external-health-monitor-controller:<custom-image-tag>
```

### External Health Monitor Controller

```bash
cd external-health-monitor
kubectl create -f deploy/kubernetes/external-health-monitor-controller
```

You can run `kubectl get pods` command to confirm if they are deployed on your cluster successfully.

Check logs of external health monitor controller as follows:

-  `kubectl logs <leader-of-external-health-monitor-controller-container-name> -c csi-external-health-monitor-controller`

Check if there are events on PVCs or Pods that report abnormal volume condition when the volume you are using is abnormal.

## csi-external-health-monitor-controller-sidecar-command-line-options

### Important optional arguments that are highly recommended to be used

- `leader-election`: Enables leader election. This is useful when there are multiple replicas of the same external-health-monitor-controller running for one CSI driver. Only one of them may be active (=leader). A new leader will be re-elected when the current leader dies or becomes unresponsive for ~15 seconds.

- `leader-election-namespace <namespace>`: The namespace where the leader election resource exists. Defaults to the pod namespace if not set.

- `leader-election-lease-duration <duration>`: Duration, in seconds, that non-leader candidates will wait to force acquire leadership. Defaults to 15 seconds.

- `leader-election-renew-deadline <duration>`: Duration, in seconds, that the acting leader will retry refreshing leadership before giving up. Defaults to 10 seconds.

- `leader-election-retry-period <duration>`: Duration, in seconds, the LeaderElector clients should wait between tries of actions. Defaults to 5 seconds.

- `http-endpoint`: The TCP network address where the HTTP server for diagnostics, including metrics and leader election health check, will listen (example: `:8080` which corresponds to port 8080 on local host). The default is empty string, which means the server is disabled.

- `metrics-path`: The HTTP path where prometheus metrics will be exposed. Default is /metrics.

- `worker-threads`: Number of worker threads for running volume checker when CSI Driver supports `ControllerGetVolume`, but not `ListVolumes`. The default value is 10.

### Other recognized arguments

- `kubeconfig <path>`: Path to Kubernetes client configuration that the external-health-monitor-controller uses to connect to the Kubernetes API server. When omitted, default token provided by Kubernetes will be used. This option is useful only when the external-health-monitor-controller does not run as a Kubernetes pod, e.g. for debugging.

- `resync <duration>`: Internal resync interval when the monitor controller re-evaluates all existing resource objects that it was watching and tries to fulfill them. It does not affect re-tries of failed calls! It should be used only when there is a bug in Kubernetes watch logic. The default is ten mintiues.

- `csiAddress <path-to-csi>`: This is the path to the CSI Driver socket inside the pod that the external-health-monitor-controller container will use to issue CSI operations (/run/csi/socket is used by default).

- `version`: Prints the current version of external-health-monitor-controller.

- `timeout <duration>`: Timeout of all calls to CSI Driver. It should be set to value that accommodates the majority of `ListVolumes`, `ControllerGetVolume` calls. 15 seconds is used by default.

- `list-volumes-interval <duration>`: Interval of monitoring volume health condition by invoking the RPC interface of `ListVolumes`. You can adjust it to change the frequency of the evaluation process. Five mintiues by default if not set.

- `enable-node-watcher <boolean>`: Enable node-watcher. node-watcher evaluates volume health condition by checking node status periodically.

- `monitor-interval <duration>`: Interval of monitoring volume health condition when CSI Driver supports `ControllerGetVolume`, but not `ListVolumes`. It is also used by nodeWatcher. You can adjust it to change the frequency of the evaluation process. One minute by default if not set.

- `volume-list-add-interval <duration>`: Interval of listing volumes and adding them to the queue when CSI driver supports `ControllerGetVolume`, but not `ListVolumes`.

- `node-list-add-interval <duration>`: Interval of listing nodes and adding them. It is used together with `monitor-interval` and `enable-node-watcher` by nodeWatcher.

- `metrics-address`: (deprecated) The TCP network address where the Prometheus metrics endpoint will run (example: :8080, which corresponds to port 8080 on local host). The default is the empty string, which means the metrics and leader election check endpoint is disabled.

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack](https://kubernetes.slack.com/messages/sig-storage)
- [Mailing List](https://groups.google.com/forum/#!forum/kubernetes-sig-storage)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
