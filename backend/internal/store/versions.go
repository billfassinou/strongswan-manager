// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"strongswan-manager/internal/domain"
)

// VersionRepo accède à l'historique de configuration des tunnels (versioning interne).
type VersionRepo struct{ pool *pgxpool.Pool }

func (r *VersionRepo) Create(ctx context.Context, v *domain.ConfigVersion) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO config_versions (id, tunnel_id, n, author_id, message, snapshot)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		v.ID, v.TunnelID, v.N, nullUUID(v.AuthorID), v.Message, v.Snapshot)
	return err
}

// NextN renvoie le prochain numéro de version pour un tunnel.
func (r *VersionRepo) NextN(ctx context.Context, tunnelID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(MAX(n),0)+1 FROM config_versions WHERE tunnel_id=$1`, tunnelID).Scan(&n)
	return n, err
}

func (r *VersionRepo) ListByTunnel(ctx context.Context, tunnelID string) ([]domain.ConfigVersion, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tunnel_id, n, COALESCE(author_id::text,''), message, snapshot, created_at
		 FROM config_versions WHERE tunnel_id=$1 ORDER BY n DESC`, tunnelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ConfigVersion
	for rows.Next() {
		var v domain.ConfigVersion
		if err := rows.Scan(&v.ID, &v.TunnelID, &v.N, &v.AuthorID, &v.Message, &v.Snapshot, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// Get récupère une version précise d'un tunnel.
func (r *VersionRepo) Get(ctx context.Context, tunnelID string, n int) (*domain.ConfigVersion, error) {
	v := &domain.ConfigVersion{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tunnel_id, n, COALESCE(author_id::text,''), message, snapshot, created_at
		 FROM config_versions WHERE tunnel_id=$1 AND n=$2`, tunnelID, n).
		Scan(&v.ID, &v.TunnelID, &v.N, &v.AuthorID, &v.Message, &v.Snapshot, &v.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return v, err
}

func nullUUID(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
