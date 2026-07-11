package httpapi

import (
	"net/http"
	"testing"

	"strongswan-manager/internal/domain"
)

func TestConfigCRUD(t *testing.T) {
	h := newHarness(t)
	tok := h.token(domain.RoleAdmin)

	// create pool
	w := h.do(http.MethodPost, "/api/v1/config/pool", tok, map[string]any{
		"name": "pool-rw", "data": map[string]any{"range": "10.9.0.0/24", "source": "Interne", "dns": "10.1.0.53"},
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: code %d, corps=%s", w.Code, w.Body.String())
	}
	id := decode(t, w)["id"].(string)

	// list
	w = h.do(http.MethodGet, "/api/v1/config/pool", tok, nil)
	items := decode(t, w)["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("%d pools (attendu 1)", len(items))
	}

	// update
	w = h.do(http.MethodPut, "/api/v1/config/pool/"+id, tok, map[string]any{
		"name": "pool-rw", "data": map[string]any{"range": "10.9.0.0/22"},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("update: code %d", w.Code)
	}

	// delete
	if w := h.do(http.MethodDelete, "/api/v1/config/pool/"+id, tok, nil); w.Code != http.StatusNoContent {
		t.Fatalf("delete: code %d", w.Code)
	}
}

func TestConfigUnknownKind(t *testing.T) {
	h := newHarness(t)
	w := h.do(http.MethodGet, "/api/v1/config/dragons", h.token(domain.RoleAdmin), nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("kind inconnu: code %d (attendu 404)", w.Code)
	}
}

func TestConfigRBAC(t *testing.T) {
	h := newHarness(t)
	// lecture autorisée à l'auditeur
	if w := h.do(http.MethodGet, "/api/v1/config/radius", h.token(domain.RoleAuditor), nil); w.Code != http.StatusOK {
		t.Fatalf("auditor GET: %d", w.Code)
	}
	// écriture interdite au viewer
	if w := h.do(http.MethodPost, "/api/v1/config/radius", h.token(domain.RoleViewer), map[string]any{"name": "x"}); w.Code != http.StatusForbidden {
		t.Fatalf("viewer POST: %d (attendu 403)", w.Code)
	}
}

func TestConfigKinds(t *testing.T) {
	h := newHarness(t)
	tok := h.token(domain.RoleAdmin)
	for _, k := range []string{"pool", "radius", "policy", "authority", "vpnuser", "alert", "daemon"} {
		w := h.do(http.MethodPost, "/api/v1/config/"+k, tok, map[string]any{"name": "x-" + k, "data": map[string]any{"v": 1}})
		if w.Code != http.StatusCreated {
			t.Fatalf("kind %s create: code %d", k, w.Code)
		}
	}
}
