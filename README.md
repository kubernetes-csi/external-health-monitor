# External Health Monitor

The External Health Monitor is part of Kubernetes implementation of [Container Storage Interface (CSI)](https://github.com/container-storage-interface/spec). It was introduced as an Alpha feature in Kubernetes v1.19.

## Overview

The [External Health Monitor](https://github.com/kubernetes/enhancements/tree/master/keps/sig-storage/1432-volume-health-monitor) is impelmeted as two components: `External Health Monitor Controller` and `External Health Monitor Agent`.

- External Health Monitor Controller:
  - The external health monitor controller will be deployed as a sidecar together with the CSI controller driver, similar to how the external-provisioner sidecar is deployed.
  - Trigger controller RPC to check the health condition of the CSI volumes.
  - The external controller sidecar will also watch for node failure events. This component can be enabled via a flag.

- External Health Monitor Agent:
  - The external health monitor agent will be deployed as a sidecar together with the CSI node driver on every Kubernetes worker node.
  - Trigger node RPC to check volume's mounting conditions.

The External Health Monitor needs to invoke the following CSI interfaces

- External Health Monitor Controller:
  - ListVolumes
  - ControllerGetVolume
- External Health Monitor Agent:
  - NodeGetVolumeStats

## Driver Support

Currently the CSI volume health monitoring interfaces are only implemented in the Mock Driver.

## Usage

External Health Monitor needs to be deployed with CSI driver.

### Build && Push Image

You can run the command as below in the root directory of the project.

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

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack](https://kubernetes.slack.com/messages/sig-storage)
- [Mailing List](https://groups.google.com/forum/#!forum/kubernetes-sig-storage)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
