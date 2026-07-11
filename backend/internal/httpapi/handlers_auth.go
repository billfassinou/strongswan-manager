package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"strongswan-manager/internal/auth"
	"strongswan-manager/internal/store"
)

type loginRequest struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	Role      string `json:"role"`
	Identity  string `json:"identity"`
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
		Token:     token,
		ExpiresAt: exp.UTC().Format("2006-01-02T15:04:05Z"),
		Role:      u.Role,
		Identity:  u.Identity,
	})
}

func (a *API) handleMe(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFrom(r.Context())
	writeJSON(w, http.StatusOK, map[string]any{
		"id":        p.UserID,
		"identity":  p.Identity,
		"role":      p.Role,
		"can_write": p.CanWrite(),
	})
}
