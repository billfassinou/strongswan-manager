//go:build integration

// Package test regroupe les tests d'intégration (boîte noire) du backend : ils
// n'utilisent que l'API exportée des paquets et s'exécutent contre de vraies
// dépendances (PostgreSQL).
//
// Les tests UNITAIRES restent, eux, dans le paquet qu'ils testent (règle Go : un
// fichier _test.go doit être dans le même répertoire/paquet pour accéder aux
// éléments non exportés et alimenter la couverture du paquet).
//
//	Lancer :  make test-integration      (démarre un postgres jetable)
//	Ou :      DATABASE_URL=postgres://... go test -tags integration ./test/
package test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"strongswan-manager/internal/domain"
	"strongswan-manager/internal/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL non défini — test d'intégration ignoré")
	}
	ctx := context.Background()
	st, err := store.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connexion: %v", err)
	}
	if err := st.Migrate(ctx); err != nil {
		t.Fatalf("migrations: %v", err)
	}
	// table rase entre les runs
	_, _ = st.Pool.Exec(ctx, `TRUNCATE tunnels, gateways, users_admin, config_versions, audit_log, secrets, certificates, cert_authorities, config_items RESTART IDENTITY CASCADE`)
	return st
}

func TestIntegrationConfigItems(t *testing.T) {
	ctx := context.Background()
	st := newTestStore(t)
	defer st.Close()

	c := &domain.ConfigItem{ID: uuid.NewString(), Kind: "pool", Name: "pool-rw", Data: []byte(`{"range":"10.9.0.0/24"}`)}
	if err := st.Config.Create(ctx, c); err != nil {
		t.Fatal(err)
	}
	// unicité (kind,name)
	dup := *c
	dup.ID = uuid.NewString()
	if err := st.Config.Create(ctx, &dup); err == nil {
		t.Fatal("doublon (kind,name) aurait dû être rejeté")
	}
	// filtrage par kind
	other := &domain.ConfigItem{ID: uuid.NewString(), Kind: "radius", Name: "r1", Data: []byte(`{}`)}
	_ = st.Config.Create(ctx, other)
	pools, _ := st.Config.List(ctx, "pool")
	if len(pools) != 1 || pools[0].Name != "pool-rw" {
		t.Fatalf("List(pool) = %+v", pools)
	}
	// update + delete
	if err := st.Config.Update(ctx, c.ID, "pool-rw", []byte(`{"range":"10.9.0.0/22"}`)); err != nil {
		t.Fatal(err)
	}
	got, _ := st.Config.Get(ctx, c.ID)
	if string(got.Data) != `{"range": "10.9.0.0/22"}` && string(got.Data) != `{"range":"10.9.0.0/22"}` {
		t.Fatalf("data après update: %s", got.Data)
	}
	if err := st.Config.Delete(ctx, c.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Config.Get(ctx, c.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("après delete: %v", err)
	}
}

func TestIntegrationPKI(t *testing.T) {
	ctx := context.Background()
	st := newTestStore(t)
	defer st.Close()

	ca := &domain.CA{ID: uuid.NewString(), Name: "Root CA", CertPEM: []byte("-----BEGIN CERTIFICATE-----\nx\n-----END CERTIFICATE-----"), KeyEnc: []byte{0x01, 0x02}}
	if err := st.CA.Create(ctx, ca); err != nil {
		t.Fatal(err)
	}
	got, err := st.CA.Get(ctx)
	if err != nil || got.Name != "Root CA" || string(got.CertPEM) != string(ca.CertPEM) {
		t.Fatalf("CA.Get: %v / %+v", err, got)
	}

	c := &domain.Certificate{
		ID: uuid.NewString(), Name: "gw-a-cert", CN: "gw-a", Kind: domain.CertKindServer,
		Serial: "abcd", Status: domain.CertValid,
		NotBefore: time.Now(), NotAfter: time.Now().AddDate(1, 0, 0),
		CertPEM: []byte("pem"), KeyEnc: []byte{0x03, 0x04},
	}
	if err := st.Certs.Create(ctx, c); err != nil {
		t.Fatal(err)
	}
	got2, err := st.Certs.GetByName(ctx, "gw-a-cert")
	if err != nil || string(got2.KeyEnc) != string(c.KeyEnc) {
		t.Fatalf("Certs.GetByName: %v / %+v", err, got2)
	}
	// List ne renvoie ni PEM ni clé
	list, _ := st.Certs.List(ctx)
	if len(list) != 1 || len(list[0].KeyEnc) != 0 || len(list[0].CertPEM) != 0 {
		t.Fatalf("List expose la clé/PEM: %+v", list)
	}
	if err := st.Certs.Revoke(ctx, c.ID); err != nil {
		t.Fatal(err)
	}
	after, _ := st.Certs.GetByName(ctx, "gw-a-cert")
	if after.Status != domain.CertRevoked {
		t.Fatalf("statut après révocation: %s", after.Status)
	}

	// CRL : le certificat révoqué apparaît dans la liste, et la CRL se persiste
	revoked, err := st.Certs.ListRevoked(ctx)
	if err != nil || len(revoked) != 1 || revoked[0].Serial != "abcd" {
		t.Fatalf("ListRevoked: %v / %+v", err, revoked)
	}
	if err := st.CA.UpdateCRL(ctx, got.ID, 1, []byte("-----BEGIN X509 CRL-----\nx\n-----END X509 CRL-----")); err != nil {
		t.Fatal(err)
	}
	ca2, _ := st.CA.Get(ctx)
	if ca2.CRLNumber != 1 || len(ca2.CRLPEM) == 0 {
		t.Fatalf("CRL non persistée: number=%d len=%d", ca2.CRLNumber, len(ca2.CRLPEM))
	}
}

func TestIntegrationSecrets(t *testing.T) {
	ctx := context.Background()
	st := newTestStore(t)
	defer st.Close()

	s := &domain.Secret{ID: uuid.NewString(), Name: "psk-dakar", Type: domain.SecretPSK, UsedBy: "paris-dakar", EncValue: []byte{0x01, 0x02, 0x03}}
	if err := st.Secrets.Create(ctx, s); err != nil {
		t.Fatal(err)
	}
	// unicité du nom
	dup := *s
	dup.ID = uuid.NewString()
	if err := st.Secrets.Create(ctx, &dup); err == nil {
		t.Fatal("doublon de nom aurait dû être rejeté")
	}
	got, err := st.Secrets.GetByName(ctx, "psk-dakar")
	if err != nil || string(got.EncValue) != string(s.EncValue) {
		t.Fatalf("GetByName: %v / %+v", err, got)
	}
	// List ne renvoie pas la valeur chiffrée
	list, _ := st.Secrets.List(ctx)
	if len(list) != 1 || len(list[0].EncValue) != 0 {
		t.Fatalf("List expose la valeur ou compte incorrect: %+v", list)
	}
	if err := st.Secrets.Delete(ctx, s.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Secrets.GetByName(ctx, "psk-dakar"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("après delete: %v", err)
	}
}

func TestIntegrationUsers(t *testing.T) {
	ctx := context.Background()
	st := newTestStore(t)
	defer st.Close()

	u := &domain.User{ID: uuid.NewString(), Identity: "nadia", Role: domain.RoleAdmin, Enabled: true, PassHash: "h"}
	if err := st.Users.Create(ctx, u); err != nil {
		t.Fatal(err)
	}
	got, err := st.Users.GetByIdentity(ctx, "nadia")
	if err != nil || got.ID != u.ID {
		t.Fatalf("GetByIdentity: %v / %+v", err, got)
	}
	if _, err := st.Users.GetByIdentity(ctx, "ghost"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("attendu ErrNotFound, obtenu %v", err)
	}
	if n, _ := st.Users.Count(ctx); n != 1 {
		t.Fatalf("Count = %d", n)
	}
}

func TestIntegrationTunnelLifecycle(t *testing.T) {
	ctx := context.Background()
	st := newTestStore(t)
	defer st.Close()

	gw := &domain.Gateway{ID: uuid.NewString(), Name: "gw-a", Endpoint: "mock", Status: "up"}
	if err := st.Gateways.Create(ctx, gw); err != nil {
		t.Fatal(err)
	}

	tun := &domain.Tunnel{
		ID: uuid.NewString(), Name: "paris-dakar", GatewayID: gw.ID, Type: domain.TypeSiteToSite,
		IKEVersion: 2, LocalAddr: "203.0.113.10", RemoteAddr: "198.51.100.20",
		LocalSubnets: []string{"10.1.0.0/16"}, RemoteSubnets: []string{"10.2.0.0/16"},
		AuthMethod: domain.AuthPSK, ProposalsIKE: []string{"aes256-sha256-modp2048"},
		ProposalsESP: []string{"aes256gcm16"}, PFS: true, Status: "installing",
		SecurityScore: 94, ConfigVersion: 1,
	}
	if err := st.Tunnels.Create(ctx, tun); err != nil {
		t.Fatal(err)
	}

	// conflit d'unicité (même gateway + name)
	dup := *tun
	dup.ID = uuid.NewString()
	if err := st.Tunnels.Create(ctx, &dup); err == nil {
		t.Fatal("le doublon (gateway,name) aurait dû être rejeté")
	}

	got, err := st.Tunnels.Get(ctx, tun.ID)
	if err != nil || got.Name != "paris-dakar" || len(got.LocalSubnets) != 1 {
		t.Fatalf("Get: %v / %+v", err, got)
	}

	got.ConfigVersion = 2
	got.SecurityScore = 100
	if err := st.Tunnels.Update(ctx, got); err != nil {
		t.Fatal(err)
	}
	if err := st.Tunnels.UpdateStatus(ctx, tun.ID, domain.StatusUp); err != nil {
		t.Fatal(err)
	}
	after, _ := st.Tunnels.Get(ctx, tun.ID)
	if after.Status != domain.StatusUp || after.ConfigVersion != 2 {
		t.Fatalf("après update: %+v", after)
	}

	list, _ := st.Tunnels.List(ctx)
	if len(list) != 1 {
		t.Fatalf("List = %d", len(list))
	}

	// versions
	n, _ := st.Versions.NextN(ctx, tun.ID)
	if n != 1 {
		t.Fatalf("NextN = %d", n)
	}
	if err := st.Versions.Create(ctx, &domain.ConfigVersion{ID: uuid.NewString(), TunnelID: tun.ID, N: 1, Message: "création", Snapshot: []byte(`{}`)}); err != nil {
		t.Fatal(err)
	}
	if v, err := st.Versions.Get(ctx, tun.ID, 1); err != nil || v.Message != "création" {
		t.Fatalf("Versions.Get: %v / %+v", err, v)
	}

	if err := st.Tunnels.Delete(ctx, tun.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Tunnels.Get(ctx, tun.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("après delete, attendu ErrNotFound, obtenu %v", err)
	}
}

func TestIntegrationAuditChainAndImmutability(t *testing.T) {
	ctx := context.Background()
	st := newTestStore(t)
	defer st.Close()

	for _, a := range []string{"login", "tunnel.create", "tunnel.delete"} {
		if err := st.Audit.Append(ctx, uuid.NewString(), a, "target-"+a); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := st.Audit.List(ctx, 10)
	if err != nil || len(entries) != 3 {
		t.Fatalf("List: %v / %d entrées", err, len(entries))
	}
	// chaînage : chaque entrée porte un hash d'intégrité non vide
	for _, e := range entries {
		if e.IntegrityHash == "" {
			t.Fatal("integrity_hash vide")
		}
	}
	// immuabilité : toute tentative d'UPDATE doit être rejetée par le trigger.
	_, err = st.Pool.Exec(ctx, `UPDATE audit_log SET action='falsifié'`)
	if err == nil {
		t.Fatal("le journal d'audit doit être immuable (UPDATE aurait dû échouer)")
	}
}
