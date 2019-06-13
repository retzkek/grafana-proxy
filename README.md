`grafana-proxy` is a simple HTTP proxy that provides transparent local
access to a remote Grafana datasource, which you have access to via an
[API token](https://grafana.com/docs/http_api/auth/).

## Installation

A container image is available on Docker hub at `retzkek/grafana-proxy`.

## Building

Run `make` to build a stand-alone binary (requires Go > v1.8).

Run `make docker` to build the binary in a container and then produce
a small Aline-Linux-based deployable image `landscape/grafana-proxy`.

## Configuration

Configuration can be provided in YML, JSON, TOML, or property file
format, in a file in the current working directoy named
`grafana-proxy` with the appropriate extension. See example below.

All configuarion values can also be specified via environment
variables, where the variable name is the underscore-separated path
prefixed with `GP_`, e.g. `GP_GRAFANA_URL`.

    grafana:
      # base Grafana URL
      url: http://localhost:3000
      # file or directory containing PEM-encoded CA certs.
      # if empty or omitted will use system defaults.
      cacerts: ""
      # Grafana API key
      key: ""
      # datasource number in Grafana
      datasource: 1

    server:
      # interface and port to serve proxy on
      address: localhost:8080

    log:
      # debug, info, warning, error
      level: info
