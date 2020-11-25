# External Health Monitor

The External Health Monitor is part of Kubernetes implementation of [Container Storage Interface (CSI)](https://github.com/container-storage-interface/spec). It was introduced as an Alpha feature in Kubernetes v1.19.

## Overview

The [External Health Monitor](https://github.com/kubernetes/enhancements/tree/master/keps/sig-storage/1432-volume-health-monitor) is implemented as two components: `External Health Monitor Controller` and `External Health Monitor Agent`.

- External Health Monitor Controller:
  - The external health monitor controller will be deployed as a sidecar together with the CSI controller driver, similar to how the external-provisioner sidecar is deployed.
  - Trigger controller RPC to check the health condition of the CSI volumes.
  - The external controller sidecar will also watch for node failure events. This component can be enabled via a flag.

- External Health Monitor Agent:
  - The external health monitor agent will be deployed as a sidecar together with the CSI node driver on every Kubernetes worker node.
  - Trigger node RPC to check volume's mounting conditions.

The External Health Monitor needs to invoke the following CSI interfaces.

- External Health Monitor Controller:
  - ListVolumes (If both `ListVolumes` and `ControllerGetVolume` are supported, `ListVolumes` will be used)
  - ControllerGetVolume
- External Health Monitor Agent:
  - NodeGetVolumeStats

## Compatibility

This information reflects the head of this branch.

| Compatible with CSI Version                                                                | Container Image             | [Min K8s Version](https://kubernetes-csi.github.io/docs/kubernetes-compatibility.html#minimum-version) | Recommend K8s version |
| ------------------------------------------------------------------------------------------ | ----------------------------| --------------- | -------------------- |
| [CSI Spec v1.3.0](https://github.com/container-storage-interface/spec/releases/tag/v1.3.0) | k8s.gcr.io/sig-storage.csi-external-health-monitor-controller | 1.19         | 1.19              |
| [CSI Spec v1.3.0](https://github.com/container-storage-interface/spec/releases/tag/v1.3.0) | k8s.gcr.io/sig-storage/csi-external-health-monitor-agent  | 1.19     | 1.19              |

## Driver Support

Currently, the CSI volume health monitoring interfaces are only implemented in the Mock Driver.

## Usage

External Health Monitor needs to be deployed with CSI driver.

### Build && Push Image

You can run the command below in the root directory of the project.

```bash
make container GOFLAGS_VENDOR=$( [ -d vendor ] && echo '-mod=vendor' )
```

And then, you can tag and push it to your own image repository.

```bash
docker tag csi-external-health-monitor-controller:latest <custom-image-repo-addr>/csi-external-health-monitor-controller:<custom-image-tag>

docker tag csi-external-health-monitor-agent:latest <custom-image-repo-addr>/csi-external-health-monitor-agent:<custom-image-tag>
```

### External Health Monitor Controller

```bash
cd external-health-monitor
kubectl create -f deploy/kubernetes/external-health-monitor-controller
```

### External Health Monitor Agent

```bash
kubectl create -f deploy/kubernetes/external-health-monitor-agent
```

You can run `kubectl get pods` command to confirm if they are deployed on your cluster successfully.

Check logs of external health monitor controller and agent as follows:

-  `kubectl logs <leader-of-external-health-monitor-controller-container-name> -c csi-external-health-monitor-controller`
-  `kubectl logs <external-health-monitor-agent-container-name> -c csi-external-health-monitor-agent`

Check if there are events on PVCs or Pods that report abnormal volume condition when the volume you are using is abnormal.

## csi-external-health-monitor-controller-sidecar-command-line-options

### Important optional arguments that are highly recommended to be used

- `leader-election`: Enables leader election. This is useful when there are multiple replicas of the same external-health-monitor-controller running for one CSI driver. Only one of them may be active (=leader). A new leader will be re-elected when the current leader dies or becomes unresponsive for ~15 seconds.

- `leader-election-namespace <namespace>`: The namespace where the leader election resource exists. Defaults to the pod namespace if not set.

- `metrics-address`: The TCP network address where the Prometheus metrics endpoint and leader election health check will run (example: :8080, which corresponds to port 8080 on local host). The default is the empty string, which means the metrics and leader election check endpoint is disabled.

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


## csi-external-health-monitor-agent-sidecar-command-line-options

### Important optional arguments that are highly recommended to be used

- `metrics-address`: The TCP network address where the prometheus metrics endpoint will run (example: :8080, which corresponds to port 8080 on localhost). The default is the empty string, which means the metrics endpoint is disabled.

- `metrics-path`: The HTTP path where prometheus metrics will be exposed. Default is /metrics.

- `worker-threads`: Number of worker threads for running volume checker by invoking RPC interface `NodeGetVolumeStats`. Default value is 10.

### Other recognized arguments

- `kubeconfig <path>`: Path to Kubernetes client configuration that the external-health-monitor-agent uses to connect to Kubernetes API server. When omitted, the default token provided by Kubernetes will be used. This option is useful only when the external-health-monitor-agent does not run as a Kubernetes pod, e.g. for debugging.

- `resync <duration>`: Internal resync interval when the monitor agent re-evaluates all existing resource objects that it was watching and tries to fulfill them. It does not affect re-tries of failed calls! It should be used only when there is a bug in Kubernetes watch logic. The default is ten mintiues.

- `monitor-interval <duration>`: Interval of monitoring volume health condition by invoking RPC interface `NodeGetVolumeStats`. You can adjust it to change the frequency of the evaluation process. One minute by default if not set.

- `csiAddress <path-to-csi>`: This is the path to the CSI Driver socket inside the pod that the external-health-monitor-agent container will use to issue CSI operations (/run/csi/socket is used by default).

- `version`: Prints the current version of external-health-monitor-agent.

- `timeout <duration>`: Timeout of all calls to CSI Driver. It should be set to value that accommodates the majority of `NodeGetVolumeStats` calls. 15 seconds is used by default.

- `kubelet-root-path`: Path to kubelet. It is used to generate the volume path. `/var/lib/kubelet` by default if not set.

### HTTP endpoint

Both sidecars optionally exposes an HTTP endpoint at
address:port specified by `--metrics-address` argument. When set,
these two paths may be exposed:

* Metrics path, as set by `--metrics-path` argument (default is
  `/metrics`) - both sidecars.
* Leader election health check at `/healthz/leader-election` - only
  in the External Health Monitor Controller.
  It is recommended to run a liveness probe against this endpoint when
  leader election is used to kill a external-health-monitor-controller
  leader that fails to connect to the API server to renew its leadership. See
  https://github.com/kubernetes-csi/csi-lib-utils/issues/66 for
  details.

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack](https://kubernetes.slack.com/messages/sig-storage)
- [Mailing List](https://groups.google.com/forum/#!forum/kubernetes-sig-storage)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
