// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSPAHandlerServesIndex(t *testing.T) {
	h := Handler()

	// racine -> index.html du build
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("/ code %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "StrongSwan Manager") {
		t.Fatalf("index.html inattendu: %s", w.Body.String()[:min(120, w.Body.Len())])
	}

	// route côté client inconnue -> repli sur l'app (index.html)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/tunnels", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "<div id=\"root\">") {
		t.Fatalf("repli SPA absent (code %d)", w.Code)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
