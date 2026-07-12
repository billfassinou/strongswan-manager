// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package httpapi

import "strongswan-manager/internal/domain"

// --- Requêtes (contrat §10) ---

type endpointDTO struct {
	Addr    string   `json:"addr"`
	Subnets []string `json:"subnets"`
}

type authDTO struct {
	Method    string `json:"method"`
	SecretRef string `json:"secret_ref"`
	CertRef   string `json:"cert_ref"`
}

type proposalsDTO struct {
	IKE []string `json:"ike"`
	ESP []string `json:"esp"`
}

// tunnelRequest correspond au corps de POST/PUT /api/v1/tunnels (§10.2).
type tunnelRequest struct {
	Name          string       `json:"name"`
	GatewayID     string       `json:"gateway_id"`
	PeerGatewayID string       `json:"peer_gateway_id"`
	PeerCertRef   string       `json:"peer_cert_ref"`
	Type          string       `json:"type"`
	IKEVersion    int          `json:"ike_version"`
	Local         endpointDTO  `json:"local"`
	Remote        endpointDTO  `json:"remote"`
	Auth          authDTO      `json:"auth"`
	Proposals     proposalsDTO `json:"proposals"`
	PFS           bool         `json:"pfs"`
}

func (req tunnelRequest) toDomain() *domain.Tunnel {
	var secretRef *string
	if req.Auth.SecretRef != "" {
		s := req.Auth.SecretRef
		secretRef = &s
	}
	var peer *string
	if req.PeerGatewayID != "" {
		p := req.PeerGatewayID
		peer = &p
	}
	var certRef *string
	if req.Auth.CertRef != "" {
		c := req.Auth.CertRef
		certRef = &c
	}
	var peerCert *string
	if req.PeerCertRef != "" {
		pc := req.PeerCertRef
		peerCert = &pc
	}
	return &domain.Tunnel{
		Name:          req.Name,
		GatewayID:     req.GatewayID,
		PeerGatewayID: peer,
		CertRef:       certRef,
		PeerCertRef:   peerCert,
		Type:          req.Type,
		IKEVersion:    req.IKEVersion,
		LocalAddr:     req.Local.Addr,
		RemoteAddr:    req.Remote.Addr,
		LocalSubnets:  req.Local.Subnets,
		RemoteSubnets: req.Remote.Subnets,
		AuthMethod:    req.Auth.Method,
		SecretRef:     secretRef,
		ProposalsIKE:  req.Proposals.IKE,
		ProposalsESP:  req.Proposals.ESP,
		PFS:           req.PFS,
	}
}

// --- Réponses ---

// tunnelResponse est la vue complète d'un tunnel renvoyée par l'API.
type tunnelResponse struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	GatewayID     string      `json:"gateway_id"`
	PeerGatewayID string      `json:"peer_gateway_id,omitempty"`
	Type          string      `json:"type"`
	IKEVersion    int         `json:"ike_version"`
	Local         endpointDTO `json:"local"`
	Remote        endpointDTO `json:"remote"`
	AuthMethod    string      `json:"auth_method"`
	ProposalsIKE  []string    `json:"proposals_ike"`
	ProposalsESP  []string    `json:"proposals_esp"`
	PFS           bool        `json:"pfs"`
	Status        string      `json:"status"`
	SecurityScore int         `json:"security_score"`
	Warnings      []string    `json:"warnings"`
	ConfigVersion int         `json:"config_version"`
}

func toTunnelResponse(t *domain.Tunnel) tunnelResponse {
	warnings := domain.ScoreTunnel(t).Notes
	if warnings == nil {
		warnings = []string{}
	}
	peer := ""
	if t.PeerGatewayID != nil {
		peer = *t.PeerGatewayID
	}
	return tunnelResponse{
		ID:            t.ID,
		Name:          t.Name,
		GatewayID:     t.GatewayID,
		PeerGatewayID: peer,
		Type:          t.Type,
		IKEVersion:    t.IKEVersion,
		Local:         endpointDTO{Addr: t.LocalAddr, Subnets: t.LocalSubnets},
		Remote:        endpointDTO{Addr: t.RemoteAddr, Subnets: t.RemoteSubnets},
		AuthMethod:    t.AuthMethod,
		ProposalsIKE:  t.ProposalsIKE,
		ProposalsESP:  t.ProposalsESP,
		PFS:           t.PFS,
		Status:        t.Status,
		SecurityScore: t.SecurityScore,
		Warnings:      warnings,
		ConfigVersion: t.ConfigVersion,
	}
}
