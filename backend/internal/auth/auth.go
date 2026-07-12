// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

// Package auth gère l'authentification locale (mot de passe + JWT) et le RBAC.
package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"strongswan-manager/internal/domain"
)

// HashPassword calcule un hash bcrypt du mot de passe.
func HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

// CheckPassword compare un mot de passe en clair à son hash.
func CheckPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

// MinPasswordLength est la longueur minimale imposée. On mise sur la longueur plutôt que
// sur des règles de composition : c'est ce que recommande le NIST (SP 800-63B), et c'est
// ce qui résiste réellement à une attaque hors ligne sur le hash.
const MinPasswordLength = 12

// ErrWeakPassword indique un mot de passe refusé par la politique.
var ErrWeakPassword = errors.New("mot de passe trop court")

// ValidatePassword applique la politique de mot de passe de la console.
func ValidatePassword(pw string) error {
	if len([]rune(pw)) < MinPasswordLength {
		return ErrWeakPassword
	}
	return nil
}

// Principal est l'utilisateur authentifié porté par le contexte de la requête.
type Principal struct {
	UserID   string
	Identity string
	Role     string
	// MustChangePassword est porté par le jeton lui-même : la console peut ainsi
	// refuser toute action sans aller interroger la base à chaque requête. Le
	// changement de mot de passe émet un nouveau jeton, sans ce drapeau.
	MustChangePassword bool
}

// CanWrite indique si le principal peut effectuer des actions modifiantes.
func (p Principal) CanWrite() bool { return domain.RoleCanWrite(p.Role) }

// Claims sont les revendications encodées dans le JWT.
type Claims struct {
	Identity           string `json:"identity"`
	Role               string `json:"role"`
	MustChangePassword bool   `json:"must_change_password,omitempty"`
	jwt.RegisteredClaims
}

// Manager émet et vérifie les JWT.
type Manager struct {
	secret []byte
	ttl    time.Duration
}

// NewManager construit un gestionnaire de jetons.
func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl}
}

// Issue crée un JWT signé pour un utilisateur.
func (m *Manager) Issue(u *domain.User) (string, time.Time, error) {
	exp := time.Now().Add(m.ttl)
	claims := Claims{
		Identity:           u.Identity,
		Role:               u.Role,
		MustChangePassword: u.MustChangePassword,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID,
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(m.secret)
	return signed, exp, err
}

// Parse valide un JWT et renvoie le principal.
func (m *Manager) Parse(tokenStr string) (*Principal, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("méthode de signature inattendue")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	return &Principal{
		UserID:             claims.Subject,
		Identity:           claims.Identity,
		Role:               claims.Role,
		MustChangePassword: claims.MustChangePassword,
	}, nil
}

type ctxKey int

const principalKey ctxKey = 0

// Middleware exige un Bearer token valide et injecte le principal dans le contexte.
func (m *Manager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			unauthorized(w)
			return
		}
		p, err := m.Parse(strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			unauthorized(w)
			return
		}
		ctx := context.WithValue(r.Context(), principalKey, p)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireWrite bloque les rôles en lecture seule (403).
func RequireWrite(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := PrincipalFrom(r.Context())
		if p == nil || !p.CanWrite() {
			forbidden(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// PrincipalFrom extrait le principal du contexte (nil si absent).
func PrincipalFrom(ctx context.Context) *Principal {
	p, _ := ctx.Value(principalKey).(*Principal)
	return p
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized","message":"authentification requise"}`))
}

func forbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_, _ = w.Write([]byte(`{"error":"forbidden","message":"action réservée — rôle en lecture seule"}`))
}
