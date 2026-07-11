package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"strongswan-manager/internal/domain"
)

// GatewayRepo accède aux passerelles StrongSwan gérées.
type GatewayRepo struct{ pool *pgxpool.Pool }

func (r *GatewayRepo) Create(ctx context.Context, g *domain.Gateway) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO gateways (id, name, endpoint, version, status) VALUES ($1,$2,$3,$4,$5)
		 ON CONFLICT (name) DO UPDATE SET endpoint=EXCLUDED.endpoint`,
		g.ID, g.Name, g.Endpoint, g.Version, g.Status)
	return err
}

func (r *GatewayRepo) List(ctx context.Context) ([]domain.Gateway, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, endpoint, version, status, created_at FROM gateways ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Gateway
	for rows.Next() {
		var g domain.Gateway
		if err := rows.Scan(&g.ID, &g.Name, &g.Endpoint, &g.Version, &g.Status, &g.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r *GatewayRepo) Get(ctx context.Context, id string) (*domain.Gateway, error) {
	g := &domain.Gateway{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, endpoint, version, status, created_at FROM gateways WHERE id=$1`, id).
		Scan(&g.ID, &g.Name, &g.Endpoint, &g.Version, &g.Status, &g.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return g, err
}

func (r *GatewayRepo) GetByName(ctx context.Context, name string) (*domain.Gateway, error) {
	g := &domain.Gateway{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, endpoint, version, status, created_at FROM gateways WHERE name=$1`, name).
		Scan(&g.ID, &g.Name, &g.Endpoint, &g.Version, &g.Status, &g.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return g, err
}

func (r *GatewayRepo) UpdateStatus(ctx context.Context, id, status, version string) error {
	_, err := r.pool.Exec(ctx, `UPDATE gateways SET status=$2, version=$3 WHERE id=$1`, id, status, version)
	return err
}
