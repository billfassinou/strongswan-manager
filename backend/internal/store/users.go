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

// ErrNotFound est renvoyé lorsqu'une ressource n'existe pas.
var ErrNotFound = errors.New("not found")

// UserRepo accède aux comptes d'administration.
type UserRepo struct{ pool *pgxpool.Pool }

// Create insère un compte.
func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users_admin (id, identity, pass_hash, role, enabled, must_change_password)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		u.ID, u.Identity, u.PassHash, u.Role, u.Enabled, u.MustChangePassword)
	return err
}

// GetByIdentity récupère un compte par identifiant.
func (r *UserRepo) GetByIdentity(ctx context.Context, identity string) (*domain.User, error) {
	u := &domain.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, identity, pass_hash, role, enabled, must_change_password
		 FROM users_admin WHERE identity=$1`, identity).
		Scan(&u.ID, &u.Identity, &u.PassHash, &u.Role, &u.Enabled, &u.MustChangePassword)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

// GetByID récupère un compte par identifiant technique.
func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	u := &domain.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, identity, pass_hash, role, enabled, must_change_password
		 FROM users_admin WHERE id=$1`, id).
		Scan(&u.ID, &u.Identity, &u.PassHash, &u.Role, &u.Enabled, &u.MustChangePassword)
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

// SetPassword remplace le mot de passe d'un compte et lève l'obligation de changement.
func (r *UserRepo) SetPassword(ctx context.Context, id, passHash string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE users_admin SET pass_hash=$2, must_change_password=false WHERE id=$1`, id, passHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
