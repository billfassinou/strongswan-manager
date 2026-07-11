// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package vici

import (
	"strconv"

	"strongswan-manager/internal/domain"
)

// ConnName dérive le nom de connexion VICI d'un tunnel (unique, stable).
func ConnName(t *domain.Tunnel) string { return t.Name }

// ChildName dérive le nom de la CHILD_SA.
func ChildName(t *domain.Tunnel) string { return t.Name + "-net" }

// authKeyword traduit la méthode d'auth métier en mot-clé VICI.
func authKeyword(method string) string {
	switch method {
	case domain.AuthPSK:
		return "psk"
	case domain.AuthCert:
		return "pubkey"
	case domain.AuthEAP:
		return "eap"
	default:
		return "pubkey"
	}
}

// BuildConn construit la représentation VICI (map imbriquée) d'une connexion à
// partir d'un tunnel. Fonction PURE (aucune dépendance govici) → testable seule.
// Chaque élément de ProposalsIKE/ESP est une proposition complète (ex. "aes256-sha256-modp2048").
func BuildConn(t *domain.Tunnel) map[string]any {
	remoteAddrs := []string{"%any"}
	if t.Type != domain.TypeRoadWarrior && t.RemoteAddr != "" {
		remoteAddrs = []string{t.RemoteAddr}
	}

	child := map[string]any{
		"local_ts":      nonEmpty(t.LocalSubnets, []string{"0.0.0.0/0"}),
		"remote_ts":     nonEmpty(t.RemoteSubnets, []string{"0.0.0.0/0"}),
		"esp_proposals": t.ProposalsESP,
	}

	local := map[string]any{"auth": authKeyword(t.AuthMethod)}
	remote := map[string]any{"auth": authKeyword(t.AuthMethod)}
	// Identités IKE : indispensables à l'auth par certificat (matching cert ↔ SAN),
	// et utiles au PSK. On utilise les adresses (les certificats portent l'IP en SAN).
	if t.LocalAddr != "" && t.LocalAddr != "%any" {
		local["id"] = t.LocalAddr
	}
	if t.RemoteAddr != "" && t.RemoteAddr != "%any" {
		remote["id"] = t.RemoteAddr
	}

	conn := map[string]any{
		"version":      strconv.Itoa(t.IKEVersion),
		"local_addrs":  []string{t.LocalAddr},
		"remote_addrs": remoteAddrs,
		"proposals":    t.ProposalsIKE,
		"local":        local,
		"remote":       remote,
		"children": map[string]any{
			ChildName(t): child,
		},
	}
	return conn
}

func nonEmpty(v, def []string) []string {
	if len(v) == 0 {
		return def
	}
	return v
}
