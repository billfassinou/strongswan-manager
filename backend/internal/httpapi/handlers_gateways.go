// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package httpapi

import "net/http"

func (a *API) handleListGateways(w http.ResponseWriter, r *http.Request) {
	gws, err := a.Gateways.List(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "lecture des passerelles impossible")
		return
	}
	out := make([]map[string]any, 0, len(gws))
	for _, g := range gws {
		out = append(out, map[string]any{
			"id": g.ID, "name": g.Name, "endpoint": g.Endpoint,
			"version": g.Version, "status": g.Status,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}
