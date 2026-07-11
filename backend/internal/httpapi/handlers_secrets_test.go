// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package httpapi

import (
	"net/http"
	"testing"

	"strongswan-manager/internal/domain"
)

func TestCreateSecret(t *testing.T) {
	h := newHarness(t)
	w := h.do(http.MethodPost, "/api/v1/secrets", h.token(domain.RoleAdmin), map[string]string{
		"name": "psk-dakar", "type": "psk", "value": "s3cret-partage", "used_by": "paris-dakar",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("code %d (attendu 201), corps=%s", w.Code, w.Body.String())
	}
	m := decode(t, w)
	if m["name"] != "psk-dakar" {
		t.Fatalf("name = %v", m["name"])
	}
	// la valeur ne doit jamais être renvoyée en clair
	if m["value"] == "s3cret-partage" {
		t.Fatal("la valeur du secret a été exposée en clair")
	}
	// stockée chiffrée
	s := h.secrets.byName["psk-dakar"]
	if s == nil || len(s.EncValue) == 0 {
		t.Fatal("secret non stocké chiffré")
	}
	if string(s.EncValue) == "s3cret-partage" {
		t.Fatal("secret stocké en clair")
	}
}

func TestCreateSecretInvalidType(t *testing.T) {
	h := newHarness(t)
	w := h.do(http.MethodPost, "/api/v1/secrets", h.token(domain.RoleAdmin), map[string]string{
		"name": "x", "type": "magic", "value": "v",
	})
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("code %d (attendu 422)", w.Code)
	}
}

func TestListSecretsNeverExposesValue(t *testing.T) {
	h := newHarness(t)
	h.do(http.MethodPost, "/api/v1/secrets", h.token(domain.RoleAdmin), map[string]string{"name": "a", "type": "psk", "value": "topsecret"})
	w := h.do(http.MethodGet, "/api/v1/secrets", h.token(domain.RoleAuditor), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("code %d", w.Code)
	}
	if got := w.Body.String(); got == "" || contains(got, "topsecret") {
		t.Fatalf("la valeur en clair figure dans la liste: %s", got)
	}
}

func TestDeleteSecret(t *testing.T) {
	h := newHarness(t)
	h.do(http.MethodPost, "/api/v1/secrets", h.token(domain.RoleAdmin), map[string]string{"name": "a", "type": "psk", "value": "v"})
	id := h.secrets.byName["a"].ID
	w := h.do(http.MethodDelete, "/api/v1/secrets/"+id, h.token(domain.RoleAdmin), nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("code %d (attendu 204)", w.Code)
	}
	if _, ok := h.secrets.byName["a"]; ok {
		t.Fatal("secret non supprimé")
	}
}

func TestSecretRBAC(t *testing.T) {
	h := newHarness(t)
	w := h.do(http.MethodPost, "/api/v1/secrets", h.token(domain.RoleViewer), map[string]string{"name": "a", "type": "psk", "value": "v"})
	if w.Code != http.StatusForbidden {
		t.Fatalf("viewer create secret: code %d (attendu 403)", w.Code)
	}
}

// TestCreateTunnelWithPSKLoadsShared vérifie que le secret PSK est bien chargé sur la
// passerelle via load-shared lors de l'application du tunnel.
func TestCreateTunnelWithPSKLoadsShared(t *testing.T) {
	h := newHarness(t)
	// secret référencé par le tunnel
	h.do(http.MethodPost, "/api/v1/secrets", h.token(domain.RoleAdmin), map[string]string{
		"name": "psk-dakar", "type": "psk", "value": "shared-key",
	})
	body := validTunnel("paris-dakar")
	body["auth"] = map[string]any{"method": "psk", "secret_ref": "psk-dakar"}
	w := h.do(http.MethodPost, "/api/v1/tunnels", h.token(domain.RoleAdmin), body)
	if w.Code != http.StatusCreated {
		t.Fatalf("code %d, corps=%s", w.Code, w.Body.String())
	}
	if !h.mock.HasConn("paris-dakar") {
		t.Fatal("connexion non chargée")
	}
	if !h.mock.HasShared("psk-dakar") {
		t.Fatal("le secret PSK n'a pas été chargé via load-shared")
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
