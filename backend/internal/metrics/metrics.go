// Package metrics expose les métriques Prometheus (source de vérité des séries, V5).
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics regroupe les collecteurs applicatifs.
type Metrics struct {
	reg          *prometheus.Registry
	TunnelStatus *prometheus.GaugeVec
	VICIErrors   prometheus.Counter
}

// New crée un registre dédié (pas le registre global) avec les collecteurs go/process.
func New() *Metrics {
	reg := prometheus.NewRegistry()
	m := &Metrics{
		reg: reg,
		TunnelStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "strongswan_tunnel_status",
			Help: "État du tunnel (1=up, 0.5=negotiating, 0=down/unknown).",
		}, []string{"tunnel", "gateway"}),
		VICIErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "strongswan_vici_errors_total",
			Help: "Nombre d'erreurs de communication VICI.",
		}),
	}
	reg.MustRegister(m.TunnelStatus, m.VICIErrors)
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	return m
}

// SetTunnel met à jour la gauge d'état d'un tunnel.
func (m *Metrics) SetTunnel(tunnel, gateway, status string) {
	var v float64
	switch status {
	case "up":
		v = 1
	case "negotiating":
		v = 0.5
	default:
		v = 0
	}
	m.TunnelStatus.WithLabelValues(tunnel, gateway).Set(v)
}

// Handler renvoie le handler HTTP /metrics.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.reg, promhttp.HandlerOpts{})
}
