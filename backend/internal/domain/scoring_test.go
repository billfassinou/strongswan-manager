package domain

import "testing"

func TestScoreTunnel(t *testing.T) {
	cases := []struct {
		name string
		in   *Tunnel
		want int
	}{
		{
			name: "moderne IKEv2 + ML-KEM + PFS",
			in:   &Tunnel{IKEVersion: 2, PFS: true, ProposalsIKE: []string{"aes256gcm16", "sha384", "ecp384", "mlkem768"}, ProposalsESP: []string{"aes256gcm16"}},
			want: 100,
		},
		{
			name: "IKEv2 sans ML-KEM",
			in:   &Tunnel{IKEVersion: 2, PFS: true, ProposalsIKE: []string{"aes256", "sha256", "modp2048"}, ProposalsESP: []string{"aes256", "sha256"}},
			want: 94,
		},
		{
			name: "legacy IKEv1 3des md5 modp1024",
			in:   &Tunnel{IKEVersion: 1, PFS: false, ProposalsIKE: []string{"3des", "md5", "modp1024"}, ProposalsESP: []string{"3des", "md5"}},
			want: 5, // 100-42-28-18-16-6 = -10 -> borné à 5 (modp présent => PFS non pénalisé)
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ScoreTunnel(c.in).Score
			if got != c.want {
				t.Fatalf("score = %d, attendu %d", got, c.want)
			}
		})
	}
}

func TestValidateTunnelRejectsWeakDH(t *testing.T) {
	tun := &Tunnel{
		Name: "x", GatewayID: "gw", Type: TypeSiteToSite, IKEVersion: 2, AuthMethod: AuthPSK,
		LocalAddr: "203.0.113.1", RemoteAddr: "198.51.100.1",
		ProposalsIKE: []string{"aes256-sha256-modp1024"}, ProposalsESP: []string{"aes256gcm16"},
	}
	ve := ValidateTunnel(tun)
	if ve == nil || !ve.HasErrors() {
		t.Fatal("modp1024 aurait dû être rejeté (422)")
	}
}
