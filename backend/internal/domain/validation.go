// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package domain

import (
	"net"
	"strings"
)

// FieldError décrit un problème de validation sur un champ précis (format §10).
type FieldError struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
}

// ValidationError agrège plusieurs FieldError ; se traduit en réponse 422.
type ValidationError struct {
	Details []FieldError
}

func (e *ValidationError) Error() string { return "validation_failed" }

// Add ajoute un constat de validation.
func (e *ValidationError) Add(field, issue string) {
	e.Details = append(e.Details, FieldError{Field: field, Issue: issue})
}

// HasErrors indique si au moins un constat bloquant a été relevé.
func (e *ValidationError) HasErrors() bool { return len(e.Details) > 0 }

var validTypes = map[string]bool{TypeSiteToSite: true, TypeRoadWarrior: true, TypeHostToHost: true}
var validAuth = map[string]bool{AuthPSK: true, AuthCert: true, AuthEAP: true}

// ValidateTunnel applique les invariants métier du §14.2. Les algorithmes faibles
// ne bloquent PAS (ils dégradent le score et remontent en warnings) ; en revanche
// une proposition explicitement dépréciée est signalée comme détail bloquant si
// strict=true (utilisé pour refuser les configs manifestement dangereuses).
func ValidateTunnel(t *Tunnel) *ValidationError {
	ve := &ValidationError{}

	if strings.TrimSpace(t.Name) == "" {
		ve.Add("name", "nom requis")
	}
	if t.GatewayID == "" {
		ve.Add("gateway_id", "passerelle requise")
	}
	if !validTypes[t.Type] {
		ve.Add("type", "type invalide (site-to-site, road-warrior, host-to-host)")
	}
	if t.IKEVersion != 1 && t.IKEVersion != 2 {
		ve.Add("ike_version", "ike_version doit valoir 1 ou 2")
	}
	if !validAuth[t.AuthMethod] {
		ve.Add("auth.method", "méthode d'authentification invalide (psk, cert, eap)")
	}
	if len(t.ProposalsIKE) == 0 {
		ve.Add("proposals.ike", "au moins une proposition IKE est requise")
	}
	if len(t.ProposalsESP) == 0 {
		ve.Add("proposals.esp", "au moins une proposition ESP est requise")
	}
	// road-warrior : l'extrémité distante peut être dynamique (%any) ; sinon adresse requise.
	if t.Type != TypeRoadWarrior && strings.TrimSpace(t.RemoteAddr) == "" {
		ve.Add("remote.addr", "adresse distante requise")
	}
	if strings.TrimSpace(t.LocalAddr) == "" {
		ve.Add("local.addr", "adresse locale requise")
	}
	validateSubnets("local.subnets", t.LocalSubnets, ve)
	validateSubnets("remote.subnets", t.RemoteSubnets, ve)

	// Proposition explicitement dangereuse (DH groupe 2) → bloquant, conformément à
	// l'exemple d'erreur 422 du §10.
	for _, p := range append(append([]string{}, t.ProposalsIKE...), t.ProposalsESP...) {
		if strings.Contains(strings.ToLower(p), "modp1024") {
			ve.Add("proposals.ike", "modp1024 déconseillé (DH groupe 2)")
			break
		}
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}

func validateSubnets(field string, subnets []string, ve *ValidationError) {
	for _, s := range subnets {
		s = strings.TrimSpace(s)
		if s == "" || s == "dynamique" || s == "%any" {
			continue
		}
		if _, _, err := net.ParseCIDR(s); err != nil {
			// tolère une IP nue (ex. host-to-host)
			if net.ParseIP(s) == nil {
				ve.Add(field, "CIDR/adresse invalide: "+s)
			}
		}
	}
}
