// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package httpapi

import (
	"net/http"
	"strconv"
)

func (a *API) handleAudit(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	entries, err := a.Audit.List(r.Context(), limit)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "lecture de l'audit impossible")
		return
	}
	out := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		out = append(out, map[string]any{
			"id": e.ID, "actor_id": e.ActorID, "action": e.Action, "target": e.Target,
			"timestamp": e.Timestamp.UTC().Format("2006-01-02T15:04:05Z"),
			"integrity_hash": e.IntegrityHash, "prev_hash": e.PrevHash,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}
