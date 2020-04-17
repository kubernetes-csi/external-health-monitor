FROM gcr.io/distroless/static:latest
LABEL maintainers="Kubernetes Authors"
LABEL description="CSI External Health Monitor Controller"

COPY ./bin/csi-external-health-monitor-controller csi-external-health-monitor-controller
ENTRYPOINT ["/csi-external-health-monitor-controller"]
