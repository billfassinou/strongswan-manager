package httpapi

import (
	"net/http"
	"testing"

	"strongswan-manager/internal/domain"
)

func TestGetCA(t *testing.T) {
	h := newHarness(t)
	w := h.do(http.MethodGet, "/api/v1/ca", h.token(domain.RoleAuditor), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("code %d", w.Code)
	}
	if !contains(w.Body.String(), "BEGIN CERTIFICATE") {
		t.Fatalf("CA PEM absent: %s", w.Body.String())
	}
}

func TestCreateCert(t *testing.T) {
	h := newHarness(t)
	w := h.do(http.MethodPost, "/api/v1/certificates", h.token(domain.RoleAdmin), map[string]any{
		"name": "gw-a-cert", "cn": "gw-a", "kind": "server", "sans": []string{"172.19.0.2"},
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("code %d, corps=%s", w.Code, w.Body.String())
	}
	m := decode(t, w)
	if m["status"] != "valid" || m["serial"] == "" {
		t.Fatalf("réponse inattendue: %v", m)
	}
	// la clé privée n'apparaît jamais dans la réponse
	if contains(w.Body.String(), "PRIVATE KEY") {
		t.Fatal("la clé privée a été exposée")
	}
	// stockée chiffrée
	c := h.api.Certs.(*fakeCerts).byName["gw-a-cert"]
	if c == nil || len(c.KeyEnc) == 0 || len(c.CertPEM) == 0 {
		t.Fatal("certificat/clé non stockés")
	}
	if contains(string(c.KeyEnc), "PRIVATE KEY") {
		t.Fatal("clé stockée en clair")
	}
}

func TestListCertsNoPrivateKey(t *testing.T) {
	h := newHarness(t)
	h.do(http.MethodPost, "/api/v1/certificates", h.token(domain.RoleAdmin), map[string]any{"name": "c1", "cn": "gw-a"})
	w := h.do(http.MethodGet, "/api/v1/certificates", h.token(domain.RoleAuditor), nil)
	if w.Code != http.StatusOK || contains(w.Body.String(), "PRIVATE KEY") {
		t.Fatalf("liste expose la clé ou code %d", w.Code)
	}
}

func TestRevokeCert(t *testing.T) {
	h := newHarness(t)
	h.do(http.MethodPost, "/api/v1/certificates", h.token(domain.RoleAdmin), map[string]any{"name": "c1", "cn": "gw-a"})
	id := h.api.Certs.(*fakeCerts).byName["c1"].ID
	w := h.do(http.MethodPost, "/api/v1/certificates/"+id+"/revoke", h.token(domain.RoleAdmin), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("revoke: code %d", w.Code)
	}
	if h.api.Certs.(*fakeCerts).byName["c1"].Status != domain.CertRevoked {
		t.Fatal("certificat non révoqué")
	}
}

func TestCertRBAC(t *testing.T) {
	h := newHarness(t)
	w := h.do(http.MethodPost, "/api/v1/certificates", h.token(domain.RoleViewer), map[string]any{"name": "c", "cn": "x"})
	if w.Code != http.StatusForbidden {
		t.Fatalf("viewer create cert: code %d (attendu 403)", w.Code)
	}
}

// TestCreateTunnelWithCertLoadsChain vérifie qu'un tunnel par certificat charge la CA,
// le certificat local et sa clé sur la passerelle (via VICI).
func TestCreateTunnelWithCertLoadsChain(t *testing.T) {
	h := newHarness(t)
	h.do(http.MethodPost, "/api/v1/certificates", h.token(domain.RoleAdmin), map[string]any{"name": "gw-a-cert", "cn": "gw-a", "sans": []string{"203.0.113.10"}})

	body := validTunnel("cert-tunnel")
	body["auth"] = map[string]any{"method": "cert", "cert_ref": "gw-a-cert"}
	w := h.do(http.MethodPost, "/api/v1/tunnels", h.token(domain.RoleAdmin), body)
	if w.Code != http.StatusCreated {
		t.Fatalf("code %d, corps=%s", w.Code, w.Body.String())
	}
	certs, caCerts, keys := h.mock.Counts()
	if certs < 1 || caCerts < 1 || keys < 1 {
		t.Fatalf("chargement incomplet: certs=%d ca=%d keys=%d", certs, caCerts, keys)
	}
}

// TestS2SCertBothGateways vérifie que le certificat de chaque extrémité est chargé sur
// sa passerelle respective (condition d'un établissement bout-en-bout par certificats).
func TestS2SCertBothGateways(t *testing.T) {
	h := newHarness(t)
	h.do(http.MethodPost, "/api/v1/certificates", h.token(domain.RoleAdmin), map[string]any{"name": "cert-a", "cn": "gw-a", "sans": []string{"203.0.113.10"}})
	h.do(http.MethodPost, "/api/v1/certificates", h.token(domain.RoleAdmin), map[string]any{"name": "cert-b", "cn": "gw-b", "sans": []string{"198.51.100.20"}})

	body := validTunnel("s2s-cert")
	body["peer_gateway_id"] = gw2ID
	body["peer_cert_ref"] = "cert-b"
	body["auth"] = map[string]any{"method": "cert", "cert_ref": "cert-a"}
	w := h.do(http.MethodPost, "/api/v1/tunnels", h.token(domain.RoleAdmin), body)
	if w.Code != http.StatusCreated {
		t.Fatalf("code %d, corps=%s", w.Code, w.Body.String())
	}
	if c, _, k := h.mock.Counts(); c < 1 || k < 1 {
		t.Fatalf("gw-a: cert/clé non chargés (c=%d k=%d)", c, k)
	}
	if c, _, k := h.mock2.Counts(); c < 1 || k < 1 {
		t.Fatalf("gw-b: cert/clé du pair non chargés (c=%d k=%d)", c, k)
	}
}
