// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

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
	if c.HTTPAddr != ":7926" {
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

// Le HTTPS doit être actif SANS configuration : c'est tout l'intérêt du défaut.
func TestLoadDefaults_TLS(t *testing.T) {
	for _, k := range []string{"TLS_ENABLED", "HTTP_REDIRECT_ADDR", "TLS_CERT", "TLS_KEY", "TLS_SANS", "ACME_DOMAIN"} {
		t.Setenv(k, "")
	}
	c := Load()
	if !c.TLSEnabled {
		t.Fatal("TLSEnabled doit valoir true par défaut — l'application démarre en HTTPS sans configuration")
	}
	if c.RedirectAddr != ":7927" {
		t.Fatalf("RedirectAddr défaut = %q, attendu :7927 (écouteur clair pour le CDP)", c.RedirectAddr)
	}
	if c.TLSCert != "" || c.TLSKey != "" || c.ACMEDomain != "" {
		t.Fatal("aucun certificat fourni ni ACME par défaut : le certificat est auto-généré")
	}
	if c.ACMECache != "./acme" {
		t.Fatalf("ACMECache défaut = %q", c.ACMECache)
	}
}

// TLS_ENABLED=false est la porte de sortie pour un déploiement derrière un reverse proxy
// qui termine déjà le TLS : elle doit rester fonctionnelle.
func TestLoadTLSDesactivable(t *testing.T) {
	t.Setenv("TLS_ENABLED", "false")
	if Load().TLSEnabled {
		t.Fatal("TLS_ENABLED=false doit désactiver le TLS")
	}
	t.Setenv("TLS_ENABLED", "0")
	if Load().TLSEnabled {
		t.Fatal("TLS_ENABLED=0 doit désactiver le TLS")
	}
	// Une valeur ininterprétable ne doit pas désactiver le TLS par accident.
	t.Setenv("TLS_ENABLED", "n'importe quoi")
	if !Load().TLSEnabled {
		t.Fatal("une valeur invalide doit retomber sur le défaut (true), jamais dégrader la sécurité")
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

func TestValidateRejectsDevSecrets(t *testing.T) {
	base := func() Config {
		return Config{JWTSecret: "unsecretbienlong", SecretsKey: "uneclebienlongue"}
	}

	if err := base().Validate(); err != nil {
		t.Fatalf("configuration saine refusée : %v", err)
	}

	c := base()
	c.JWTSecret = DefaultJWTSecret
	if err := c.Validate(); err == nil {
		t.Fatal("JWT_SECRET par défaut accepté : le serveur démarrerait avec un secret public")
	}

	c = base()
	c.SecretsKey = DefaultSecretsKey
	if err := c.Validate(); err == nil {
		t.Fatal("SECRETS_KEY par défaut acceptée")
	}

	// Le lab doit continuer à démarrer.
	c = Config{JWTSecret: DefaultJWTSecret, SecretsKey: DefaultSecretsKey, AllowInsecureDefaults: true}
	if err := c.Validate(); err != nil {
		t.Fatalf("ALLOW_INSECURE_DEFAULTS n'a pas levé le blocage : %v", err)
	}
}

func TestLoadAllowInsecureDefaults(t *testing.T) {
	t.Setenv("ALLOW_INSECURE_DEFAULTS", "true")
	if !Load().AllowInsecureDefaults {
		t.Fatal("ALLOW_INSECURE_DEFAULTS=true non pris en compte")
	}
}
