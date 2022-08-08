#! /bin/bash

. release-tools/prow.sh

# There are multiple reasons why installing the external-health-monitor-controller
# cannot use the "normal" code:
# - the RBAC file is not where https://github.com/kubernetes-csi/csi-release-tools/blob/7fe51491d8f56d0d459bf9abb9692bc9f96d2b75/prow.sh#L1243-L1258
#   expects it
# - the variable for overriding the RBAC source in https://github.com/kubernetes-csi/csi-driver-host-path/blob/9be5dd74a7fc2436c4334820156056b74821998e/deploy/util/deploy-hostpath.sh#L157-L158
#   does not match the command name (csi-external-health-monitor-controller -> CSI_EXTERNAL_HEALTH_MONITOR_CONTROLLER_RBAC !=
#   CSI_EXTERNALHEALTH_MONITOR_RBAC_YAML)
#
# The following hack works around that mismatch. It was added because it could be rolled
# out without updating csi-release-tools and csi-driver-host-path.
CSI_PROW_DRIVER_INSTALL_ORIGINAL="${CSI_PROW_DRIVER_INSTALL}"
CSI_PROW_DRIVER_INSTALL=install_csi_driver_health_monitor
install_csi_driver_health_monitor () (
    set -x
    images="$1"

    if ${CSI_PROW_BUILD_JOB}; then
       # Add the RBAC env variable. The prow.sh code did not find the file.
        images+=" CSI_EXTERNALHEALTH_MONITOR_RBAC=$(pwd)/deploy/kubernetes/external-health-monitor-controller/rbac.yaml"
    fi

    "${CSI_PROW_DRIVER_INSTALL_ORIGINAL}" "$images"
    set +x
)

main
