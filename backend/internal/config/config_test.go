package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	// s'assure qu'aucune variable n'est héritée
	for _, k := range []string{"HTTP_ADDR", "JWT_TTL", "VICI_ENDPOINTS", "POLL_INTERVAL", "CORS_ORIGINS"} {
		t.Setenv(k, "")
	}
	c := Load()
	if c.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr défaut = %q", c.HTTPAddr)
	}
	if c.JWTTTL != time.Hour {
		t.Fatalf("JWTTTL défaut = %v", c.JWTTTL)
	}
	if len(c.CORSOrigins) != 1 || c.CORSOrigins[0] != "*" {
		t.Fatalf("CORSOrigins défaut = %v", c.CORSOrigins)
	}
	if len(c.ViciEndpoints) != 0 {
		t.Fatalf("ViciEndpoints défaut non vide: %v", c.ViciEndpoints)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("JWT_TTL", "30m")
	t.Setenv("POLL_INTERVAL", "5") // secondes
	t.Setenv("VICI_ENDPOINTS", "gw-a=unix:/a, gw-b=tcp:h:1")
	t.Setenv("CORS_ORIGINS", "https://x.io, https://y.io")

	c := Load()
	if c.HTTPAddr != ":9090" {
		t.Fatalf("HTTPAddr = %q", c.HTTPAddr)
	}
	if c.JWTTTL != 30*time.Minute {
		t.Fatalf("JWTTTL = %v", c.JWTTTL)
	}
	if c.PollInterval != 5*time.Second {
		t.Fatalf("PollInterval = %v (attendu 5s)", c.PollInterval)
	}
	if len(c.ViciEndpoints) != 2 || c.ViciEndpoints[0] != "gw-a=unix:/a" || c.ViciEndpoints[1] != "gw-b=tcp:h:1" {
		t.Fatalf("ViciEndpoints = %v", c.ViciEndpoints)
	}
	if len(c.CORSOrigins) != 2 {
		t.Fatalf("CORSOrigins = %v", c.CORSOrigins)
	}
}
