package vici

import "testing"

import "strongswan-manager/internal/domain"

func TestBuildConn(t *testing.T) {
	tun := &domain.Tunnel{
		Name: "paris-dakar", Type: domain.TypeSiteToSite, IKEVersion: 2,
		LocalAddr: "203.0.113.10", RemoteAddr: "198.51.100.20",
		LocalSubnets: []string{"10.1.0.0/16"}, RemoteSubnets: []string{"10.2.0.0/16"},
		AuthMethod:   domain.AuthPSK,
		ProposalsIKE: []string{"aes256-sha256-modp2048"}, ProposalsESP: []string{"aes256gcm16"},
	}
	conn := BuildConn(tun)

	if conn["version"] != "2" {
		t.Fatalf("version = %v", conn["version"])
	}
	local := conn["local"].(map[string]any)
	if local["auth"] != "psk" {
		t.Fatalf("auth = %v", local["auth"])
	}
	children := conn["children"].(map[string]any)
	if _, ok := children["paris-dakar-net"]; !ok {
		t.Fatalf("child manquant: %v", children)
	}
	child := children["paris-dakar-net"].(map[string]any)
	ts := child["remote_ts"].([]string)
	if len(ts) != 1 || ts[0] != "10.2.0.0/16" {
		t.Fatalf("remote_ts = %v", ts)
	}
}

func TestBuildConnRoadWarrior(t *testing.T) {
	tun := &domain.Tunnel{
		Name: "rw", Type: domain.TypeRoadWarrior, IKEVersion: 2, LocalAddr: "203.0.113.10",
		AuthMethod: domain.AuthEAP, ProposalsIKE: []string{"aes256gcm16-prfsha384-ecp384"},
		ProposalsESP: []string{"aes256gcm16"},
	}
	conn := BuildConn(tun)
	ra := conn["remote_addrs"].([]string)
	if len(ra) != 1 || ra[0] != "%any" {
		t.Fatalf("road warrior remote_addrs = %v (attendu %%any)", ra)
	}
}
