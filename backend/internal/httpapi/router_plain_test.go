// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Le CDP doit rester servi EN CLAIR : c'est charon (plugin curl) qui le lit, et il
// refuserait un certificat signé par notre propre CA. Si ce test casse, la révocation
// de certificats cesse silencieusement de fonctionner sur les passerelles.
func TestPlainRouter_CRLServieEnClair(t *testing.T) {
	api := newHarness(t).api
	srv := httptest.NewServer(api.PlainRouter("7926", nil))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/crl.der")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("/crl.der doit répondre 200 en clair, obtenu %d", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); ct != "application/pkix-crl" {
		t.Errorf("content-type attendu application/pkix-crl, obtenu %q", ct)
	}
}

func TestPlainRouter_HealthzEnClair(t *testing.T) {
	api := newHarness(t).api
	srv := httptest.NewServer(api.PlainRouter("7926", nil))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("/healthz doit répondre 200 en clair, obtenu %d", res.StatusCode)
	}
}

// Tout le reste doit partir en HTTPS, en 308 : un 302 transformerait un POST en GET et
// ferait silencieusement disparaître le corps de la requête.
func TestPlainRouter_RedirigeVersHTTPS(t *testing.T) {
	api := newHarness(t).api
	h := api.PlainRouter("7926", nil)

	for _, method := range []string{http.MethodGet, http.MethodPost} {
		req := httptest.NewRequest(method, "http://vpn.example.org:7927/api/v1/tunnels?x=1", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusPermanentRedirect {
			t.Errorf("%s: attendu 308, obtenu %d", method, rec.Code)
		}
		want := "https://vpn.example.org:7926/api/v1/tunnels?x=1"
		if got := rec.Header().Get("Location"); got != want {
			t.Errorf("%s: Location attendu %q, obtenu %q", method, want, got)
		}
	}
}

// Sur le port HTTPS standard, l'URL de redirection ne doit pas porter « :443 ».
func TestPlainRouter_Redirection443SansPort(t *testing.T) {
	api := newHarness(t).api
	req := httptest.NewRequest(http.MethodGet, "http://vpn.example.org/me", nil)
	rec := httptest.NewRecorder()
	api.PlainRouter("443", nil).ServeHTTP(rec, req)

	if got, want := rec.Header().Get("Location"), "https://vpn.example.org/me"; got != want {
		t.Errorf("Location attendu %q, obtenu %q", want, got)
	}
}

// Un challenge ACME doit être servi en clair : le rediriger ferait échouer la validation
// Let's Encrypt, et donc l'obtention du certificat.
func TestPlainRouter_ChallengeACMENonRedirige(t *testing.T) {
	api := newHarness(t).api
	acme := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("jeton-de-validation"))
	})
	req := httptest.NewRequest(http.MethodGet, "http://vpn.example.org/.well-known/acme-challenge/abc", nil)
	rec := httptest.NewRecorder()
	api.PlainRouter("7926", acme).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("le challenge ACME doit être servi (200), obtenu %d", rec.Code)
	}
	if rec.Body.String() != "jeton-de-validation" {
		t.Errorf("corps du challenge inattendu: %q", rec.Body.String())
	}
}
