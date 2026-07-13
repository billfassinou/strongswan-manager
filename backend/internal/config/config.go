// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

// Package config charge la configuration depuis l'environnement (12-factor).
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Valeurs par défaut destinées au lab : elles sont publiques, donc inutilisables en
// production. Validate refuse de laisser démarrer le serveur avec.
const (
	DefaultJWTSecret  = "dev-insecure-change-me"
	DefaultSecretsKey = "dev-insecure-secrets-key-change-me"
)

// Config regroupe tous les paramètres d'exécution du serveur.
type Config struct {
	HTTPAddr          string        // adresse d'écoute principale, TLS si TLSEnabled (ex. ":7926")
	DatabaseURL       string        // DSN PostgreSQL
	JWTSecret         string        // secret de signature des JWT
	JWTTTL            time.Duration // durée de vie des jetons
	SeedAdminPassword string        // mot de passe du compte admin seedé au démarrage
	ViciEndpoints     []string      // passerelles VICI à enrôler au démarrage: "nom=réseau:adresse"
	PollInterval      time.Duration // période de sondage VICI
	SecretsKey        string        // clé de chiffrement applicatif des secrets (32 octets base64/hex ou passphrase)
	CORSOrigins       []string      // origines autorisées pour le front
	CRLURL            string        // URL publique du CRL Distribution Point (dans les certificats)
	CRLValidity       time.Duration // fenêtre nextUpdate des CRL

	// TLS. L'écouteur principal sert en HTTPS par défaut ; l'écouteur clair
	// (RedirectAddr) sert le CDP /crl.der et redirige le reste vers HTTPS.
	TLSEnabled   bool     // false derrière un reverse proxy qui termine déjà le TLS
	RedirectAddr string   // écouteur en clair : /crl.der + redirection 308 (ex. ":7927")
	TLSCert      string   // chemin d'un certificat PEM fourni (prioritaire sur l'auto-génération)
	TLSKey       string   // chemin de la clé privée PEM correspondante
	TLSSans      []string // SAN du certificat auto-généré (défaut: localhost/127.0.0.1/::1/hostname)
	ACMEDomain   string   // si défini: Let's Encrypt. Exige un domaine public joignable en HTTP-01.
	ACMEEmail    string   // contact ACME (avis d'expiration)
	ACMECache    string   // répertoire de cache des certificats ACME

	// AllowInsecureDefaults autorise le démarrage avec les secrets de développement.
	// Posé par le docker-compose du lab et par les tests ; JAMAIS en production.
	AllowInsecureDefaults bool
}

// Validate refuse de démarrer avec les secrets de développement, qui sont publics (ils
// sont dans le dépôt). Sans ce garde-fou, une installation « qui marche » livre une
// console dont n'importe qui peut forger les jetons et déchiffrer les secrets.
func (c Config) Validate() error {
	if c.AllowInsecureDefaults {
		return nil
	}
	const remedy = "Générez-en une : openssl rand -hex 32. " +
		"⚠️ SECRETS_KEY chiffre les secrets, la CA et la clé TLS : fixez-la AVANT le premier " +
		"démarrage, elle ne peut plus être changée ensuite. " +
		"Pour un lab, et uniquement pour un lab : ALLOW_INSECURE_DEFAULTS=true"

	jwtBad := c.JWTSecret == DefaultJWTSecret
	keyBad := c.SecretsKey == DefaultSecretsKey
	switch {
	case jwtBad && keyBad:
		return fmt.Errorf("JWT_SECRET et SECRETS_KEY gardent leur valeur de développement, qui est publique. %s", remedy)
	case jwtBad:
		return fmt.Errorf("JWT_SECRET garde sa valeur de développement, qui est publique. %s", remedy)
	case keyBad:
		return fmt.Errorf("SECRETS_KEY garde sa valeur de développement, qui est publique. %s", remedy)
	}
	return nil
}

// Load lit la configuration depuis les variables d'environnement, avec des
// valeurs par défaut adaptées au lab dockerisé.
func Load() Config {
	return Config{
		HTTPAddr:          env("HTTP_ADDR", ":7926"),
		DatabaseURL:       env("DATABASE_URL", "postgres://swan:swan@postgres:5432/swan?sslmode=disable"),
		JWTSecret:         env("JWT_SECRET", "dev-insecure-change-me"),
		JWTTTL:            envDuration("JWT_TTL", time.Hour),
		SeedAdminPassword: env("SEED_ADMIN_PASSWORD", "admin1234"),
		ViciEndpoints:     envList("VICI_ENDPOINTS", nil),
		PollInterval:      envDuration("POLL_INTERVAL", 3*time.Second),
		SecretsKey:        env("SECRETS_KEY", "dev-insecure-secrets-key-change-me"),
		CORSOrigins:       envList("CORS_ORIGINS", []string{"*"}),
		CRLURL:            env("CRL_URL", ""),
		CRLValidity:       envDuration("CRL_VALIDITY", 24*time.Hour),

		TLSEnabled:   envBool("TLS_ENABLED", true),
		RedirectAddr: env("HTTP_REDIRECT_ADDR", ":7927"),
		TLSCert:      env("TLS_CERT", ""),
		TLSKey:       env("TLS_KEY", ""),
		TLSSans:      envList("TLS_SANS", nil),
		ACMEDomain:   env("ACME_DOMAIN", ""),
		ACMEEmail:    env("ACME_EMAIL", ""),
		ACMECache:    env("ACME_CACHE", "./acme"),

		AllowInsecureDefaults: envBool("ALLOW_INSECURE_DEFAULTS", false),
	}
}

func envBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
		if n, err := strconv.Atoi(v); err == nil {
			return time.Duration(n) * time.Second
		}
	}
	return def
}

func envList(key string, def []string) []string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		parts := strings.Split(v, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if p = strings.TrimSpace(p); p != "" {
				out = append(out, p)
			}
		}
		return out
	}
	return def
}
