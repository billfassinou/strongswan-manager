// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package httpapi

import (
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// PlainRouter est le routeur de l'écouteur EN CLAIR (HTTP_REDIRECT_ADDR).
//
// Il n'existe que pour une raison : le point de distribution de CRL doit rester en HTTP.
// charon récupère /crl.der avec son plugin curl ; servi en HTTPS derrière un certificat
// signé par notre propre CA, il le refuserait — et pour valider ce certificat il lui
// faudrait justement la CRL. La RFC 5280 tranche cette circularité en servant les CDP
// en HTTP simple. La CRL est une donnée signée et publique : la servir en clair
// n'expose rien.
//
// Tout le reste est redirigé en 308 vers HTTPS (308 et non 302 : la méthode et le corps
// doivent être préservés, sinon un POST redirigé deviendrait un GET).
//
// httpsPort est le port de l'écouteur TLS ("7926") ; vide, la redirection vise le port
// par défaut. acme, s'il est non nil, traite les challenges HTTP-01 de Let's Encrypt.
func (a *API) PlainRouter(httpsPort string, acme http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	// Le CDP, en clair et sans authentification : c'est charon qui le lit.
	r.Get("/crl.der", a.handleGetCRLDer)
	// Sonde de vie : utile aux healthchecks qui ne savent pas gérer un certificat auto-signé.
	r.Get("/healthz", a.handleHealth)

	r.NotFound(redirectToHTTPS(httpsPort, acme))
	r.MethodNotAllowed(redirectToHTTPS(httpsPort, acme))
	return r
}

func redirectToHTTPS(httpsPort string, acme http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Les challenges ACME doivent être servis EN CLAIR, sans redirection :
		// les rediriger ferait échouer la validation Let's Encrypt.
		if acme != nil && strings.HasPrefix(r.URL.Path, "/.well-known/acme-challenge/") {
			acme.ServeHTTP(w, r)
			return
		}
		host := r.Host
		if h, _, err := net.SplitHostPort(r.Host); err == nil {
			host = h
		}
		if httpsPort != "" && httpsPort != "443" {
			host = net.JoinHostPort(host, httpsPort)
		}
		u := *r.URL
		u.Scheme, u.Host = "https", host
		http.Redirect(w, r, u.String(), http.StatusPermanentRedirect)
	}
}
