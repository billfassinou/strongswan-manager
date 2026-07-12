// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"strongswan-manager/internal/auth"
	"strongswan-manager/internal/domain"
	"strongswan-manager/internal/metrics"
	"strongswan-manager/internal/pki"
	"strongswan-manager/internal/secrets"
	"strongswan-manager/internal/vici"
	"strongswan-manager/internal/ws"
)

// harness regroupe l'API sous test et les fakes/mocks pour les assertions.
type harness struct {
	api     *API
	mock    *vici.Mock
	mock2   *vici.Mock
	audit   *fakeAudit
	tuns    *fakeTunnels
	secrets *fakeSecrets
	authM   *auth.Manager
	router  http.Handler
}

const gwID = "gw1"
const gw2ID = "gw2"

func newHarness(t *testing.T) *harness {
	t.Helper()
	adminHash, _ := auth.HashPassword("secret")
	disabledHash, _ := auth.HashPassword("secret")

	users := &fakeUsers{byIdentity: map[string]*domain.User{
		"admin":   {ID: "u-admin", Identity: "admin", Role: domain.RoleAdmin, Enabled: true, PassHash: adminHash},
		"blocked": {ID: "u-blk", Identity: "blocked", Role: domain.RoleViewer, Enabled: false, PassHash: disabledHash},
	}}
	gws := &fakeGateways{m: map[string]*domain.Gateway{
		gwID:  {ID: gwID, Name: "gw-a", Endpoint: "mock", Version: "6.0.1", Status: "up"},
		gw2ID: {ID: gw2ID, Name: "gw-b", Endpoint: "mock", Version: "6.0.1", Status: "up"},
	}}
	tuns := &fakeTunnels{m: map[string]*domain.Tunnel{}}
	vers := &fakeVersions{byTunnel: map[string][]domain.ConfigVersion{}}
	aud := &fakeAudit{}
	sec := newFakeSecrets()
	ciph, _ := secrets.NewCipher("test-key")

	// CA interne de test (clé chiffrée avec le cipher de test)
	caCertPEM, caKeyPEM, _ := pki.GenerateCA("Test Root CA")
	caKeyEnc, _ := ciph.Encrypt(caKeyPEM)
	ca := &fakeCA{ca: &domain.CA{ID: "ca1", Name: "Test Root CA", CertPEM: caCertPEM, KeyEnc: caKeyEnc}}
	certs := newFakeCerts()

	mock := vici.NewMock()
	mock2 := vici.NewMock()
	reg := vici.NewRegistry()
	reg.Set(gwID, mock)
	reg.Set(gw2ID, mock2)

	authM := auth.NewManager("test-secret", time.Hour)
	api := &API{
		Users: users, Gateways: gws, Tunnels: tuns, Versions: vers, Audit: aud, Secrets: sec, Cipher: ciph,
		CA: ca, Certs: certs, Config: newFakeConfig(),
		Vici: reg, Auth: authM, Hub: ws.NewHub(), Metrics: metrics.New(),
		Log: nil, OpenAPI: []byte("openapi: 3.0.3"), SwaggerHTML: []byte("<html></html>"),
		CORSOrigins: []string{"*"},
	}
	return &harness{api: api, mock: mock, mock2: mock2, audit: aud, tuns: tuns, secrets: sec, authM: authM, router: api.Router()}
}

// token émet un JWT pour un rôle donné.
func (h *harness) token(role string) string {
	tok, _, _ := h.authM.Issue(&domain.User{ID: "u-" + role, Identity: role, Role: role, Enabled: true})
	return tok
}

// do exécute une requête et renvoie la réponse.
func (h *harness) do(method, path, token string, body any) *httptest.ResponseRecorder {
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, r)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)
	return w
}

func decode(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("réponse JSON invalide: %v (corps=%s)", err, w.Body.String())
	}
	return m
}

// validTunnel renvoie un corps de requête de tunnel valide.
func validTunnel(name string) map[string]any {
	return map[string]any{
		"name": name, "gateway_id": gwID, "type": "site-to-site", "ike_version": 2,
		"local":     map[string]any{"addr": "203.0.113.10", "subnets": []string{"10.1.0.0/16"}},
		"remote":    map[string]any{"addr": "198.51.100.20", "subnets": []string{"10.2.0.0/16"}},
		"auth":      map[string]any{"method": "psk"},
		"proposals": map[string]any{"ike": []string{"aes256-sha256-modp2048"}, "esp": []string{"aes256gcm16"}},
		"pfs":       true,
	}
}

// createTunnel crée un tunnel et renvoie son id.
func (h *harness) createTunnel(t *testing.T, name string) string {
	t.Helper()
	w := h.do(http.MethodPost, "/api/v1/tunnels", h.token(domain.RoleAdmin), validTunnel(name))
	if w.Code != http.StatusCreated {
		t.Fatalf("création %s: code %d (corps=%s)", name, w.Code, w.Body.String())
	}
	return decode(t, w)["id"].(string)
}

// --- Tests ---

func TestHealthz(t *testing.T) {
	h := newHarness(t)
	w := h.do(http.MethodGet, "/healthz", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("healthz code %d", w.Code)
	}
}

func TestLogin(t *testing.T) {
	h := newHarness(t)

	// succès
	w := h.do(http.MethodPost, "/api/v1/auth/login", "", map[string]string{"identity": "admin", "password": "secret"})
	if w.Code != http.StatusOK {
		t.Fatalf("login ok: code %d", w.Code)
	}
	if decode(t, w)["role"] != "admin" {
		t.Fatalf("rôle inattendu")
	}
	// mauvais mot de passe
	if w := h.do(http.MethodPost, "/api/v1/auth/login", "", map[string]string{"identity": "admin", "password": "x"}); w.Code != http.StatusUnauthorized {
		t.Fatalf("mauvais mdp: code %d (attendu 401)", w.Code)
	}
	// inconnu
	if w := h.do(http.MethodPost, "/api/v1/auth/login", "", map[string]string{"identity": "ghost", "password": "x"}); w.Code != http.StatusUnauthorized {
		t.Fatalf("inconnu: code %d (attendu 401)", w.Code)
	}
	// compte désactivé
	if w := h.do(http.MethodPost, "/api/v1/auth/login", "", map[string]string{"identity": "blocked", "password": "secret"}); w.Code != http.StatusForbidden {
		t.Fatalf("désactivé: code %d (attendu 403)", w.Code)
	}
}

func TestProtectedRequiresAuth(t *testing.T) {
	h := newHarness(t)
	if w := h.do(http.MethodGet, "/api/v1/tunnels", "", nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("sans token: code %d (attendu 401)", w.Code)
	}
}

func TestMe(t *testing.T) {
	h := newHarness(t)
	w := h.do(http.MethodGet, "/api/v1/me", h.token(domain.RoleOperator), nil)
	m := decode(t, w)
	if m["role"] != "operator" || m["can_write"] != true {
		t.Fatalf("/me inattendu: %v", m)
	}
}

func TestCreateTunnel(t *testing.T) {
	h := newHarness(t)
	w := h.do(http.MethodPost, "/api/v1/tunnels", h.token(domain.RoleAdmin), validTunnel("paris-dakar"))
	if w.Code != http.StatusCreated {
		t.Fatalf("code %d (attendu 201), corps=%s", w.Code, w.Body.String())
	}
	m := decode(t, w)
	if int(m["security_score"].(float64)) != 94 {
		t.Fatalf("security_score = %v (attendu 94)", m["security_score"])
	}
	if int(m["config_version"].(float64)) != 1 {
		t.Fatalf("config_version = %v (attendu 1)", m["config_version"])
	}
	// La connexion a bien été chargée dans VICI (mock).
	if !h.mock.HasConn("paris-dakar") {
		t.Fatal("la connexion n'a pas été chargée via VICI")
	}
	// Une entrée d'audit a été écrite.
	if len(h.audit.entries) == 0 || h.audit.entries[len(h.audit.entries)-1].Action != "tunnel.create" {
		t.Fatalf("audit tunnel.create manquant: %+v", h.audit.entries)
	}
}

func TestCreateTunnelWeakDHReturns422(t *testing.T) {
	h := newHarness(t)
	body := validTunnel("weak")
	body["proposals"] = map[string]any{"ike": []string{"aes256-sha256-modp1024"}, "esp": []string{"aes256gcm16"}}
	w := h.do(http.MethodPost, "/api/v1/tunnels", h.token(domain.RoleAdmin), body)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("code %d (attendu 422)", w.Code)
	}
	m := decode(t, w)
	if m["error"] != "validation_failed" {
		t.Fatalf("error = %v", m["error"])
	}
	if m["correlation_id"] == nil || m["correlation_id"] == "" {
		t.Fatal("correlation_id manquant (format §10)")
	}
	if _, ok := m["details"].([]any); !ok {
		t.Fatalf("details manquants: %v", m["details"])
	}
}

func TestCreateTunnelUnknownGateway(t *testing.T) {
	h := newHarness(t)
	body := validTunnel("x")
	body["gateway_id"] = "does-not-exist"
	w := h.do(http.MethodPost, "/api/v1/tunnels", h.token(domain.RoleAdmin), body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code %d (attendu 400)", w.Code)
	}
}

func TestRBACViewerCannotWrite(t *testing.T) {
	h := newHarness(t)
	w := h.do(http.MethodPost, "/api/v1/tunnels", h.token(domain.RoleViewer), validTunnel("x"))
	if w.Code != http.StatusForbidden {
		t.Fatalf("viewer POST: code %d (attendu 403)", w.Code)
	}
	// auditeur non plus
	if w := h.do(http.MethodPost, "/api/v1/tunnels", h.token(domain.RoleAuditor), validTunnel("x")); w.Code != http.StatusForbidden {
		t.Fatalf("auditor POST: code %d (attendu 403)", w.Code)
	}
	// mais l'auditeur peut lire
	if w := h.do(http.MethodGet, "/api/v1/tunnels", h.token(domain.RoleAuditor), nil); w.Code != http.StatusOK {
		t.Fatalf("auditor GET: code %d (attendu 200)", w.Code)
	}
}

func TestGetTunnelNotFound(t *testing.T) {
	h := newHarness(t)
	if w := h.do(http.MethodGet, "/api/v1/tunnels/nope", h.token(domain.RoleAdmin), nil); w.Code != http.StatusNotFound {
		t.Fatalf("code %d (attendu 404)", w.Code)
	}
}

func TestUpdateIncrementsVersion(t *testing.T) {
	h := newHarness(t)
	id := h.createTunnel(t, "paris-dakar")
	body := validTunnel("paris-dakar")
	body["proposals"] = map[string]any{"ike": []string{"aes256gcm16-sha384-ecp384-mlkem768"}, "esp": []string{"aes256gcm16"}}
	w := h.do(http.MethodPut, "/api/v1/tunnels/"+id, h.token(domain.RoleAdmin), body)
	if w.Code != http.StatusOK {
		t.Fatalf("update: code %d", w.Code)
	}
	m := decode(t, w)
	if int(m["config_version"].(float64)) != 2 {
		t.Fatalf("config_version = %v (attendu 2)", m["config_version"])
	}
	if int(m["security_score"].(float64)) != 100 {
		t.Fatalf("security_score = %v (attendu 100 après passage ML-KEM)", m["security_score"])
	}
	// deux versions enregistrées
	vw := h.do(http.MethodGet, "/api/v1/tunnels/"+id+"/versions", h.token(domain.RoleAdmin), nil)
	items := decode(t, vw)["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("%d versions (attendu 2)", len(items))
	}
}

func TestRollback(t *testing.T) {
	h := newHarness(t)
	id := h.createTunnel(t, "paris-dakar") // v1: modp2048, score 94
	body := validTunnel("paris-dakar")
	body["proposals"] = map[string]any{"ike": []string{"aes256gcm16-sha384-ecp384-mlkem768"}, "esp": []string{"aes256gcm16"}}
	h.do(http.MethodPut, "/api/v1/tunnels/"+id, h.token(domain.RoleAdmin), body) // v2: score 100

	w := h.do(http.MethodPost, "/api/v1/tunnels/"+id+"/rollback", h.token(domain.RoleAdmin), map[string]int{"version": 1})
	if w.Code != http.StatusOK {
		t.Fatalf("rollback: code %d", w.Code)
	}
	m := decode(t, w)
	if int(m["security_score"].(float64)) != 94 {
		t.Fatalf("après rollback v1, score = %v (attendu 94)", m["security_score"])
	}
	if int(m["config_version"].(float64)) != 3 {
		t.Fatalf("config_version = %v (attendu 3)", m["config_version"])
	}
}

func TestDeleteUnloadsVICI(t *testing.T) {
	h := newHarness(t)
	id := h.createTunnel(t, "paris-dakar")
	if !h.mock.HasConn("paris-dakar") {
		t.Fatal("précondition: conn chargée")
	}
	w := h.do(http.MethodDelete, "/api/v1/tunnels/"+id, h.token(domain.RoleAdmin), nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete: code %d (attendu 204)", w.Code)
	}
	if h.mock.HasConn("paris-dakar") {
		t.Fatal("la connexion aurait dû être déchargée (unload) via VICI")
	}
}

func TestInitiateAction(t *testing.T) {
	h := newHarness(t)
	id := h.createTunnel(t, "paris-dakar")
	w := h.do(http.MethodPost, "/api/v1/tunnels/"+id+"/initiate", h.token(domain.RoleAdmin), nil)
	if w.Code != http.StatusAccepted {
		t.Fatalf("initiate: code %d (attendu 202)", w.Code)
	}
}

func TestAuditList(t *testing.T) {
	h := newHarness(t)
	h.createTunnel(t, "t1")
	h.createTunnel(t, "t2")
	w := h.do(http.MethodGet, "/api/v1/audit?limit=10", h.token(domain.RoleAuditor), nil)
	items := decode(t, w)["items"].([]any)
	if len(items) < 2 {
		t.Fatalf("%d entrées d'audit (attendu >= 2)", len(items))
	}
}

func TestMetricsEndpoint(t *testing.T) {
	h := newHarness(t)
	h.api.Metrics.SetTunnel("paris-dakar", gwID, "up")
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "strongswan_tunnel_status") {
		t.Fatalf("/metrics n'expose pas la gauge (code %d)", w.Code)
	}
}

func TestOpenAPIServed(t *testing.T) {
	h := newHarness(t)
	w := h.do(http.MethodGet, "/api/v1/openapi.yaml", "", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "openapi") {
		t.Fatalf("openapi non servi (code %d)", w.Code)
	}
}

// --- Mot de passe initial imposé (comptes seedés) ---

// seeded ajoute un compte portant encore le mot de passe d'installation.
func (h *harness) seeded(t *testing.T, identity, password string) string {
	t.Helper()
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	u := &domain.User{
		ID: "u-" + identity, Identity: identity, Role: domain.RoleAdmin, Enabled: true,
		PassHash: hash, MustChangePassword: true,
	}
	h.api.Users.(*fakeUsers).byIdentity[identity] = u
	tok, _, err := h.authM.Issue(u)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func TestLoginSignalsPasswordChangeRequired(t *testing.T) {
	h := newHarness(t)
	h.seeded(t, "fresh", "motdepasseinstall")

	rec := h.do(http.MethodPost, "/api/v1/auth/login", "",
		map[string]string{"identity": "fresh", "password": "motdepasseinstall"})
	if rec.Code != http.StatusOK {
		t.Fatalf("login = %d", rec.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["must_change_password"] != true {
		t.Fatalf("must_change_password absent de la réponse de login : %v", body)
	}
}

func TestSeededAccountIsLockedOutOfTheAPI(t *testing.T) {
	h := newHarness(t)
	tok := h.seeded(t, "fresh", "motdepasseinstall")

	// Le compte voit /me (pour savoir qu'il doit changer son mot de passe)…
	if rec := h.do(http.MethodGet, "/api/v1/me", tok, nil); rec.Code != http.StatusOK {
		t.Fatalf("/me = %d, attendu 200", rec.Code)
	}
	// …mais rien d'autre, y compris en lecture.
	for _, path := range []string{"/api/v1/tunnels", "/api/v1/gateways", "/api/v1/secrets"} {
		rec := h.do(http.MethodGet, path, tok, nil)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("%s = %d, attendu 403 tant que le mot de passe initial n'est pas changé", path, rec.Code)
		}
	}
}

func TestChangePasswordUnlocksTheAPI(t *testing.T) {
	h := newHarness(t)
	tok := h.seeded(t, "fresh", "motdepasseinstall")

	// Mot de passe actuel faux.
	rec := h.do(http.MethodPost, "/api/v1/me/password", tok,
		map[string]string{"current_password": "faux", "new_password": "unNouveauMotDePasse"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("mot de passe actuel erroné = %d, attendu 401", rec.Code)
	}

	// Trop court → 422.
	rec = h.do(http.MethodPost, "/api/v1/me/password", tok,
		map[string]string{"current_password": "motdepasseinstall", "new_password": "court"})
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("mot de passe trop court = %d, attendu 422", rec.Code)
	}

	// Identique à l'ancien → 422.
	rec = h.do(http.MethodPost, "/api/v1/me/password", tok,
		map[string]string{"current_password": "motdepasseinstall", "new_password": "motdepasseinstall"})
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("mot de passe inchangé = %d, attendu 422", rec.Code)
	}

	// Changement valide → nouveau jeton, sans le drapeau.
	rec = h.do(http.MethodPost, "/api/v1/me/password", tok,
		map[string]string{"current_password": "motdepasseinstall", "new_password": "unNouveauMotDePasse"})
	if rec.Code != http.StatusOK {
		t.Fatalf("changement de mot de passe = %d (%s)", rec.Code, rec.Body.String())
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	newTok, _ := body["token"].(string)
	if newTok == "" || body["must_change_password"] != false {
		t.Fatalf("réponse inattendue : %v", body)
	}

	// L'ancien jeton reste bloqué, le nouveau ouvre l'API.
	if rec := h.do(http.MethodGet, "/api/v1/tunnels", tok, nil); rec.Code != http.StatusForbidden {
		t.Fatalf("ancien jeton = %d, attendu 403", rec.Code)
	}
	if rec := h.do(http.MethodGet, "/api/v1/tunnels", newTok, nil); rec.Code != http.StatusOK {
		t.Fatalf("nouveau jeton = %d, attendu 200", rec.Code)
	}

	// Et l'ancien mot de passe ne fonctionne plus.
	rec = h.do(http.MethodPost, "/api/v1/auth/login", "",
		map[string]string{"identity": "fresh", "password": "motdepasseinstall"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("login avec l'ancien mot de passe = %d, attendu 401", rec.Code)
	}
}
