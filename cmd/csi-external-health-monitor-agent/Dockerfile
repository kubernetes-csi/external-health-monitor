FROM gcr.io/distroless/static:latest-amd64
LABEL maintainers="Kubernetes Authors"
LABEL description="CSI External Health Monitor Agent"
ARG binary=./bin/csi-external-health-monitor-agent

COPY ${binary} csi-external-health-monitor-agent
ENTRYPOINT ["/csi-external-health-monitor-agent"]
