// Package store gère la connexion PostgreSQL, les migrations et les repositories.
package store

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"strongswan-manager/migrations"
)

// Store encapsule le pool de connexions et expose les repositories.
type Store struct {
	Pool     *pgxpool.Pool
	Users    *UserRepo
	Gateways *GatewayRepo
	Tunnels  *TunnelRepo
	Versions *VersionRepo
	Audit    *AuditRepo
	Secrets  *SecretRepo
	CA       *CARepo
	Certs    *CertRepo
	Config   *ConfigRepo
}

// New ouvre le pool et instancie les repositories.
func New(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connexion postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	s := &Store{Pool: pool}
	s.Users = &UserRepo{pool}
	s.Gateways = &GatewayRepo{pool}
	s.Tunnels = &TunnelRepo{pool}
	s.Versions = &VersionRepo{pool}
	s.Audit = &AuditRepo{pool}
	s.Secrets = &SecretRepo{pool}
	s.CA = &CARepo{pool}
	s.Certs = &CertRepo{pool}
	s.Config = &ConfigRepo{pool}
	return s, nil
}

// Close libère le pool.
func (s *Store) Close() { s.Pool.Close() }

// Migrate applique les migrations *.up.sql embarquées non encore appliquées.
// Runner minimal (pas de dépendance externe) : suit les versions dans schema_migrations.
func (s *Store) Migrate(ctx context.Context) error {
	if _, err := s.Pool.Exec(ctx,
		`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil {
		return fmt.Errorf("table schema_migrations: %w", err)
	}

	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return err
	}
	var ups []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			ups = append(ups, e.Name())
		}
	}
	sort.Strings(ups)

	for _, name := range ups {
		version := strings.TrimSuffix(name, ".up.sql")
		var exists bool
		if err := s.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, version).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}
		sqlBytes, err := migrations.FS.ReadFile(name)
		if err != nil {
			return err
		}
		if _, err := s.Pool.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
		if _, err := s.Pool.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES ($1)`, version); err != nil {
			return err
		}
	}
	return nil
}
