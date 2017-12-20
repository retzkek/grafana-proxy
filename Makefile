BUILD=`date -u +%Y-%m-%dT%H:%M:%SZ`
build:
	go build -ldflags="-X main.build=${BUILD}"

docker:
	docker build -t landscape/grafana-proxy .
