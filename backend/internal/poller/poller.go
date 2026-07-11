// Package poller sonde périodiquement l'état des SA via VICI et propage les
// changements vers la base, les métriques et les clients WebSocket.
package poller

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"strongswan-manager/internal/domain"
	"strongswan-manager/internal/metrics"
	"strongswan-manager/internal/vici"
	"strongswan-manager/internal/ws"
)

// tunnelStore et gatewayStore sont les portions du store dont le poller dépend
// (interfaces → testable sans PostgreSQL).
type tunnelStore interface {
	List(ctx context.Context) ([]domain.Tunnel, error)
	UpdateStatus(ctx context.Context, id, status string) error
}

type gatewayStore interface {
	UpdateStatus(ctx context.Context, id, status, version string) error
}

// Poller boucle sur les passerelles enregistrées.
type Poller struct {
	tunnels  tunnelStore
	gateways gatewayStore
	registry *vici.Registry
	hub      *ws.Hub
	metrics  *metrics.Metrics
	interval time.Duration
	log      *slog.Logger
}

// New construit un poller.
func New(tunnels tunnelStore, gateways gatewayStore, reg *vici.Registry, hub *ws.Hub, m *metrics.Metrics, interval time.Duration, log *slog.Logger) *Poller {
	return &Poller{tunnels: tunnels, gateways: gateways, registry: reg, hub: hub, metrics: m, interval: interval, log: log}
}

// Run bloque jusqu'à l'annulation du contexte, en sondant à intervalle régulier.
func (p *Poller) Run(ctx context.Context) {
	t := time.NewTicker(p.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			p.tick(ctx)
		}
	}
}

func (p *Poller) tick(ctx context.Context) {
	// état live agrégé par nom de connexion + rafraîchissement du statut passerelle
	live := map[string]string{}
	p.registry.Each(func(gatewayID string, a vici.Adapter) {
		if ver, err := a.Version(ctx); err == nil {
			_ = p.gateways.UpdateStatus(ctx, gatewayID, "up", ver)
		} else {
			_ = p.gateways.UpdateStatus(ctx, gatewayID, "unknown", "")
		}
		sas, err := a.ListSAs(ctx)
		if err != nil {
			p.metrics.VICIErrors.Inc()
			return
		}
		for _, sa := range sas {
			live[sa.Name] = sa.Status
		}
	})

	tunnels, err := p.tunnels.List(ctx)
	if err != nil {
		p.log.Warn("poller: lecture tunnels", "err", err)
		return
	}
	for i := range tunnels {
		t := &tunnels[i]
		newStatus := domain.StatusDown
		if st, ok := live[t.Name]; ok {
			newStatus = st
		}
		p.metrics.SetTunnel(t.Name, t.GatewayID, newStatus)
		if newStatus != t.Status {
			_ = p.tunnels.UpdateStatus(ctx, t.ID, newStatus)
			p.emit(t.ID, t.Name, newStatus)
		}
	}
}

func (p *Poller) emit(id, name, status string) {
	msg, _ := json.Marshal(map[string]any{
		"type":   "tunnel_status",
		"id":     id,
		"name":   name,
		"status": status,
	})
	p.hub.Broadcast(msg)
}
