FROM gcr.io/distroless/static:latest-amd64
LABEL maintainers="Kubernetes Authors"
LABEL description="CSI External Health Monitor Controller"
ARG binary=./bin/csi-external-health-monitor-controller

COPY ${binary} csi-external-health-monitor-controller
ENTRYPOINT ["/csi-external-health-monitor-controller"]
