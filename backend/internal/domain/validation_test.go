package domain

import "testing"

func validBase() *Tunnel {
	return &Tunnel{
		Name: "ok", GatewayID: "gw", Type: TypeSiteToSite, IKEVersion: 2, AuthMethod: AuthPSK,
		LocalAddr: "203.0.113.10", RemoteAddr: "198.51.100.20",
		LocalSubnets: []string{"10.1.0.0/16"}, RemoteSubnets: []string{"10.2.0.0/16"},
		ProposalsIKE: []string{"aes256-sha256-modp2048"}, ProposalsESP: []string{"aes256gcm16"},
	}
}

func TestValidateTunnelValid(t *testing.T) {
	if ve := ValidateTunnel(validBase()); ve != nil {
		t.Fatalf("tunnel valide rejeté: %+v", ve.Details)
	}
}

func TestValidateTunnelFields(t *testing.T) {
	cases := []struct {
		name  string
		mut   func(*Tunnel)
		field string
	}{
		{"nom vide", func(t *Tunnel) { t.Name = "" }, "name"},
		{"passerelle vide", func(t *Tunnel) { t.GatewayID = "" }, "gateway_id"},
		{"type invalide", func(t *Tunnel) { t.Type = "mesh" }, "type"},
		{"ike invalide", func(t *Tunnel) { t.IKEVersion = 3 }, "ike_version"},
		{"auth invalide", func(t *Tunnel) { t.AuthMethod = "magic" }, "auth.method"},
		{"sans proposition ike", func(t *Tunnel) { t.ProposalsIKE = nil }, "proposals.ike"},
		{"sans proposition esp", func(t *Tunnel) { t.ProposalsESP = nil }, "proposals.esp"},
		{"adresse locale vide", func(t *Tunnel) { t.LocalAddr = "" }, "local.addr"},
		{"CIDR invalide", func(t *Tunnel) { t.LocalSubnets = []string{"999.0.0.0/8"} }, "local.subnets"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tun := validBase()
			c.mut(tun)
			ve := ValidateTunnel(tun)
			if ve == nil {
				t.Fatalf("aurait dû échouer")
			}
			found := false
			for _, d := range ve.Details {
				if d.Field == c.field {
					found = true
				}
			}
			if !found {
				t.Fatalf("champ %q non signalé: %+v", c.field, ve.Details)
			}
		})
	}
}

func TestRoadWarriorAllowsDynamicRemote(t *testing.T) {
	tun := validBase()
	tun.Type = TypeRoadWarrior
	tun.RemoteAddr = "" // dynamique
	tun.RemoteSubnets = []string{"dynamique"}
	tun.AuthMethod = AuthEAP
	if ve := ValidateTunnel(tun); ve != nil {
		t.Fatalf("road warrior sans remote.addr rejeté: %+v", ve.Details)
	}
}

func TestHostToHostAllowsBareIP(t *testing.T) {
	tun := validBase()
	tun.Type = TypeHostToHost
	tun.LocalSubnets = []string{"192.0.2.44"}
	tun.RemoteSubnets = []string{"51.20.0.9"}
	if ve := ValidateTunnel(tun); ve != nil {
		t.Fatalf("host-to-host avec IP nue rejeté: %+v", ve.Details)
	}
}
