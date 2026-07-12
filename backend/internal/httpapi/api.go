// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

// Package httpapi expose l'API REST + WebSocket (contrat §10) et le routage.
package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"strongswan-manager/internal/auth"
	"strongswan-manager/internal/metrics"
	"strongswan-manager/internal/vici"
	"strongswan-manager/internal/ws"
)

// API porte les dépendances des handlers (repos sous forme d'interfaces → testable).
type API struct {
	Users       usersStore
	Gateways    gatewaysStore
	Tunnels     tunnelsStore
	Versions    versionsStore
	Audit       auditStore
	Secrets     secretsStore
	Cipher      cipher
	CA          caStore
	Certs       certsStore
	Config      configStore
	Vici        *vici.Registry
	Auth        *auth.Manager
	Hub         *ws.Hub
	Metrics     *metrics.Metrics
	Log         *slog.Logger
	OpenAPI     []byte
	SwaggerHTML []byte
	CORSOrigins []string
	CRLURL      string        // URL publique du CDP (pointée par les certificats)
	CRLValidity time.Duration // fenêtre nextUpdate des CRL générées
	SPA         http.Handler  // sert le front React embarqué (optionnel)
}

// Router construit le routeur HTTP complet.
func (a *API) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(a.cors)

	// Endpoints publics
	r.Get("/healthz", a.handleHealth)
	r.Handle("/metrics", a.Metrics.Handler())
	// CRL en DER, non authentifiée : c'est le point de distribution (CDP) que les
	// passerelles récupèrent (plugin curl) pour valider la révocation.
	r.Get("/crl.der", a.handleGetCRLDer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/login", a.handleLogin)
		r.Get("/openapi.yaml", a.handleOpenAPI)
		r.Get("/docs", a.handleDocs)
		r.Get("/ws", a.handleWS) // auth par ?token= (le navigateur ne peut pas poser d'en-tête sur un WS)

		// Endpoints protégés
		r.Group(func(r chi.Router) {
			r.Use(a.Auth.Middleware)
			// Tant que le compte porte le mot de passe posé à l'installation, seuls /me et
			// /me/password répondent : l'API ne s'ouvre qu'une fois le mot de passe changé.
			r.Use(a.requirePasswordChanged)
			r.Get("/me", a.handleMe)
			r.Post("/me/password", a.handleChangePassword)
			r.Get("/gateways", a.handleListGateways)
			r.Get("/secrets", a.handleListSecrets)
			r.Get("/certificates", a.handleListCerts)
			r.Get("/ca", a.handleGetCA)
			r.Get("/crl", a.handleGetCRL)
			r.Get("/config/{kind}", a.handleListConfig)
			r.Get("/tunnels", a.handleListTunnels)
			r.Get("/tunnels/{id}", a.handleGetTunnel)
			r.Get("/tunnels/{id}/versions", a.handleTunnelVersions)
			r.Get("/audit", a.handleAudit)

			// Actions modifiantes : réservées aux rôles avec droit d'écriture
			r.Group(func(r chi.Router) {
				r.Use(auth.RequireWrite)
				r.Post("/secrets", a.handleCreateSecret)
				r.Delete("/secrets/{id}", a.handleDeleteSecret)
				r.Post("/certificates", a.handleCreateCert)
				r.Post("/certificates/{id}/revoke", a.handleRevokeCert)
				r.Post("/crl/publish", a.handlePublishCRL)
				r.Post("/config/{kind}", a.handleCreateConfig)
				r.Put("/config/{kind}/{id}", a.handleUpdateConfig)
				r.Delete("/config/{kind}/{id}", a.handleDeleteConfig)
				r.Post("/tunnels", a.handleCreateTunnel)
				r.Put("/tunnels/{id}", a.handleUpdateTunnel)
				r.Delete("/tunnels/{id}", a.handleDeleteTunnel)
				r.Post("/tunnels/{id}/initiate", a.handleTunnelAction("initiate"))
				r.Post("/tunnels/{id}/terminate", a.handleTunnelAction("terminate"))
				r.Post("/tunnels/{id}/rekey", a.handleTunnelAction("rekey"))
				r.Post("/tunnels/{id}/rollback", a.handleRollback)
			})
		})
	})

	// Front React embarqué (sert /, /assets/*, et repli SPA). Doit être enregistré après
	// les routes API pour ne pas les masquer.
	if a.SPA != nil {
		r.Handle("/*", a.SPA)
	}

	return r
}

func (a *API) cors(next http.Handler) http.Handler {
	allowed := "*"
	if len(a.CORSOrigins) > 0 {
		allowed = a.CORSOrigins[0]
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowed)
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	_, _ = w.Write(a.OpenAPI)
}

func (a *API) handleDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(a.SwaggerHTML)
}

func (a *API) handleWS(w http.ResponseWriter, r *http.Request) {
	if tok := r.URL.Query().Get("token"); tok != "" {
		if _, err := a.Auth.Parse(tok); err != nil {
			writeError(w, r, http.StatusUnauthorized, "unauthorized", "jeton invalide")
			return
		}
	}
	a.Hub.Handler(w, r)
}
