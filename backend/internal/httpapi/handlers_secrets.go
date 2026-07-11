// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"strongswan-manager/internal/auth"
	"strongswan-manager/internal/domain"
)

type secretRequest struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Value  string `json:"value"`
	UsedBy string `json:"used_by"`
}

var validSecretTypes = map[string]bool{
	domain.SecretPSK: true, domain.SecretEAP: true, domain.SecretXAuth: true,
}

// secretView masque toujours la valeur (jamais renvoyée en clair).
func secretView(s domain.Secret) map[string]any {
	return map[string]any{"id": s.ID, "name": s.Name, "type": s.Type, "used_by": s.UsedBy, "value": "••••••••"}
}

func (a *API) handleListSecrets(w http.ResponseWriter, r *http.Request) {
	list, err := a.Secrets.List(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "lecture des secrets impossible")
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, s := range list {
		out = append(out, secretView(s))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (a *API) handleCreateSecret(w http.ResponseWriter, r *http.Request) {
	var req secretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "bad_request", "corps JSON invalide")
		return
	}
	if req.Name == "" || req.Value == "" {
		writeError(w, r, http.StatusBadRequest, "bad_request", "name et value requis")
		return
	}
	if !validSecretTypes[req.Type] {
		writeError(w, r, http.StatusUnprocessableEntity, "validation_failed", "type de secret invalide (psk, eap, xauth)")
		return
	}
	enc, err := a.Cipher.Encrypt([]byte(req.Value))
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "chiffrement impossible")
		return
	}
	s := &domain.Secret{ID: uuid.NewString(), Name: req.Name, Type: req.Type, UsedBy: req.UsedBy, EncValue: enc}
	if err := a.Secrets.Create(r.Context(), s); err != nil {
		writeError(w, r, http.StatusConflict, "conflict", "un secret de ce nom existe déjà")
		return
	}
	p := auth.PrincipalFrom(r.Context())
	_ = a.Audit.Append(r.Context(), actorID(p), "secret.create", s.Name)
	writeJSON(w, http.StatusCreated, secretView(*s))
}

func (a *API) handleDeleteSecret(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := a.Secrets.Delete(r.Context(), id); err != nil {
		a.notFoundOrInternal(w, r, err, "secret introuvable")
		return
	}
	p := auth.PrincipalFrom(r.Context())
	_ = a.Audit.Append(r.Context(), actorID(p), "secret.delete", id)
	w.WriteHeader(http.StatusNoContent)
}
