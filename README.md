# Volume Health Monitor

The external health monitor controller is a [CSI](https://github.com/container-storage-interface/spec) sidecar that watches the volumes of a CSI driver, periodically checks their health through the driver's controller service, and records adverse conditions on the `PersistentVolumeClaim.Status.HealthStatus` field of the bound claims. It implements the controller side of [KEP-1432](https://github.com/kubernetes/enhancements/tree/master/keps/sig-storage/1432-volume-health-monitor).

## Overview

The sidecar is deployed together with the CSI controller driver, similar to how the external-provisioner sidecar is deployed.

A CSI driver opts in by advertising controller service capabilities:

- `GET_VOLUME_HEALTH`: the sidecar polls each bound volume individually with `ControllerGetVolumeHealth`.
- `LIST_VOLUME_HEALTH`: the sidecar retrieves health for all volumes in one paginated `ControllerListVolumeHealth` pass per cycle. This mode is preferred when advertised. Per the CSI specification, a driver that advertises `LIST_VOLUME_HEALTH` must also advertise `GET_VOLUME_HEALTH`, because the sidecar uses the Get RPC to confirm the state of previously unhealthy volumes that are absent from a list response.

The sidecar exits at startup if the driver advertises neither capability, or advertises `LIST_VOLUME_HEALTH` without `GET_VOLUME_HEALTH`.

Health conditions reported by the driver are written to `pvc.status.healthStatus.healthConditions` as `(status, reason, message)` entries, where `status` is one of `Inaccessible`, `DataLoss`, or `Degraded`. The driver's report is authoritative. Conditions the driver no longer reports are removed, an explicitly healthy report clears the field, and nothing is written when nothing has changed.

Health conditions of a type this sidecar does not recognize are never written to `pvc.status.healthStatus`, as the CSI spec requires. They are surfaced instead as a `Warning` event with reason `UnknownVolumeHealthCondition` on the PVC and counted in the `csi_volume_health_unknown_condition_total` metric.

The `CSIVolumeHealth` feature gate must be enabled on kube-apiserver for `pvc.status.healthStatus` to be persisted. The sidecar itself has no feature gate. Deploying it is how a cluster opts in on the controller side. When the gate is disabled, the API server silently drops the field. The sidecar detects this, logs it, counts it in the `csi_volume_health_status_writes_dropped_total` metric, and keeps retrying so reporting resumes as soon as the gate is enabled.

## Compatibility

This information reflects the head of this branch.

| Compatible with CSI Version                                                                          | Container Image             |
| ----------------------------------------------------------------------------------------------------- | ----------------------------|
| [CSI Spec v1.13.0](https://github.com/container-storage-interface/spec/releases/tag/v1.13.0-rc1) | registry.k8s.io/sig-storage/csi-external-health-monitor-controller |

## Usage

External Health Monitor Controller needs to be deployed with CSI driver.

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

Check `pvc.status.healthStatus` for controller-reported abnormal volume health when the volume you are using is abnormal.

## Command line options

### Important optional arguments that are highly recommended to be used

- `leader-election`: Enables leader election. This is useful when there are multiple replicas of the same external-health-monitor-controller running for one CSI driver. Only one of them may be active (=leader). A new leader will be re-elected when the current leader dies or becomes unresponsive for ~15 seconds.

- `leader-election-namespace <namespace>`: The namespace where the leader election resource exists. Defaults to the pod namespace if not set.

- `leader-election-lease-duration <duration>`: Duration, in seconds, that non-leader candidates will wait to force acquire leadership. Defaults to 15 seconds.

- `leader-election-renew-deadline <duration>`: Duration, in seconds, that the acting leader will retry refreshing leadership before giving up. Defaults to 10 seconds.

- `leader-election-retry-period <duration>`: Duration, in seconds, the LeaderElector clients should wait between tries of actions. Defaults to 5 seconds.

- `http-endpoint`: The TCP network address where the HTTP server for diagnostics, including metrics and leader election health check, will listen (example: `:8080` which corresponds to port 8080 on local host). The default is empty string, which means the server is disabled.

- `metrics-path`: The HTTP path where prometheus metrics will be exposed. Default is /metrics.

- `worker-threads`: Maximum number of concurrent `ControllerGetVolumeHealth` calls. Each volume is checked once per `monitor-interval` regardless of this value. The default value is 10.

### Other recognized arguments

- `kubeconfig <path>`: Path to Kubernetes client configuration that the external-health-monitor-controller uses to connect to the Kubernetes API server. When omitted, default token provided by Kubernetes will be used. This option is useful only when the external-health-monitor-controller does not run as a Kubernetes pod, e.g. for debugging.

- `resync <duration>`: Internal resync interval when the monitor controller re-evaluates all existing resource objects that it was watching and tries to fulfill them. It does not affect re-tries of failed calls! It should be used only when there is a bug in Kubernetes watch logic. The default is ten minutes.

- `csiAddress <path-to-csi>`: This is the path to the CSI Driver socket inside the pod that the external-health-monitor-controller container will use to issue CSI operations (/run/csi/socket is used by default).

- `version`: Prints the current version of external-health-monitor-controller.

- `timeout <duration>`: Timeout of each call to the CSI driver and of each PVC status patch. It should be set to a value that accommodates the majority of `ControllerListVolumeHealth` pages and `ControllerGetVolumeHealth` calls. 15 seconds is used by default.

- `list-volumes-interval <duration>`: Interval of monitoring volume health condition by invoking `ControllerListVolumeHealth`. You can adjust it to change the frequency of the evaluation process. Five minutes by default if not set.

- `monitor-interval <duration>`: Interval of monitoring volume health condition when CSI Driver supports `ControllerGetVolumeHealth`, but not `ControllerListVolumeHealth`. You can adjust it to change the frequency of the evaluation process. One minute by default if not set.

- `volume-list-add-interval <duration>`: Interval of listing volumes and adding them to the queue when CSI driver supports `ControllerGetVolumeHealth`, but not `ControllerListVolumeHealth`.

- `metrics-address`: (deprecated) The TCP network address where the Prometheus metrics endpoint will run (example: :8080, which corresponds to port 8080 on local host). The default is the empty string, which means the metrics and leader election check endpoint is disabled.

- `--automaxprocs`: Automatically set the `GOMAXPROCS` environment variable to match the configured Linux container CPU quota. Defaults to false.

* [Arguments set by the `k8s.io/component-base/logs` package for klog](https://github.com/kubernetes/component-base/blob/v0.28.0-rc.0/logs/api/v1/options.go#L337-L355) are supported, such as `--v <log level>` and `--logging-format <log format>`.

## Metrics

The sidecar exposes the following metrics when `--http-endpoint` is set:

- `csi_volume_health_probe_total`: cumulative count of health probes, broken down by CSI method and result.
- `csi_volume_health_probe_duration_seconds`: health probe RPC latency, broken down by CSI method.
- `csi_controller_volume_health_status`: one series per condition status currently reported on an unhealthy volume's PVC, with value 1.
- `csi_volume_health_status_writes_dropped_total`: cumulative count of status writes dropped by the API server, which indicates the `CSIVolumeHealth` feature gate is disabled.
- `csi_volume_health_unknown_condition_total`: cumulative count of health entries observed with a CSI error type unknown to this sidecar, broken down by the raw CSI status value.

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack](https://kubernetes.slack.com/messages/sig-storage)
- [Mailing List](https://groups.google.com/forum/#!forum/kubernetes-sig-storage)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
