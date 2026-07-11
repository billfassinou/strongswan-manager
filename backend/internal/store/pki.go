package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"strongswan-manager/internal/domain"
)

// CARepo accède à l'autorité de certification interne.
type CARepo struct{ pool *pgxpool.Pool }

func (r *CARepo) Create(ctx context.Context, c *domain.CA) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO cert_authorities (id, name, cert_pem, key_enc) VALUES ($1,$2,$3,$4)`,
		c.ID, c.Name, string(c.CertPEM), c.KeyEnc)
	return err
}

// Get renvoie la première CA (déploiement mono-CA de cette tranche).
func (r *CARepo) Get(ctx context.Context) (*domain.CA, error) {
	c := &domain.CA{}
	var pemStr, crlStr string
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, cert_pem, key_enc, crl_number, crl_pem FROM cert_authorities ORDER BY created_at LIMIT 1`).
		Scan(&c.ID, &c.Name, &pemStr, &c.KeyEnc, &c.CRLNumber, &crlStr)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	c.CertPEM = []byte(pemStr)
	c.CRLPEM = []byte(crlStr)
	return c, err
}

// UpdateCRL enregistre une nouvelle CRL et son numéro monotone.
func (r *CARepo) UpdateCRL(ctx context.Context, id string, number int64, crlPEM []byte) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE cert_authorities SET crl_number=$2, crl_pem=$3 WHERE id=$1`, id, number, string(crlPEM))
	return err
}

// CertRepo accède aux certificats émis.
type CertRepo struct{ pool *pgxpool.Pool }

func (r *CertRepo) Create(ctx context.Context, c *domain.Certificate) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO certificates (id, ca_id, name, cn, kind, serial, status, not_before, not_after, cert_pem, key_enc)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		c.ID, nullUUID(""), c.Name, c.CN, c.Kind, c.Serial, c.Status, c.NotBefore, c.NotAfter, string(c.CertPEM), c.KeyEnc)
	return err
}

func (r *CertRepo) GetByName(ctx context.Context, name string) (*domain.Certificate, error) {
	c := &domain.Certificate{}
	var pemStr string
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, cn, kind, serial, status, not_before, not_after, cert_pem, key_enc
		 FROM certificates WHERE name=$1`, name).
		Scan(&c.ID, &c.Name, &c.CN, &c.Kind, &c.Serial, &c.Status, &c.NotBefore, &c.NotAfter, &pemStr, &c.KeyEnc)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	c.CertPEM = []byte(pemStr)
	return c, err
}

// List renvoie les métadonnées des certificats (sans la clé privée).
func (r *CertRepo) List(ctx context.Context) ([]domain.Certificate, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, cn, kind, serial, status, not_before, not_after FROM certificates ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Certificate
	for rows.Next() {
		var c domain.Certificate
		if err := rows.Scan(&c.ID, &c.Name, &c.CN, &c.Kind, &c.Serial, &c.Status, &c.NotBefore, &c.NotAfter); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CertRepo) Revoke(ctx context.Context, id string) error {
	ct, err := r.pool.Exec(ctx, `UPDATE certificates SET status='revoked', revoked_at=now() WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListRevoked renvoie les certificats révoqués (série + date) pour générer la CRL.
func (r *CertRepo) ListRevoked(ctx context.Context) ([]domain.RevokedCert, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT serial, COALESCE(revoked_at, now()) FROM certificates WHERE status='revoked'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.RevokedCert
	for rows.Next() {
		var rc domain.RevokedCert
		if err := rows.Scan(&rc.Serial, &rc.RevokedAt); err != nil {
			return nil, err
		}
		out = append(out, rc)
	}
	return out, rows.Err()
}
