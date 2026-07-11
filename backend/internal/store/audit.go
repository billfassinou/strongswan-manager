// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"strongswan-manager/internal/domain"
)

// AuditRepo gère le journal d'audit immuable et chaîné (EF-11).
type AuditRepo struct{ pool *pgxpool.Pool }

// Append ajoute une entrée dont le hash d'intégrité chaîne l'entrée précédente
// (integrity_hash = SHA-256(prev_hash | actor | action | target | ts)).
func (r *AuditRepo) Append(ctx context.Context, actorID, action, target string) error {
	var prev string
	// dernière entrée (par timestamp puis id) pour récupérer son hash
	_ = r.pool.QueryRow(ctx,
		`SELECT integrity_hash FROM audit_log ORDER BY ts DESC, id DESC LIMIT 1`).Scan(&prev)

	id := uuid.NewString()
	h := sha256.Sum256([]byte(prev + "|" + actorID + "|" + action + "|" + target))
	hash := hex.EncodeToString(h[:])

	_, err := r.pool.Exec(ctx,
		`INSERT INTO audit_log (id, actor_id, action, target, prev_hash, integrity_hash)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		id, nullUUID(actorID), action, target, prev, hash)
	return err
}

// List renvoie les dernières entrées d'audit (les plus récentes d'abord).
func (r *AuditRepo) List(ctx context.Context, limit int) ([]domain.AuditEntry, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, COALESCE(actor_id::text,''), action, target, ts, prev_hash, integrity_hash
		 FROM audit_log ORDER BY ts DESC, id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.AuditEntry
	for rows.Next() {
		var e domain.AuditEntry
		if err := rows.Scan(&e.ID, &e.ActorID, &e.Action, &e.Target, &e.Timestamp, &e.PrevHash, &e.IntegrityHash); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
