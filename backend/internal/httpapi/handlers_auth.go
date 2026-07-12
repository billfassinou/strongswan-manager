// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"strongswan-manager/internal/auth"
	"strongswan-manager/internal/domain"
	"strongswan-manager/internal/store"
)

type loginRequest struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token              string `json:"token"`
	ExpiresAt          string `json:"expires_at"`
	Role               string `json:"role"`
	Identity           string `json:"identity"`
	MustChangePassword bool   `json:"must_change_password"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (a *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "bad_request", "corps JSON invalide")
		return
	}
	u, err := a.Users.GetByIdentity(r.Context(), req.Identity)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, r, http.StatusUnauthorized, "invalid_credentials", "identifiant ou mot de passe incorrect")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "internal", "erreur interne")
		return
	}
	if !u.Enabled {
		writeError(w, r, http.StatusForbidden, "account_disabled", "compte désactivé")
		return
	}
	if !auth.CheckPassword(u.PassHash, req.Password) {
		writeError(w, r, http.StatusUnauthorized, "invalid_credentials", "identifiant ou mot de passe incorrect")
		return
	}
	token, exp, err := a.Auth.Issue(u)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "émission du jeton impossible")
		return
	}
	_ = a.Audit.Append(r.Context(), u.ID, "login", u.Identity)
	writeJSON(w, http.StatusOK, loginResponse{
		Token:              token,
		ExpiresAt:          exp.UTC().Format("2006-01-02T15:04:05Z"),
		Role:               u.Role,
		Identity:           u.Identity,
		MustChangePassword: u.MustChangePassword,
	})
}

func (a *API) handleMe(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	writeJSON(w, http.StatusOK, map[string]any{
		"id":                   p.UserID,
		"identity":             p.Identity,
		"role":                 p.Role,
		"can_write":            p.CanWrite(),
		"must_change_password": p.MustChangePassword,
	})
}

// handleChangePassword change le mot de passe du compte connecté. Accessible à TOUS les
// rôles (y compris ceux en lecture seule : c'est leur propre compte, pas une configuration).
// En cas de succès, un nouveau jeton est émis — l'ancien porte encore le drapeau
// « must_change_password » et resterait bloqué par requirePasswordChanged.
func (a *API) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "bad_request", "corps JSON invalide")
		return
	}

	u, err := a.Users.GetByID(r.Context(), p.UserID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "erreur interne")
		return
	}
	if !auth.CheckPassword(u.PassHash, req.CurrentPassword) {
		writeError(w, r, http.StatusUnauthorized, "invalid_credentials", "mot de passe actuel incorrect")
		return
	}
	ve := &domain.ValidationError{}
	if err := auth.ValidatePassword(req.NewPassword); err != nil {
		ve.Add("new_password", fmt.Sprintf("au moins %d caractères", auth.MinPasswordLength))
	}
	if req.NewPassword == req.CurrentPassword {
		ve.Add("new_password", "doit différer du mot de passe actuel")
	}
	if len(ve.Details) > 0 {
		writeValidation(w, r, ve, "mot de passe refusé")
		return
	}

	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "erreur interne")
		return
	}
	if err := a.Users.SetPassword(r.Context(), u.ID, hash); err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "erreur interne")
		return
	}
	u.PassHash = hash
	u.MustChangePassword = false

	token, exp, err := a.Auth.Issue(u)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "émission du jeton impossible")
		return
	}
	_ = a.Audit.Append(r.Context(), u.ID, "password_change", u.Identity)
	writeJSON(w, http.StatusOK, loginResponse{
		Token:              token,
		ExpiresAt:          exp.UTC().Format("2006-01-02T15:04:05Z"),
		Role:               u.Role,
		Identity:           u.Identity,
		MustChangePassword: false,
	})
}

// requirePasswordChanged interdit tout usage de l'API tant que le compte utilise encore le
// mot de passe posé à l'installation. Sans ce garde-fou, le blocage ne vivrait que dans le
// front — et l'API resterait ouverte à qui connaît le mot de passe d'usine.
func (a *API) requirePasswordChanged(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := auth.PrincipalFrom(r.Context())
		if p != nil && p.MustChangePassword && r.URL.Path != "/api/v1/me" && r.URL.Path != "/api/v1/me/password" {
			writeError(w, r, http.StatusForbidden, "password_change_required",
				"mot de passe initial : changez-le (POST /api/v1/me/password) avant d'utiliser la console")
			return
		}
		next.ServeHTTP(w, r)
	})
}
