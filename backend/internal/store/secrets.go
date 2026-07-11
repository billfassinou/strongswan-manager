package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"strongswan-manager/internal/domain"
)

// SecretRepo accède au coffre de secrets (valeurs chiffrées au repos).
type SecretRepo struct{ pool *pgxpool.Pool }

func (r *SecretRepo) Create(ctx context.Context, s *domain.Secret) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO secrets (id, name, type, enc_value, used_by) VALUES ($1,$2,$3,$4,$5)`,
		s.ID, s.Name, s.Type, s.EncValue, s.UsedBy)
	return err
}

func (r *SecretRepo) GetByName(ctx context.Context, name string) (*domain.Secret, error) {
	s := &domain.Secret{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, type, enc_value, used_by FROM secrets WHERE name=$1`, name).
		Scan(&s.ID, &s.Name, &s.Type, &s.EncValue, &s.UsedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return s, err
}

// List renvoie les secrets SANS leur valeur chiffrée (métadonnées uniquement).
func (r *SecretRepo) List(ctx context.Context) ([]domain.Secret, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name, type, used_by FROM secrets ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Secret
	for rows.Next() {
		var s domain.Secret
		if err := rows.Scan(&s.ID, &s.Name, &s.Type, &s.UsedBy); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *SecretRepo) Delete(ctx context.Context, id string) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM secrets WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
