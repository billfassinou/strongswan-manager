package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"strongswan-manager/internal/domain"
)

// ErrNotFound est renvoyé lorsqu'une ressource n'existe pas.
var ErrNotFound = errors.New("not found")

// UserRepo accède aux comptes d'administration.
type UserRepo struct{ pool *pgxpool.Pool }

// Create insère un compte.
func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users_admin (id, identity, pass_hash, role, enabled) VALUES ($1,$2,$3,$4,$5)`,
		u.ID, u.Identity, u.PassHash, u.Role, u.Enabled)
	return err
}

// GetByIdentity récupère un compte par identifiant.
func (r *UserRepo) GetByIdentity(ctx context.Context, identity string) (*domain.User, error) {
	u := &domain.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, identity, pass_hash, role, enabled FROM users_admin WHERE identity=$1`, identity).
		Scan(&u.ID, &u.Identity, &u.PassHash, &u.Role, &u.Enabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

// GetByID récupère un compte par identifiant technique.
func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	u := &domain.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, identity, pass_hash, role, enabled FROM users_admin WHERE id=$1`, id).
		Scan(&u.ID, &u.Identity, &u.PassHash, &u.Role, &u.Enabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

// Count renvoie le nombre de comptes (pour décider du seed initial).
func (r *UserRepo) Count(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM users_admin`).Scan(&n)
	return n, err
}
