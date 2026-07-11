package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"strongswan-manager/internal/auth"
	"strongswan-manager/internal/domain"
	"strongswan-manager/internal/store"
	"strongswan-manager/internal/vici"
)

func (a *API) handleListTunnels(w http.ResponseWriter, r *http.Request) {
	tunnels, err := a.Tunnels.List(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "lecture des tunnels impossible")
		return
	}
	out := make([]tunnelResponse, 0, len(tunnels))
	for i := range tunnels {
		out = append(out, toTunnelResponse(&tunnels[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (a *API) handleGetTunnel(w http.ResponseWriter, r *http.Request) {
	t, err := a.Tunnels.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		a.notFoundOrInternal(w, r, err, "tunnel introuvable")
		return
	}
	writeJSON(w, http.StatusOK, toTunnelResponse(t))
}

func (a *API) handleCreateTunnel(w http.ResponseWriter, r *http.Request) {
	var req tunnelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "bad_request", "corps JSON invalide")
		return
	}
	t := req.toDomain()

	if ve := domain.ValidateTunnel(t); ve != nil {
		writeValidation(w, r, ve, "Proposition cryptographique faible détectée")
		return
	}
	// La passerelle doit exister.
	if _, err := a.Gateways.Get(r.Context(), t.GatewayID); err != nil {
		writeError(w, r, http.StatusBadRequest, "unknown_gateway", "passerelle inconnue")
		return
	}

	t.ID = uuid.NewString()
	t.Status = domain.StatusInstalling
	t.SecurityScore = domain.ScoreTunnel(t).Score
	t.ConfigVersion = 1

	if err := a.Tunnels.Create(r.Context(), t); err != nil {
		writeError(w, r, http.StatusConflict, "conflict", "un tunnel de ce nom existe déjà sur cette passerelle")
		return
	}
	a.snapshot(r.Context(), t, "création")
	p := auth.PrincipalFrom(r.Context())
	_ = a.Audit.Append(r.Context(), actorID(p), "tunnel.create", t.Name)

	// Application à chaud via VICI (si la passerelle a un adaptateur enregistré).
	if err := a.applyToVICI(r.Context(), t); err != nil {
		// cohérence : on retire l'enregistrement et on signale l'échec d'application
		_ = a.Tunnels.Delete(r.Context(), t.ID)
		writeError(w, r, http.StatusBadGateway, "vici_error", "application VICI échouée : "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toTunnelResponse(t))
}

func (a *API) handleUpdateTunnel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	existing, err := a.Tunnels.Get(r.Context(), id)
	if err != nil {
		a.notFoundOrInternal(w, r, err, "tunnel introuvable")
		return
	}
	var req tunnelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "bad_request", "corps JSON invalide")
		return
	}
	t := req.toDomain()
	t.ID = existing.ID
	t.GatewayID = existing.GatewayID // la passerelle n'est pas modifiable
	if ve := domain.ValidateTunnel(t); ve != nil {
		writeValidation(w, r, ve, "Proposition cryptographique faible détectée")
		return
	}
	t.SecurityScore = domain.ScoreTunnel(t).Score
	t.ConfigVersion = existing.ConfigVersion + 1
	t.Status = existing.Status

	if err := a.Tunnels.Update(r.Context(), t); err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "mise à jour impossible")
		return
	}
	a.snapshot(r.Context(), t, "mise à jour")
	p := auth.PrincipalFrom(r.Context())
	_ = a.Audit.Append(r.Context(), actorID(p), "tunnel.update", t.Name)

	if err := a.applyToVICI(r.Context(), t); err != nil {
		writeError(w, r, http.StatusBadGateway, "vici_error", "application VICI échouée : "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toTunnelResponse(t))
}

func (a *API) handleDeleteTunnel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	t, err := a.Tunnels.Get(r.Context(), id)
	if err != nil {
		a.notFoundOrInternal(w, r, err, "tunnel introuvable")
		return
	}
	if adapter, ok := a.Vici.Get(t.GatewayID); ok {
		_ = adapter.UnloadConn(r.Context(), vici.ConnName(t))
	}
	if t.PeerGatewayID != nil && *t.PeerGatewayID != "" {
		if adapter, ok := a.Vici.Get(*t.PeerGatewayID); ok {
			_ = adapter.UnloadConn(r.Context(), vici.ConnName(t))
		}
	}
	if err := a.Tunnels.Delete(r.Context(), id); err != nil {
		a.notFoundOrInternal(w, r, err, "tunnel introuvable")
		return
	}
	p := auth.PrincipalFrom(r.Context())
	_ = a.Audit.Append(r.Context(), actorID(p), "tunnel.delete", t.Name)
	w.WriteHeader(http.StatusNoContent)
}

// handleTunnelAction pilote une CHILD_SA via VICI (initiate/terminate/rekey).
func (a *API) handleTunnelAction(action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t, err := a.Tunnels.Get(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			a.notFoundOrInternal(w, r, err, "tunnel introuvable")
			return
		}
		adapter, ok := a.Vici.Get(t.GatewayID)
		if !ok {
			writeError(w, r, http.StatusServiceUnavailable, "no_adapter", "aucune passerelle VICI joignable")
			return
		}
		child := vici.ChildName(t)
		switch action {
		case "initiate":
			err = adapter.Initiate(r.Context(), child)
		case "terminate":
			err = adapter.Terminate(r.Context(), child)
		case "rekey":
			err = adapter.Rekey(r.Context(), child)
		}
		if err != nil {
			writeError(w, r, http.StatusBadGateway, "vici_error", err.Error())
			return
		}
		p := auth.PrincipalFrom(r.Context())
		_ = a.Audit.Append(r.Context(), actorID(p), "tunnel."+action, t.Name)
		writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted", "action": action})
	}
}

func (a *API) handleTunnelVersions(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	vers, err := a.Versions.ListByTunnel(r.Context(), id)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "lecture des versions impossible")
		return
	}
	out := make([]map[string]any, 0, len(vers))
	for _, v := range vers {
		out = append(out, map[string]any{
			"id": v.ID, "n": v.N, "author_id": v.AuthorID, "message": v.Message,
			"created_at": v.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

type rollbackRequest struct {
	Version int `json:"version"`
}

func (a *API) handleRollback(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req rollbackRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	cur, err := a.Tunnels.Get(r.Context(), id)
	if err != nil {
		a.notFoundOrInternal(w, r, err, "tunnel introuvable")
		return
	}
	target := req.Version
	if target == 0 {
		target = cur.ConfigVersion - 1 // par défaut, version précédente
	}
	v, err := a.Versions.Get(r.Context(), id, target)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "unknown_version", "version cible inexistante")
		return
	}
	var restored domain.Tunnel
	if err := json.Unmarshal(v.Snapshot, &restored); err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "instantané illisible")
		return
	}
	restored.ID = cur.ID
	restored.Status = cur.Status
	restored.ConfigVersion = cur.ConfigVersion + 1
	restored.SecurityScore = domain.ScoreTunnel(&restored).Score
	if err := a.Tunnels.Update(r.Context(), &restored); err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "restauration impossible")
		return
	}
	a.snapshot(r.Context(), &restored, "rollback vers v"+itoa(target))
	p := auth.PrincipalFrom(r.Context())
	_ = a.Audit.Append(r.Context(), actorID(p), "tunnel.rollback", restored.Name)
	if err := a.applyToVICI(r.Context(), &restored); err != nil {
		writeError(w, r, http.StatusBadGateway, "vici_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toTunnelResponse(&restored))
}

// --- helpers ---

// applyToVICI charge la connexion (et le PSK) sur la passerelle du tunnel, et — pour un
// site-à-site géré des deux côtés (PeerGatewayID) — la connexion miroir sur la passerelle
// pair, ce qui permet un établissement bout-en-bout piloté depuis la console.
func (a *API) applyToVICI(ctx context.Context, t *domain.Tunnel) error {
	if err := a.applyOne(ctx, t.GatewayID, t); err != nil {
		return err
	}
	if t.PeerGatewayID != nil && *t.PeerGatewayID != "" {
		if err := a.applyOne(ctx, *t.PeerGatewayID, mirrorTunnel(t)); err != nil {
			return err
		}
	}
	return nil
}

// applyOne applique une connexion (+ PSK) sur une passerelle donnée.
func (a *API) applyOne(ctx context.Context, gatewayID string, t *domain.Tunnel) error {
	adapter, ok := a.Vici.Get(gatewayID)
	if !ok {
		return nil // mode DB-seul (aucune passerelle enregistrée) : pas d'application
	}
	if err := adapter.LoadConn(ctx, vici.ConnName(t), vici.BuildConn(t)); err != nil {
		return err
	}
	switch {
	case t.AuthMethod == domain.AuthPSK && t.SecretRef != nil && a.Secrets != nil && a.Cipher != nil:
		s, err := a.Secrets.GetByName(ctx, *t.SecretRef)
		if errors.Is(err, store.ErrNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		plain, err := a.Cipher.Decrypt(s.EncValue)
		if err != nil {
			return err
		}
		return adapter.LoadShared(ctx, s.Name, "ike", ikeOwners(t), string(plain))

	case t.AuthMethod == domain.AuthCert && t.CertRef != nil && a.Certs != nil && a.CA != nil && a.Cipher != nil:
		// charge la CA (validation), la CRL éventuelle, puis le certificat local et sa clé
		if ca, err := a.CA.Get(ctx); err == nil {
			if err := adapter.LoadCert(ctx, string(ca.CertPEM), "ca"); err != nil {
				return err
			}
			// La révocation est vérifiée par charon via le CRL Distribution Point (CDP)
			// présent dans les certificats (fetch par le plugin curl), pas par un push VICI :
			// strongSwan n'expose pas de commande de chargement de CRL.
		}
		c, err := a.Certs.GetByName(ctx, *t.CertRef)
		if errors.Is(err, store.ErrNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := adapter.LoadCert(ctx, string(c.CertPEM), ""); err != nil {
			return err
		}
		key, err := a.Cipher.Decrypt(c.KeyEnc)
		if err != nil {
			return err
		}
		return adapter.LoadKey(ctx, string(key))
	}
	return nil
}

// mirrorTunnel renvoie la vue « côté pair » d'un tunnel : adresses et sélecteurs de
// trafic locaux/distants inversés (le PSK et les propositions restent identiques).
func mirrorTunnel(t *domain.Tunnel) *domain.Tunnel {
	m := *t
	m.LocalAddr, m.RemoteAddr = t.RemoteAddr, t.LocalAddr
	m.LocalSubnets, m.RemoteSubnets = t.RemoteSubnets, t.LocalSubnets
	// côté pair, le certificat local est celui du pair
	m.CertRef = t.PeerCertRef
	return &m
}

// ikeOwners construit la liste des identités IKE propriétaires du PSK (adresses locale
// et distante, hors valeurs dynamiques).
func ikeOwners(t *domain.Tunnel) []string {
	var out []string
	for _, a := range []string{t.LocalAddr, t.RemoteAddr} {
		if a != "" && a != "%any" {
			out = append(out, a)
		}
	}
	return out
}

func (a *API) snapshot(ctx context.Context, t *domain.Tunnel, msg string) {
	snap, _ := json.Marshal(t)
	p := auth.PrincipalFrom(ctx)
	_ = a.Versions.Create(ctx, &domain.ConfigVersion{
		ID: uuid.NewString(), TunnelID: t.ID, N: t.ConfigVersion,
		AuthorID: actorID(p), Message: msg, Snapshot: snap,
	})
}

func (a *API) notFoundOrInternal(w http.ResponseWriter, r *http.Request, err error, msg string) {
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, r, http.StatusNotFound, "not_found", msg)
		return
	}
	writeError(w, r, http.StatusInternalServerError, "internal", "erreur interne")
}

func actorID(p *auth.Principal) string {
	if p == nil {
		return ""
	}
	return p.UserID
}

func itoa(n int) string {
	if n < 0 {
		return "-" + itoa(-n)
	}
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + itoa(n%10)
}
