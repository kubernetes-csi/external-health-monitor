FROM gcr.io/distroless/static:latest
LABEL maintainers="Kubernetes Authors"
LABEL description="CSI External Health Monitor Agent"

COPY ./bin/csi-external-health-monitor-agent csi-external-health-monitor-agent
ENTRYPOINT ["/csi-external-health-monitor-agent"]
