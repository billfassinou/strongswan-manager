package httpapi

import (
	"net/http"
	"testing"

	"strongswan-manager/internal/domain"
)

// TestSiteToSiteConfiguresBothGateways vérifie qu'un tunnel site-à-site avec
// peer_gateway_id charge la connexion ET le PSK sur les DEUX passerelles (locale + pair),
// avec les extrémités inversées côté pair — condition de l'établissement bout-en-bout.
func TestSiteToSiteConfiguresBothGateways(t *testing.T) {
	h := newHarness(t)
	h.do(http.MethodPost, "/api/v1/secrets", h.token(domain.RoleAdmin), map[string]string{
		"name": "psk-s2s", "type": "psk", "value": "clef",
	})

	body := validTunnel("hub")
	body["peer_gateway_id"] = gw2ID
	body["auth"] = map[string]any{"method": "psk", "secret_ref": "psk-s2s"}
	w := h.do(http.MethodPost, "/api/v1/tunnels", h.token(domain.RoleAdmin), body)
	if w.Code != http.StatusCreated {
		t.Fatalf("code %d, corps=%s", w.Code, w.Body.String())
	}
	id := decode(t, w)["id"].(string)

	// connexion + PSK sur les deux passerelles
	if !h.mock.HasConn("hub") || !h.mock2.HasConn("hub") {
		t.Fatal("la connexion doit être chargée sur les deux passerelles")
	}
	if !h.mock.HasShared("psk-s2s") || !h.mock2.HasShared("psk-s2s") {
		t.Fatal("le PSK doit être chargé sur les deux passerelles")
	}

	// suppression → unload sur les deux passerelles
	if w := h.do(http.MethodDelete, "/api/v1/tunnels/"+id, h.token(domain.RoleAdmin), nil); w.Code != http.StatusNoContent {
		t.Fatalf("delete: %d", w.Code)
	}
	if h.mock.HasConn("hub") || h.mock2.HasConn("hub") {
		t.Fatal("la connexion doit être déchargée des deux passerelles")
	}
}

func TestMirrorTunnelSwapsEndpoints(t *testing.T) {
	peer := "gw-b"
	orig := &domain.Tunnel{
		Name: "hub", GatewayID: "gw-a", PeerGatewayID: &peer,
		LocalAddr: "10.0.0.1", RemoteAddr: "10.0.0.2",
		LocalSubnets: []string{"192.168.10.0/24"}, RemoteSubnets: []string{"192.168.20.0/24"},
	}
	m := mirrorTunnel(orig)
	if m.LocalAddr != "10.0.0.2" || m.RemoteAddr != "10.0.0.1" {
		t.Fatalf("adresses non inversées: %s / %s", m.LocalAddr, m.RemoteAddr)
	}
	if m.LocalSubnets[0] != "192.168.20.0/24" || m.RemoteSubnets[0] != "192.168.10.0/24" {
		t.Fatalf("sous-réseaux non inversés: %v / %v", m.LocalSubnets, m.RemoteSubnets)
	}
	// l'original n'est pas modifié
	if orig.LocalAddr != "10.0.0.1" {
		t.Fatal("mirrorTunnel a muté l'original")
	}
}
