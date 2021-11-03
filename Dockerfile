FROM golang:alpine as builder
WORKDIR /build
COPY . /build/grafana-proxy
RUN apk add --update --no-cache --virtual build-dependencies \
 && cd grafana-proxy  \
 && go build -tags netgo -ldflags="-X main.BUILD=$(date -u +%Y-%m-%dT%H:%M:%SZ)"

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root
COPY --from=builder /build/grafana-proxy/grafana-proxy ./
COPY --from=builder /build/grafana-proxy/grafana-proxy.yml ./
CMD ["./grafana-proxy"]
