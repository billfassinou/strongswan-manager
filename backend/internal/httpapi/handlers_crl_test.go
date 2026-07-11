package httpapi

import (
	"net/http"
	"testing"

	"strongswan-manager/internal/domain"
)

func TestGetCRL(t *testing.T) {
	h := newHarness(t)
	w := h.do(http.MethodGet, "/api/v1/crl", h.token(domain.RoleAuditor), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("code %d", w.Code)
	}
	if !contains(w.Body.String(), "BEGIN X509 CRL") {
		t.Fatalf("CRL PEM absente: %s", w.Body.String())
	}
}

func TestRevokeRegeneratesCRL(t *testing.T) {
	h := newHarness(t)
	h.do(http.MethodPost, "/api/v1/certificates", h.token(domain.RoleAdmin), map[string]any{"name": "cert-a", "cn": "gw-a", "sans": []string{"203.0.113.10"}})
	id := h.api.Certs.(*fakeCerts).byName["cert-a"].ID

	before := h.api.CA.(*fakeCA).ca.CRLNumber
	w := h.do(http.MethodPost, "/api/v1/certificates/"+id+"/revoke", h.token(domain.RoleAdmin), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("revoke: code %d", w.Code)
	}
	// la CRL a été régénérée (numéro incrémenté) et persistée
	if h.api.CA.(*fakeCA).ca.CRLNumber <= before {
		t.Fatal("le numéro de CRL n'a pas été incrémenté")
	}
	if len(h.api.CA.(*fakeCA).ca.CRLPEM) == 0 {
		t.Fatal("CRL non persistée")
	}
}

func TestCRLDerPublicNoAuth(t *testing.T) {
	h := newHarness(t)
	// endpoint CDP public : accessible SANS jeton (les passerelles y accèdent en clair)
	w := h.do(http.MethodGet, "/crl.der", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("/crl.der: code %d (attendu 200 sans auth)", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/pkix-crl" {
		t.Fatalf("Content-Type = %q", w.Header().Get("Content-Type"))
	}
	// c'est bien du DER (commence par SEQUENCE 0x30), pas du PEM
	if b := w.Body.Bytes(); len(b) == 0 || b[0] != 0x30 {
		t.Fatal("le corps n'est pas un CRL DER")
	}
}

func TestPublishCRL(t *testing.T) {
	h := newHarness(t)
	w := h.do(http.MethodPost, "/api/v1/crl/publish", h.token(domain.RoleAdmin), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("publish: code %d", w.Code)
	}
	if _, ok := decode(t, w)["crl_number"]; !ok {
		t.Fatal("crl_number absent de la réponse")
	}
}

func TestCRLPublishRBAC(t *testing.T) {
	h := newHarness(t)
	if w := h.do(http.MethodPost, "/api/v1/crl/publish", h.token(domain.RoleViewer), nil); w.Code != http.StatusForbidden {
		t.Fatalf("viewer publish CRL: code %d (attendu 403)", w.Code)
	}
}
