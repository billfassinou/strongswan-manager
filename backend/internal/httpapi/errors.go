// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"strongswan-manager/internal/domain"
)

// writeJSON sérialise v en JSON avec le code HTTP donné.
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// errorBody est le format d'erreur structuré du §10.
type errorBody struct {
	Error         string              `json:"error"`
	Message       string              `json:"message"`
	Details       []domain.FieldError `json:"details,omitempty"`
	CorrelationID string              `json:"correlation_id"`
}

// writeError renvoie une erreur simple au format §10.
func writeError(w http.ResponseWriter, r *http.Request, code int, errKey, message string) {
	writeJSON(w, code, errorBody{Error: errKey, Message: message, CorrelationID: middleware.GetReqID(r.Context())})
}

// writeValidation renvoie une 422 avec le détail des champs invalides (§10).
func writeValidation(w http.ResponseWriter, r *http.Request, ve *domain.ValidationError, message string) {
	writeJSON(w, http.StatusUnprocessableEntity, errorBody{
		Error:         "validation_failed",
		Message:       message,
		Details:       ve.Details,
		CorrelationID: middleware.GetReqID(r.Context()),
	})
}
