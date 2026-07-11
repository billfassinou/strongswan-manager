package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"strongswan-manager/internal/domain"
)

// TunnelRepo accède aux tunnels.
type TunnelRepo struct{ pool *pgxpool.Pool }

const tunnelCols = `id, name, gateway_id, peer_gateway_id, type, ike_version, local_addr, remote_addr,
	local_subnets, remote_subnets, auth_method, secret_ref, cert_ref, peer_cert_ref,
	proposals_ike, proposals_esp, pfs, status, security_score, config_version, created_at, updated_at`

func scanTunnel(row pgx.Row) (*domain.Tunnel, error) {
	t := &domain.Tunnel{}
	err := row.Scan(&t.ID, &t.Name, &t.GatewayID, &t.PeerGatewayID, &t.Type, &t.IKEVersion, &t.LocalAddr, &t.RemoteAddr,
		&t.LocalSubnets, &t.RemoteSubnets, &t.AuthMethod, &t.SecretRef, &t.CertRef, &t.PeerCertRef,
		&t.ProposalsIKE, &t.ProposalsESP, &t.PFS, &t.Status, &t.SecurityScore, &t.ConfigVersion,
		&t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return t, err
}

func (r *TunnelRepo) Create(ctx context.Context, t *domain.Tunnel) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO tunnels (`+tunnelCols+`)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)`,
		t.ID, t.Name, t.GatewayID, t.PeerGatewayID, t.Type, t.IKEVersion, t.LocalAddr, t.RemoteAddr,
		t.LocalSubnets, t.RemoteSubnets, t.AuthMethod, t.SecretRef, t.CertRef, t.PeerCertRef,
		t.ProposalsIKE, t.ProposalsESP, t.PFS, t.Status, t.SecurityScore, t.ConfigVersion,
		t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *TunnelRepo) Get(ctx context.Context, id string) (*domain.Tunnel, error) {
	return scanTunnel(r.pool.QueryRow(ctx, `SELECT `+tunnelCols+` FROM tunnels WHERE id=$1`, id))
}

func (r *TunnelRepo) List(ctx context.Context) ([]domain.Tunnel, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+tunnelCols+` FROM tunnels ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Tunnel
	for rows.Next() {
		t, err := scanTunnel(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

func (r *TunnelRepo) Update(ctx context.Context, t *domain.Tunnel) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE tunnels SET name=$2, peer_gateway_id=$3, type=$4, ike_version=$5, local_addr=$6, remote_addr=$7,
		 local_subnets=$8, remote_subnets=$9, auth_method=$10, secret_ref=$11, cert_ref=$12, peer_cert_ref=$13,
		 proposals_ike=$14, proposals_esp=$15, pfs=$16, status=$17, security_score=$18,
		 config_version=$19, updated_at=now() WHERE id=$1`,
		t.ID, t.Name, t.PeerGatewayID, t.Type, t.IKEVersion, t.LocalAddr, t.RemoteAddr,
		t.LocalSubnets, t.RemoteSubnets, t.AuthMethod, t.SecretRef, t.CertRef, t.PeerCertRef,
		t.ProposalsIKE, t.ProposalsESP, t.PFS, t.Status, t.SecurityScore, t.ConfigVersion)
	return err
}

func (r *TunnelRepo) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE tunnels SET status=$2 WHERE id=$1`, id, status)
	return err
}

func (r *TunnelRepo) Delete(ctx context.Context, id string) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM tunnels WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
