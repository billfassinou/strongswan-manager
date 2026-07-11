// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package poller

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"strongswan-manager/internal/domain"
	"strongswan-manager/internal/metrics"
	"strongswan-manager/internal/vici"
	"strongswan-manager/internal/ws"
)

type fakeTunnels struct {
	list    []domain.Tunnel
	updated map[string]string
}

func (f *fakeTunnels) List(context.Context) ([]domain.Tunnel, error) { return f.list, nil }
func (f *fakeTunnels) UpdateStatus(_ context.Context, id, status string) error {
	f.updated[id] = status
	return nil
}

type fakeGateways struct{ status, version string }

func (f *fakeGateways) UpdateStatus(_ context.Context, _, status, version string) error {
	f.status, f.version = status, version
	return nil
}

func TestPollerTickUpdatesStatusAndGateway(t *testing.T) {
	ctx := context.Background()

	// mock VICI : une conn "t1" établie → list-sas renvoie t1=up, version connue
	mock := vici.NewMock()
	_ = mock.LoadConn(ctx, "t1", map[string]any{"version": "2"})
	_ = mock.Initiate(ctx, "t1-net")
	reg := vici.NewRegistry()
	reg.Set("g1", mock)

	ft := &fakeTunnels{
		list:    []domain.Tunnel{{ID: "id1", Name: "t1", GatewayID: "g1", Status: domain.StatusDown}},
		updated: map[string]string{},
	}
	fg := &fakeGateways{}
	m := metrics.New()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	p := New(ft, fg, reg, ws.NewHub(), m, time.Second, log)
	p.tick(ctx)

	if ft.updated["id1"] != domain.StatusUp {
		t.Fatalf("statut tunnel = %q (attendu up)", ft.updated["id1"])
	}
	if fg.status != "up" || fg.version != "6.0.1" {
		t.Fatalf("statut passerelle = %q / %q (attendu up / 6.0.1)", fg.status, fg.version)
	}
}

func TestPollerNoChangeNoUpdate(t *testing.T) {
	ctx := context.Background()
	mock := vici.NewMock()
	_ = mock.LoadConn(ctx, "t1", map[string]any{})
	_ = mock.Initiate(ctx, "t1-net") // t1 = up
	reg := vici.NewRegistry()
	reg.Set("g1", mock)

	// le tunnel est déjà "up" → aucun UpdateStatus attendu
	ft := &fakeTunnels{
		list:    []domain.Tunnel{{ID: "id1", Name: "t1", GatewayID: "g1", Status: domain.StatusUp}},
		updated: map[string]string{},
	}
	p := New(ft, &fakeGateways{}, reg, ws.NewHub(), metrics.New(), time.Second, slog.New(slog.NewTextHandler(io.Discard, nil)))
	p.tick(ctx)

	if _, ok := ft.updated["id1"]; ok {
		t.Fatal("aucun UpdateStatus ne devait être émis pour un statut inchangé")
	}
}
