// Commande server : point d'entrée du backend StrongSwan Manager (tranche verticale).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"

	"strongswan-manager/internal/auth"
	"strongswan-manager/internal/config"
	"strongswan-manager/internal/domain"
	"strongswan-manager/internal/httpapi"
	"strongswan-manager/internal/metrics"
	"strongswan-manager/internal/pki"
	"strongswan-manager/internal/poller"
	"strongswan-manager/internal/secrets"
	"strongswan-manager/internal/store"
	"strongswan-manager/internal/vici"
	"strongswan-manager/internal/ws"
	"strongswan-manager/openapi"
	"strongswan-manager/web"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	if err := run(log); err != nil {
		log.Error("arrêt sur erreur fatale", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg := config.Load()

	baseCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- Base de données + migrations ---
	st, err := waitForStore(baseCtx, cfg.DatabaseURL, log)
	if err != nil {
		return err
	}
	defer st.Close()
	if err := st.Migrate(baseCtx); err != nil {
		return err
	}
	log.Info("migrations appliquées")

	if err := seedUsers(baseCtx, st, cfg.SeedAdminPassword, log); err != nil {
		return err
	}

	cipher, err := secrets.NewCipher(cfg.SecretsKey)
	if err != nil {
		return err
	}
	if err := ensureCA(baseCtx, st, cipher, log); err != nil {
		return err
	}

	// --- Registre VICI (passerelles) ---
	registry := vici.NewRegistry()
	if err := setupGateways(baseCtx, st, registry, cfg, log); err != nil {
		return err
	}

	// --- Composants transverses ---
	hub := ws.NewHub()
	mx := metrics.New()
	authMgr := auth.NewManager(cfg.JWTSecret, cfg.JWTTTL)

	// --- Poller temps réel ---
	pl := poller.New(st.Tunnels, st.Gateways, registry, hub, mx, cfg.PollInterval, log)
	go pl.Run(baseCtx)

	// --- API HTTP ---
	api := &httpapi.API{
		Users: st.Users, Gateways: st.Gateways, Tunnels: st.Tunnels, Versions: st.Versions, Audit: st.Audit,
		Secrets: st.Secrets, Cipher: cipher, CA: st.CA, Certs: st.Certs, Config: st.Config,
		Vici: registry, Auth: authMgr, Hub: hub, Metrics: mx, Log: log,
		OpenAPI: openapi.Spec, SwaggerHTML: []byte(openapi.DocsHTML), CORSOrigins: cfg.CORSOrigins,
		CRLURL: cfg.CRLURL, CRLValidity: cfg.CRLValidity,
		SPA: web.Handler(),
	}
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           api.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Arrêt gracieux
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Info("arrêt demandé, fermeture…")
		cancel()
		shCtx, shCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shCancel()
		_ = srv.Shutdown(shCtx)
	}()

	log.Info("serveur démarré", "addr", cfg.HTTPAddr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// waitForStore réessaie la connexion à PostgreSQL (le conteneur DB peut démarrer après).
func waitForStore(ctx context.Context, dsn string, log *slog.Logger) (*store.Store, error) {
	var lastErr error
	for i := 0; i < 30; i++ {
		st, err := store.New(ctx, dsn)
		if err == nil {
			return st, nil
		}
		lastErr = err
		log.Info("attente de PostgreSQL…", "tentative", i+1)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return nil, lastErr
}

// seedUsers crée les 4 comptes de démonstration si la base est vide.
func seedUsers(ctx context.Context, st *store.Store, password string, log *slog.Logger) error {
	n, err := st.Users.Count(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	roles := []struct{ identity, role string }{
		{"admin", domain.RoleAdmin},
		{"operator", domain.RoleOperator},
		{"auditor", domain.RoleAuditor},
		{"viewer", domain.RoleViewer},
	}
	for _, r := range roles {
		hash, err := auth.HashPassword(password)
		if err != nil {
			return err
		}
		u := &domain.User{ID: uuid.NewString(), Identity: r.identity, Role: r.role, Enabled: true, PassHash: hash}
		if err := st.Users.Create(ctx, u); err != nil {
			return err
		}
	}
	log.Info("comptes seedés (admin/operator/auditor/viewer)", "mot_de_passe", "cf. SEED_ADMIN_PASSWORD")
	return nil
}

// ensureCA génère l'autorité de certification interne si elle n'existe pas encore.
func ensureCA(ctx context.Context, st *store.Store, cipher *secrets.Cipher, log *slog.Logger) error {
	if _, err := st.CA.Get(ctx); err == nil {
		return nil
	}
	certPEM, keyPEM, err := pki.GenerateCA("StrongSwan Manager Root CA")
	if err != nil {
		return err
	}
	keyEnc, err := cipher.Encrypt(keyPEM)
	if err != nil {
		return err
	}
	if err := st.CA.Create(ctx, &domain.CA{ID: uuid.NewString(), Name: "StrongSwan Manager Root CA", CertPEM: certPEM, KeyEnc: keyEnc}); err != nil {
		return err
	}
	log.Info("autorité de certification interne générée")
	return nil
}

// setupGateways enregistre les passerelles et leurs adaptateurs VICI.
// Sans VICI_ENDPOINTS, une passerelle "gw-local" adossée à un adaptateur mock est
// créée pour que l'API et le poller soient pleinement exerçables sans lab.
func setupGateways(ctx context.Context, st *store.Store, reg *vici.Registry, cfg config.Config, log *slog.Logger) error {
	if len(cfg.ViciEndpoints) == 0 {
		g, err := ensureGateway(ctx, st, "gw-local", "mock")
		if err != nil {
			return err
		}
		reg.Set(g.ID, vici.NewMock())
		_ = st.Gateways.UpdateStatus(ctx, g.ID, "up", "6.0.1")
		log.Info("aucun VICI_ENDPOINTS : adaptateur mock enregistré", "gateway", g.Name)
		return nil
	}
	for _, spec := range cfg.ViciEndpoints {
		name, endpoint, ok := strings.Cut(spec, "=")
		if !ok {
			log.Warn("VICI_ENDPOINTS: format attendu nom=endpoint", "valeur", spec)
			continue
		}
		g, err := ensureGateway(ctx, st, strings.TrimSpace(name), strings.TrimSpace(endpoint))
		if err != nil {
			return err
		}
		adapter, err := vici.New(g.Endpoint)
		if err != nil {
			log.Warn("adaptateur VICI non créé", "gateway", g.Name, "err", err)
			continue
		}
		reg.Set(g.ID, adapter)
		// Détection de version (V1) — best-effort.
		if ver, err := adapter.Version(ctx); err == nil {
			_ = st.Gateways.UpdateStatus(ctx, g.ID, "up", ver)
			log.Info("passerelle VICI enrôlée", "gateway", g.Name, "version", ver)
		} else {
			_ = st.Gateways.UpdateStatus(ctx, g.ID, "unknown", "")
			log.Warn("passerelle injoignable à l'enrôlement", "gateway", g.Name, "err", err)
		}
	}
	return nil
}

func ensureGateway(ctx context.Context, st *store.Store, name, endpoint string) (*domain.Gateway, error) {
	if g, err := st.Gateways.GetByName(ctx, name); err == nil {
		return g, nil
	}
	g := &domain.Gateway{ID: uuid.NewString(), Name: name, Endpoint: endpoint, Status: "unknown"}
	if err := st.Gateways.Create(ctx, g); err != nil {
		return nil, err
	}
	return st.Gateways.GetByName(ctx, name)
}
