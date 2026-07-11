// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package tlsx

import (
	"context"
	"crypto/x509"
	"net"
	"testing"
	"time"

	"strongswan-manager/internal/domain"
	"strongswan-manager/internal/pki"
	"strongswan-manager/internal/secrets"
	"strongswan-manager/internal/store"
)

// --- doublures en mémoire ---

type fakeRepo struct {
	cur   *domain.ServerCert
	saves int
}

func (f *fakeRepo) Get(context.Context) (*domain.ServerCert, error) {
	if f.cur == nil {
		return nil, store.ErrNotFound
	}
	return f.cur, nil
}
func (f *fakeRepo) Save(_ context.Context, c *domain.ServerCert) error {
	f.cur = c
	f.saves++
	return nil
}

type errRepo struct{ err error }

func (e errRepo) Get(context.Context) (*domain.ServerCert, error) { return nil, e.err }
func (e errRepo) Save(context.Context, *domain.ServerCert) error  { return nil }

type fakeCA struct{ ca *domain.CA }

func (f fakeCA) Get(context.Context) (*domain.CA, error) { return f.ca, nil }

func newCA(t *testing.T, c *secrets.Cipher) fakeCA {
	t.Helper()
	certPEM, keyPEM, err := pki.GenerateCA("Test CA")
	if err != nil {
		t.Fatal(err)
	}
	keyEnc, err := c.Encrypt(keyPEM)
	if err != nil {
		t.Fatal(err)
	}
	return fakeCA{&domain.CA{ID: "ca", Name: "Test CA", CertPEM: certPEM, KeyEnc: keyEnc}}
}

func newCipher(t *testing.T) *secrets.Cipher {
	t.Helper()
	c, err := secrets.NewCipher("test-key-pour-les-tests")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// --- tests ---

func TestEnsureServerCert_GenereEtPersiste(t *testing.T) {
	ctx := context.Background()
	c := newCipher(t)
	repo := &fakeRepo{}

	cert, err := EnsureServerCert(ctx, repo, newCA(t, c), c, []string{"localhost", "127.0.0.1"})
	if err != nil {
		t.Fatalf("émission: %v", err)
	}
	if repo.saves != 1 {
		t.Fatalf("le certificat doit être persisté une fois, saves=%d", repo.saves)
	}
	if len(cert.Certificate) == 0 {
		t.Fatal("certificat vide")
	}

	// Les SAN doivent être exploitables par un navigateur : DNS *et* IP.
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatal(err)
	}
	if err := leaf.VerifyHostname("localhost"); err != nil {
		t.Errorf("SAN localhost absent: %v", err)
	}
	found := false
	for _, ip := range leaf.IPAddresses {
		if ip.Equal(net.ParseIP("127.0.0.1")) {
			found = true
		}
	}
	if !found {
		t.Error("SAN 127.0.0.1 absent — le navigateur rejetterait https://127.0.0.1")
	}
	// Sans ServerAuth, un navigateur refuse le certificat.
	if !hasServerAuth(leaf) {
		t.Error("ExtKeyUsage ServerAuth manquant")
	}
}

// Le certificat doit être RÉUTILISÉ au redémarrage : sinon son empreinte change à chaque
// fois et l'administrateur revoit un avertissement navigateur, qu'il finit par ignorer.
func TestEnsureServerCert_ReutiliseLePersiste(t *testing.T) {
	ctx := context.Background()
	c := newCipher(t)
	ca := newCA(t, c)
	repo := &fakeRepo{}
	sans := []string{"localhost", "127.0.0.1"}

	first, err := EnsureServerCert(ctx, repo, ca, c, sans)
	if err != nil {
		t.Fatal(err)
	}
	second, err := EnsureServerCert(ctx, repo, ca, c, sans)
	if err != nil {
		t.Fatal(err)
	}
	if repo.saves != 1 {
		t.Errorf("aucune réémission attendue au 2e appel, saves=%d", repo.saves)
	}
	if string(first.Certificate[0]) != string(second.Certificate[0]) {
		t.Error("le certificat a changé entre deux démarrages — la persistance ne fonctionne pas")
	}
}

// L'ordre et la casse des SAN ne doivent pas provoquer de réémission inutile.
func TestEnsureServerCert_SansInsensibleALOrdre(t *testing.T) {
	ctx := context.Background()
	c := newCipher(t)
	ca := newCA(t, c)
	repo := &fakeRepo{}

	if _, err := EnsureServerCert(ctx, repo, ca, c, []string{"localhost", "127.0.0.1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := EnsureServerCert(ctx, repo, ca, c, []string{"127.0.0.1", "LOCALHOST"}); err != nil {
		t.Fatal(err)
	}
	if repo.saves != 1 {
		t.Errorf("mêmes SAN dans un autre ordre → aucune réémission, saves=%d", repo.saves)
	}
}

func TestEnsureServerCert_ReemetSiSansChangent(t *testing.T) {
	ctx := context.Background()
	c := newCipher(t)
	ca := newCA(t, c)
	repo := &fakeRepo{}

	if _, err := EnsureServerCert(ctx, repo, ca, c, []string{"localhost"}); err != nil {
		t.Fatal(err)
	}
	cert, err := EnsureServerCert(ctx, repo, ca, c, []string{"localhost", "vpn.example.org"})
	if err != nil {
		t.Fatal(err)
	}
	if repo.saves != 2 {
		t.Fatalf("un nouveau SAN doit déclencher une réémission, saves=%d", repo.saves)
	}
	leaf, _ := x509.ParseCertificate(cert.Certificate[0])
	if err := leaf.VerifyHostname("vpn.example.org"); err != nil {
		t.Errorf("le nouveau SAN n'est pas dans le certificat: %v", err)
	}
}

func TestEnsureServerCert_ReemetSiExpire(t *testing.T) {
	ctx := context.Background()
	c := newCipher(t)
	ca := newCA(t, c)
	// Certificat déjà persisté mais expirant dans 2 jours (< renewBefore).
	repo := &fakeRepo{cur: &domain.ServerCert{
		ID: "vieux", CertPEM: []byte("peu importe"), KeyEnc: []byte("x"),
		SANs: "localhost", NotAfter: time.Now().Add(48 * time.Hour),
	}}
	if _, err := EnsureServerCert(ctx, repo, ca, c, []string{"localhost"}); err != nil {
		t.Fatal(err)
	}
	if repo.saves != 1 {
		t.Error("un certificat proche de l'expiration doit être réémis")
	}
}

// Une base injoignable ne doit PAS être confondue avec « pas encore de certificat » :
// régénérer en silence masquerait la panne et changerait l'empreinte.
func TestEnsureServerCert_ErreurDeBaseRemonte(t *testing.T) {
	c := newCipher(t)
	_, err := EnsureServerCert(context.Background(), errRepo{err: context.DeadlineExceeded},
		newCA(t, c), c, []string{"localhost"})
	if err == nil {
		t.Fatal("une erreur de base doit remonter, pas déclencher une réémission")
	}
}

func TestDefaultSANs(t *testing.T) {
	sans := DefaultSANs()
	for _, want := range []string{"localhost", "127.0.0.1", "::1"} {
		if !contains(sans, want) {
			t.Errorf("SAN par défaut %q manquant: %v", want, sans)
		}
	}
}

func TestFromFiles_ExigeLesDeux(t *testing.T) {
	if _, err := FromFiles("cert.pem", ""); err == nil {
		t.Error("TLS_CERT sans TLS_KEY doit échouer explicitement")
	}
	if _, err := FromFiles("/inexistant.pem", "/inexistant.key"); err == nil {
		t.Error("un chemin inexistant doit échouer")
	}
}

// --- utilitaires ---

func hasServerAuth(c *x509.Certificate) bool {
	for _, u := range c.ExtKeyUsage {
		if u == x509.ExtKeyUsageServerAuth {
			return true
		}
	}
	return false
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
