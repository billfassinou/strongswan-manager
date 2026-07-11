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

// ServerTLSRepo persiste le certificat TLS du serveur lui-même.
//
// La table ne contient qu'une ligne : Save remplace donc la précédente. C'est
// volontaire — on ne veut jamais deux certificats serveur concurrents.
type ServerTLSRepo struct{ pool *pgxpool.Pool }

// Get renvoie le certificat serveur courant, ou ErrNotFound s'il n'y en a pas.
func (r *ServerTLSRepo) Get(ctx context.Context) (*domain.ServerCert, error) {
	c := &domain.ServerCert{}
	var pemStr string
	err := r.pool.QueryRow(ctx,
		`SELECT id, cert_pem, key_enc, sans, not_after FROM server_tls ORDER BY created_at DESC LIMIT 1`).
		Scan(&c.ID, &pemStr, &c.KeyEnc, &c.SANs, &c.NotAfter)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	c.CertPEM = []byte(pemStr)
	return c, err
}

// Save remplace le certificat serveur (une seule ligne conservée).
func (r *ServerTLSRepo) Save(ctx context.Context, c *domain.ServerCert) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM server_tls`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO server_tls (id, cert_pem, key_enc, sans, not_after) VALUES ($1,$2,$3,$4,$5)`,
		c.ID, string(c.CertPEM), c.KeyEnc, c.SANs, c.NotAfter); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
