// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"strongswan-manager/internal/domain"
)

// ConfigRepo gère les entités de configuration génériques (config_items).
type ConfigRepo struct{ pool *pgxpool.Pool }

func (r *ConfigRepo) Create(ctx context.Context, c *domain.ConfigItem) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO config_items (id, kind, name, data) VALUES ($1,$2,$3,$4)`,
		c.ID, c.Kind, c.Name, jsonOr(c.Data))
	return err
}

func (r *ConfigRepo) List(ctx context.Context, kind string) ([]domain.ConfigItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, kind, name, data FROM config_items WHERE kind=$1 ORDER BY name`, kind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ConfigItem
	for rows.Next() {
		var c domain.ConfigItem
		if err := rows.Scan(&c.ID, &c.Kind, &c.Name, &c.Data); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *ConfigRepo) Get(ctx context.Context, id string) (*domain.ConfigItem, error) {
	c := &domain.ConfigItem{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, kind, name, data FROM config_items WHERE id=$1`, id).
		Scan(&c.ID, &c.Kind, &c.Name, &c.Data)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (r *ConfigRepo) Update(ctx context.Context, id, name string, data json.RawMessage) error {
	ct, err := r.pool.Exec(ctx,
		`UPDATE config_items SET name=$2, data=$3, updated_at=now() WHERE id=$1`, id, name, jsonOr(data))
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ConfigRepo) Delete(ctx context.Context, id string) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM config_items WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func jsonOr(d json.RawMessage) json.RawMessage {
	if len(d) == 0 {
		return json.RawMessage("{}")
	}
	return d
}
