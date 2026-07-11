// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

// Package tlsx fournit le certificat TLS du serveur.
//
// Trois sources, par ordre de priorité :
//  1. ACME (Let's Encrypt) si un domaine public est configuré — câblé dans cmd/server ;
//  2. un certificat fourni par l'exploitant (TLS_CERT / TLS_KEY) ;
//  3. à défaut, un certificat auto-généré, signé par la CA interne et persisté en base.
//
// Le cas 3 est ce qui permet à l'application de démarrer en HTTPS sans aucune
// configuration. Le certificat est persisté (et non régénéré à chaque démarrage) pour
// que son empreinte reste stable : sinon l'administrateur reverrait un avertissement
// navigateur après chaque redémarrage, et prendrait l'habitude de l'ignorer.
package tlsx

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"strongswan-manager/internal/domain"
	"strongswan-manager/internal/pki"
	"strongswan-manager/internal/store"
)

// Repo persiste le certificat serveur. Implémenté par store.ServerTLSRepo.
type Repo interface {
	Get(ctx context.Context) (*domain.ServerCert, error)
	Save(ctx context.Context, c *domain.ServerCert) error
}

// CAProvider donne accès à l'autorité interne. Implémenté par store.CARepo.
type CAProvider interface {
	Get(ctx context.Context) (*domain.CA, error)
}

// Cipher chiffre la clé privée au repos. Implémenté par secrets.Cipher.
type Cipher interface {
	Encrypt(plain []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

// renewBefore : on réémet le certificat quand il lui reste moins de 30 jours.
const renewBefore = 30 * 24 * time.Hour

// FromFiles charge un certificat fourni par l'exploitant.
func FromFiles(certPath, keyPath string) (tls.Certificate, error) {
	if certPath == "" || keyPath == "" {
		return tls.Certificate{}, errors.New("TLS_CERT et TLS_KEY doivent être fournis ensemble")
	}
	if _, err := os.Stat(certPath); err != nil {
		return tls.Certificate{}, fmt.Errorf("certificat TLS illisible: %w", err)
	}
	return tls.LoadX509KeyPair(certPath, keyPath)
}

// DefaultSANs renvoie les noms sous lesquels le serveur est joignable par défaut.
// Un certificat sans SAN correspondant est rejeté par les navigateurs modernes (le
// Common Name seul n'est plus honoré depuis Chrome 58).
func DefaultSANs() []string {
	sans := []string{"localhost", "127.0.0.1", "::1"}
	if h, err := os.Hostname(); err == nil && h != "" && h != "localhost" {
		sans = append(sans, h)
	}
	return sans
}

// EnsureServerCert renvoie le certificat serveur, en le réémettant si nécessaire.
//
// Il est réémis lorsqu'il est absent, expiré (ou proche de l'être), ou lorsque les SAN
// demandés ont changé — sinon le certificat persisté est réutilisé tel quel.
func EnsureServerCert(ctx context.Context, repo Repo, ca CAProvider, cipher Cipher, sans []string) (tls.Certificate, error) {
	if len(sans) == 0 {
		sans = DefaultSANs()
	}
	want := normalizeSANs(sans)

	cur, err := repo.Get(ctx)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		// Toute erreur autre qu'« absent » (base injoignable…) doit remonter : réémettre
		// en silence masquerait un vrai problème et changerait l'empreinte du certificat.
		return tls.Certificate{}, err
	}
	if err == nil && cur.SANs == want && time.Until(cur.NotAfter) > renewBefore {
		keyPEM, err := cipher.Decrypt(cur.KeyEnc)
		if err != nil {
			return tls.Certificate{}, fmt.Errorf("clé TLS illisible (SECRETS_KEY a-t-elle changé ?): %w", err)
		}
		return tls.X509KeyPair(cur.CertPEM, keyPEM)
	}

	authority, err := ca.Get(ctx)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("autorité de certification indisponible: %w", err)
	}
	caKey, err := cipher.Decrypt(authority.KeyEnc)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("clé de la CA illisible: %w", err)
	}

	// Pas de CRL Distribution Point sur le certificat du serveur lui-même : un
	// navigateur qui tenterait de le vérifier créerait une dépendance circulaire.
	issued, err := pki.IssueCert(authority.CertPEM, caKey, sans[0], sans, nil)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("émission du certificat serveur: %w", err)
	}
	keyEnc, err := cipher.Encrypt(issued.KeyPEM)
	if err != nil {
		return tls.Certificate{}, err
	}
	if err := repo.Save(ctx, &domain.ServerCert{
		ID: uuid.NewString(), CertPEM: issued.CertPEM, KeyEnc: keyEnc,
		SANs: want, NotAfter: issued.NotAfter,
	}); err != nil {
		return tls.Certificate{}, fmt.Errorf("persistance du certificat serveur: %w", err)
	}
	return tls.X509KeyPair(issued.CertPEM, issued.KeyPEM)
}

// normalizeSANs sérialise les SAN de façon stable, pour comparer l'ancien et le nouveau
// jeu sans être sensible à l'ordre ni à la casse.
func normalizeSANs(sans []string) string {
	out := make([]string, 0, len(sans))
	for _, s := range sans {
		if s = strings.ToLower(strings.TrimSpace(s)); s != "" {
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return strings.Join(out, ",")
}
