package vici

import (
	"context"
	"testing"

	"strongswan-manager/internal/domain"
)

func TestMockLoadInitiateUnload(t *testing.T) {
	ctx := context.Background()
	m := NewMock()

	if v, _ := m.Version(ctx); v == "" {
		t.Fatal("version vide")
	}

	// charge une connexion
	if err := m.LoadConn(ctx, "paris-dakar", map[string]any{"version": "2"}); err != nil {
		t.Fatal(err)
	}
	if !m.HasConn("paris-dakar") {
		t.Fatal("conn non chargée")
	}

	// down avant initiate
	sas, _ := m.ListSAs(ctx)
	if len(sas) != 1 || sas[0].Status != domain.StatusDown {
		t.Fatalf("état initial inattendu: %+v", sas)
	}

	// initiate → up (le child porte le suffixe -net)
	if err := m.Initiate(ctx, "paris-dakar-net"); err != nil {
		t.Fatal(err)
	}
	sas, _ = m.ListSAs(ctx)
	if sas[0].Status != domain.StatusUp {
		t.Fatalf("après initiate: %s (attendu up)", sas[0].Status)
	}

	// terminate → down
	_ = m.Terminate(ctx, "paris-dakar-net")
	sas, _ = m.ListSAs(ctx)
	if sas[0].Status != domain.StatusDown {
		t.Fatalf("après terminate: %s", sas[0].Status)
	}

	// unload
	_ = m.UnloadConn(ctx, "paris-dakar")
	if m.HasConn("paris-dakar") {
		t.Fatal("conn toujours présente après unload")
	}
}

func TestConnOf(t *testing.T) {
	if connOf("paris-dakar-net") != "paris-dakar" {
		t.Fatal("connOf incorrect")
	}
}
