package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	version = "0.1"
	build   = "."

	config *string = flag.StringP("config", "c", "", "explicit config file name")
)

func init() {
	fmt.Printf("grafana-proxy v%s build %s\n\n", version, build)
	flag.Parse()

	viper.SetDefault("grafana.url", "http://localhost:3000")
	viper.SetDefault("grafana.cacerts", "")
	viper.SetDefault("grafana.datasource", 1)
	viper.SetDefault("grafana.key", "")
	viper.SetDefault("server.address", "localhost:8080")
	viper.SetDefault("log.level", "info")

	viper.SetConfigName("grafana-proxy")
	viper.AddConfigPath(".")
	if *config != "" {
		viper.SetConfigFile(*config)
	}
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config file: %s \n", err)
	}

	viper.SetEnvPrefix("GP")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()
}

func main() {
	log.WithFields(log.Fields{
		"file": viper.ConfigFileUsed(),
	}).Info("loaded config")

	switch viper.GetString("log.level") {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "warning":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	u, err := url.Parse(singleJoiningSlash(viper.GetString("grafana.url"), fmt.Sprintf("/api/datasources/proxy/%d/", viper.GetInt("grafana.datasource"))))
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("proxying to %s", u)

	certs, err := LoadCerts(viper.GetString("grafana.cacerts"))
	if err != nil {
		log.Fatal(err)
	}
	t := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: certs,
		},
	}
	proxy := NewGrafanaProxy(u, viper.GetString("grafana.key"), t)
	log.Error(http.ListenAndServe(viper.GetString("server.address"), loggingHandler(proxy)))
}

// loggingHandler wraps an http.Handler to log each request
func loggingHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.EscapedPath()
		h.ServeHTTP(w, r)
		// get real origin for forwarded requests
		var remoteAddr string
		if remote := r.Header.Get("X-Real-IP"); remote != "" {
			remoteAddr = remote
		} else if remote := r.Header.Get("X-Forwarded-For"); remote != "" {
			remoteAddr = remote
		} else {
			remoteAddr = r.RemoteAddr
		}
		log.WithFields(log.Fields{
			"origin": remoteAddr,
			"length": r.ContentLength,
			"agent":  r.UserAgent(),
			"path":   path,
			"time":   time.Since(start).Nanoseconds(),
		}).Info("handled request")
	})
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

// NewGrafanaProxy is based on httputil.NewSingleHostReverseProxy but adds the auth headers.
func NewGrafanaProxy(target *url.URL, key string, transport *http.Transport) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
	}
	return &httputil.ReverseProxy{
		Director:  director,
		Transport: transport,
	}
}
