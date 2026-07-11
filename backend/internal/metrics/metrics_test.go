package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestSetTunnel(t *testing.T) {
	m := New()
	cases := map[string]float64{"up": 1, "negotiating": 0.5, "down": 0, "unknown": 0}
	for status, want := range cases {
		m.SetTunnel("t", "gw", status)
		got := testutil.ToFloat64(m.TunnelStatus.WithLabelValues("t", "gw"))
		if got != want {
			t.Fatalf("status %q → %v (attendu %v)", status, got, want)
		}
	}
}

func TestVICIErrorsCounter(t *testing.T) {
	m := New()
	m.VICIErrors.Inc()
	m.VICIErrors.Inc()
	if got := testutil.ToFloat64(m.VICIErrors); got != 2 {
		t.Fatalf("compteur = %v (attendu 2)", got)
	}
}
