package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "consul_members"

var (
	listenAddress = flag.String("listen-address", ":9142", "Address to listen on for telemetry")
	metricsPath = flag.String("telemetry-path", "/metrics", "Path under which to expose metrics")

	membersGauge = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "details"),
		"Consul member details gauge with constant value of 1 and information in labels",
		[]string{"name", "version", "addr", "status", "statusText", "server"}, nil,
	)

	memberStatuses = map[int]string{
		0: "None",
		1: "Alive",
		2: "Leaving",
		3: "Left",
		4: "Failed",
	}
)

// Middleware type
type Middleware func(http.HandlerFunc) http.HandlerFunc

// MiddlewareHandler type
type MiddlewareHandler func(http.Handler) http.Handler

// Exporter struct
type Exporter struct {
	Agent *api.Agent
}

// NewExporter func
func NewExporter(agent *api.Agent) *Exporter {
	return &Exporter{Agent: agent}
}

// Describe func
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- membersGauge
}

// Collect func
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	members, _ := e.Agent.Members(false)

	for _, member := range members {
		buildDetails := strings.Split(member.Tags["build"], ":")

		ch <- prometheus.MustNewConstMetric(
			membersGauge,
			prometheus.GaugeValue,
			1,
			member.Name,
			buildDetails[0],
			member.Addr,
			strconv.Itoa(member.Status),
			memberStatuses[member.Status],
			strconv.FormatBool(member.Tags["role"] == "consul"),
		)
	}
}

func logWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r) // call original
		start := time.Now()
		defer func() {
			log.Printf("%s %s %s %s %s %s\n", r.Host, r.Method, r.RemoteAddr, r.RequestURI, time.Since(start), r.UserAgent())
		}()
	})
}

func rootHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Consul Members Exporter</title></head>
			<body>
			<h1>Consul Members Exporter</h1>
			<p><a href='` + *metricsPath + `'>Metrics</a></p>
			</body>
			</html>`))
	})
}


func main() {
	flag.Parse()

	// Get a new client
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}

	agent := client.Agent()

	exporter := NewExporter(agent)
	prometheus.MustRegister(exporter)

	http.Handle("/", logWrapper(rootHandler()))
	http.Handle(*metricsPath, logWrapper(promhttp.Handler()))

	log.Printf("Consul Members Exporter started :: listening on %s\n", *listenAddress)

	http.ListenAndServe(*listenAddress, nil)
}
