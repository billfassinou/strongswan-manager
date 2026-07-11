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

// Types d'entités de configuration gérés par le CRUD générique (§ EF-21 à EF-25, EF-07/09).
var configKinds = map[string]bool{
	"pool": true, "radius": true, "policy": true, "authority": true,
	"vpnuser": true, "alert": true, "daemon": true,
}

type configRequest struct {
	Name string          `json:"name"`
	Data json.RawMessage `json:"data"`
}

func configView(c domain.ConfigItem) map[string]any {
	var data any
	_ = json.Unmarshal(c.Data, &data)
	return map[string]any{"id": c.ID, "kind": c.Kind, "name": c.Name, "data": data}
}

func (a *API) handleListConfig(w http.ResponseWriter, r *http.Request) {
	kind := chi.URLParam(r, "kind")
	if !configKinds[kind] {
		writeError(w, r, http.StatusNotFound, "unknown_kind", "type de configuration inconnu")
		return
	}
	items, err := a.Config.List(r.Context(), kind)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "lecture impossible")
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, c := range items {
		out = append(out, configView(c))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (a *API) handleCreateConfig(w http.ResponseWriter, r *http.Request) {
	kind := chi.URLParam(r, "kind")
	if !configKinds[kind] {
		writeError(w, r, http.StatusNotFound, "unknown_kind", "type de configuration inconnu")
		return
	}
	var req configRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "bad_request", "corps JSON invalide")
		return
	}
	if req.Name == "" {
		writeError(w, r, http.StatusBadRequest, "bad_request", "name requis")
		return
	}
	c := &domain.ConfigItem{ID: uuid.NewString(), Kind: kind, Name: req.Name, Data: req.Data}
	if err := a.Config.Create(r.Context(), c); err != nil {
		writeError(w, r, http.StatusConflict, "conflict", "un élément de ce nom existe déjà")
		return
	}
	audit(a, r, "config."+kind+".create", req.Name)
	writeJSON(w, http.StatusCreated, configView(*c))
}

func (a *API) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req configRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "bad_request", "corps JSON invalide")
		return
	}
	if err := a.Config.Update(r.Context(), id, req.Name, req.Data); err != nil {
		a.notFoundOrInternal(w, r, err, "élément introuvable")
		return
	}
	audit(a, r, "config.update", req.Name)
	c, _ := a.Config.Get(r.Context(), id)
	if c != nil {
		writeJSON(w, http.StatusOK, configView(*c))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "updated"})
}

func (a *API) handleDeleteConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := a.Config.Delete(r.Context(), id); err != nil {
		a.notFoundOrInternal(w, r, err, "élément introuvable")
		return
	}
	audit(a, r, "config.delete", id)
	w.WriteHeader(http.StatusNoContent)
}

func audit(a *API, r *http.Request, action, target string) {
	p := auth.PrincipalFrom(r.Context())
	_ = a.Audit.Append(r.Context(), actorID(p), action, target)
}
