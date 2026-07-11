// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package httpapi

import (
	"context"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"strongswan-manager/internal/auth"
	"strongswan-manager/internal/domain"
	"strongswan-manager/internal/pki"
)

type certRequest struct {
	Name string   `json:"name"`
	CN   string   `json:"cn"`
	Kind string   `json:"kind"` // server | client
	SANs []string `json:"sans"`
}

func certView(c domain.Certificate) map[string]any {
	return map[string]any{
		"id": c.ID, "name": c.Name, "cn": c.CN, "kind": c.Kind, "serial": c.Serial,
		"status":     c.Status,
		"not_before": c.NotBefore.UTC().Format("2006-01-02T15:04:05Z"),
		"not_after":  c.NotAfter.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func (a *API) handleGetCA(w http.ResponseWriter, r *http.Request) {
	ca, err := a.CA.Get(r.Context())
	if err != nil {
		writeError(w, r, http.StatusNotFound, "no_ca", "aucune autorité de certification")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": ca.ID, "name": ca.Name, "cert_pem": string(ca.CertPEM)})
}

func (a *API) handleListCerts(w http.ResponseWriter, r *http.Request) {
	list, err := a.Certs.List(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "lecture des certificats impossible")
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, c := range list {
		out = append(out, certView(c))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (a *API) handleCreateCert(w http.ResponseWriter, r *http.Request) {
	var req certRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "bad_request", "corps JSON invalide")
		return
	}
	if req.Name == "" || req.CN == "" {
		writeError(w, r, http.StatusBadRequest, "bad_request", "name et cn requis")
		return
	}
	kind := domain.CertKindServer
	if req.Kind == domain.CertKindClient {
		kind = domain.CertKindClient
	}

	ca, err := a.CA.Get(r.Context())
	if err != nil {
		writeError(w, r, http.StatusFailedDependency, "no_ca", "autorité de certification indisponible")
		return
	}
	caKey, err := a.Cipher.Decrypt(ca.KeyEnc)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "clé CA illisible")
		return
	}
	var cdps []string
	if a.CRLURL != "" {
		cdps = []string{a.CRLURL}
	}
	issued, err := pki.IssueCert(ca.CertPEM, caKey, req.CN, req.SANs, cdps)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "émission impossible: "+err.Error())
		return
	}
	keyEnc, err := a.Cipher.Encrypt(issued.KeyPEM)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "chiffrement de la clé impossible")
		return
	}
	c := &domain.Certificate{
		ID: uuid.NewString(), Name: req.Name, CN: req.CN, Kind: kind, Serial: issued.Serial,
		Status: domain.CertValid, NotBefore: issued.NotAfter.AddDate(-1, 0, 0), NotAfter: issued.NotAfter,
		CertPEM: issued.CertPEM, KeyEnc: keyEnc,
	}
	if err := a.Certs.Create(r.Context(), c); err != nil {
		writeError(w, r, http.StatusConflict, "conflict", "un certificat de ce nom existe déjà")
		return
	}
	p := auth.PrincipalFrom(r.Context())
	_ = a.Audit.Append(r.Context(), actorID(p), "cert.issue", c.Name)
	writeJSON(w, http.StatusCreated, certView(*c))
}

func (a *API) handleRevokeCert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := a.Certs.Revoke(r.Context(), id); err != nil {
		a.notFoundOrInternal(w, r, err, "certificat introuvable")
		return
	}
	// régénère la CRL ; les passerelles la récupèrent via le CDP (fetch curl) à la
	// prochaine validation de certificat.
	_, _ = a.regenerateCRL(r.Context())
	p := auth.PrincipalFrom(r.Context())
	_ = a.Audit.Append(r.Context(), actorID(p), "cert.revoke", id)
	writeJSON(w, http.StatusOK, map[string]any{"status": "revoked"})
}

func (a *API) handleGetCRL(w http.ResponseWriter, r *http.Request) {
	crlPEM, err := a.regenerateCRL(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "génération CRL impossible: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/x-pem-file")
	_, _ = w.Write(crlPEM)
}

// handleGetCRLDer sert la CRL en DER, SANS authentification : c'est l'URL pointée par le
// CRL Distribution Point des certificats et récupérée par charon (plugin curl).
func (a *API) handleGetCRLDer(w http.ResponseWriter, r *http.Request) {
	crlPEM, err := a.regenerateCRL(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "génération CRL impossible")
		return
	}
	w.Header().Set("Content-Type", "application/pkix-crl")
	if block, _ := pem.Decode(crlPEM); block != nil {
		_, _ = w.Write(block.Bytes)
		return
	}
	_, _ = w.Write(crlPEM)
}

func (a *API) handlePublishCRL(w http.ResponseWriter, r *http.Request) {
	if _, err := a.regenerateCRL(r.Context()); err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal", "génération CRL impossible")
		return
	}
	ca, _ := a.CA.Get(r.Context())
	p := auth.PrincipalFrom(r.Context())
	_ = a.Audit.Append(r.Context(), actorID(p), "crl.publish", "")
	writeJSON(w, http.StatusOK, map[string]any{"crl_number": ca.CRLNumber, "distribution_point": a.CRLURL})
}

// regenerateCRL (re)génère la CRL à partir des certificats révoqués et la persiste sur la CA.
func (a *API) regenerateCRL(ctx context.Context) ([]byte, error) {
	ca, err := a.CA.Get(ctx)
	if err != nil {
		return nil, err
	}
	caKey, err := a.Cipher.Decrypt(ca.KeyEnc)
	if err != nil {
		return nil, err
	}
	revoked, err := a.Certs.ListRevoked(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]pki.Revoked, 0, len(revoked))
	for _, rc := range revoked {
		list = append(list, pki.Revoked{Serial: rc.Serial, At: rc.RevokedAt})
	}
	validity := a.CRLValidity
	if validity <= 0 {
		validity = 24 * time.Hour
	}
	crlPEM, err := pki.GenerateCRL(ca.CertPEM, caKey, list, ca.CRLNumber+1, validity)
	if err != nil {
		return nil, err
	}
	if err := a.CA.UpdateCRL(ctx, ca.ID, ca.CRLNumber+1, crlPEM); err != nil {
		return nil, err
	}
	return crlPEM, nil
}
