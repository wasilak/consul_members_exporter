package main

import (
	"flag"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "consul_members"

var (
	listenAddress = flag.String("web.listen-address", ":9142",
		"Address to listen on for telemetry")
	metricsPath = flag.String("web.telemetry-path", "/metrics",
		"Path under which to expose metrics")

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

type Exporter struct {
	Agent *api.Agent
}

func NewExporter(agent *api.Agent) *Exporter {
	return &Exporter{Agent: agent}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- membersGauge
}

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

	http.Handle(*metricsPath, promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<html>
			<head><title>Consul Members Exporter</title></head>
			<body>
			<h1>Consul Members Exporter</h1>
			<p><a href='` + *metricsPath + `'>Metrics</a></p>
			</body>
			</html>`))
	})

	http.ListenAndServe(*listenAddress, nil)
}
