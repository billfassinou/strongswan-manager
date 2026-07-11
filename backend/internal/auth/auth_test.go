// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"strongswan-manager/internal/domain"
)

func TestPasswordHashCheck(t *testing.T) {
	h, err := HashPassword("s3cret")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword(h, "s3cret") {
		t.Fatal("mot de passe correct refusé")
	}
	if CheckPassword(h, "wrong") {
		t.Fatal("mot de passe incorrect accepté")
	}
}

func TestJWTIssueParse(t *testing.T) {
	m := NewManager("secret", time.Hour)
	u := &domain.User{ID: "u1", Identity: "nadia", Role: domain.RoleAdmin}
	tok, exp, err := m.Issue(u)
	if err != nil {
		t.Fatal(err)
	}
	if !exp.After(time.Now()) {
		t.Fatal("expiration dans le passé")
	}
	p, err := m.Parse(tok)
	if err != nil {
		t.Fatal(err)
	}
	if p.UserID != "u1" || p.Identity != "nadia" || p.Role != domain.RoleAdmin {
		t.Fatalf("principal inattendu: %+v", p)
	}
	if !p.CanWrite() {
		t.Fatal("admin devrait pouvoir écrire")
	}
}

func TestJWTRejectsWrongSecret(t *testing.T) {
	tok, _, _ := NewManager("secret-a", time.Hour).Issue(&domain.User{ID: "x", Role: domain.RoleViewer})
	if _, err := NewManager("secret-b", time.Hour).Parse(tok); err == nil {
		t.Fatal("un jeton signé avec un autre secret aurait dû être rejeté")
	}
}

func TestJWTExpired(t *testing.T) {
	m := NewManager("secret", -time.Minute) // déjà expiré
	tok, _, _ := m.Issue(&domain.User{ID: "x", Role: domain.RoleViewer})
	if _, err := m.Parse(tok); err == nil {
		t.Fatal("un jeton expiré aurait dû être rejeté")
	}
}

func TestMiddlewareRejectsMissingOrBadToken(t *testing.T) {
	m := NewManager("secret", time.Hour)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h := m.Middleware(next)

	// sans en-tête
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("sans token: %d", w.Code)
	}
	// jeton invalide
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not-a-jwt")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("jeton invalide: %d", w.Code)
	}
}

func TestMiddlewareInjectsPrincipalAndRequireWrite(t *testing.T) {
	m := NewManager("secret", time.Hour)
	tok, _, _ := m.Issue(&domain.User{ID: "u1", Identity: "op", Role: domain.RoleOperator})

	var seen *Principal
	protected := m.Middleware(RequireWrite(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = PrincipalFrom(r.Context())
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	protected.ServeHTTP(w, req)
	if w.Code != http.StatusOK || seen == nil || seen.Identity != "op" {
		t.Fatalf("principal non injecté (code %d, principal %+v)", w.Code, seen)
	}

	// viewer → 403 via RequireWrite
	vtok, _, _ := m.Issue(&domain.User{ID: "u2", Role: domain.RoleViewer})
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer "+vtok)
	w = httptest.NewRecorder()
	protected.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("viewer via RequireWrite: %d (attendu 403)", w.Code)
	}
}

func TestRoleCanWrite(t *testing.T) {
	for _, r := range []string{domain.RoleAdmin, domain.RoleOperator} {
		if !domain.RoleCanWrite(r) {
			t.Fatalf("%s devrait écrire", r)
		}
	}
	for _, r := range []string{domain.RoleAuditor, domain.RoleViewer, "inconnu"} {
		if domain.RoleCanWrite(r) {
			t.Fatalf("%s ne devrait pas écrire", r)
		}
	}
}
