// Package config charge la configuration depuis l'environnement (12-factor).
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config regroupe tous les paramètres d'exécution du serveur.
type Config struct {
	HTTPAddr    string        // adresse d'écoute HTTP (ex. ":8080")
	DatabaseURL string        // DSN PostgreSQL
	JWTSecret   string        // secret de signature des JWT
	JWTTTL      time.Duration // durée de vie des jetons
	SeedAdminPassword string  // mot de passe du compte admin seedé au démarrage
	ViciEndpoints []string    // passerelles VICI à enrôler au démarrage: "nom=réseau:adresse"
	PollInterval  time.Duration // période de sondage VICI
	SecretsKey    string      // clé de chiffrement applicatif des secrets (32 octets base64/hex ou passphrase)
	CORSOrigins   []string    // origines autorisées pour le front
	CRLURL        string        // URL publique du CRL Distribution Point (dans les certificats)
	CRLValidity   time.Duration // fenêtre nextUpdate des CRL
}

// Load lit la configuration depuis les variables d'environnement, avec des
// valeurs par défaut adaptées au lab dockerisé.
func Load() Config {
	return Config{
		HTTPAddr:          env("HTTP_ADDR", ":8080"),
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
	}
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
