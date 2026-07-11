// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package pki

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

// Revoked décrit un certificat révoqué (numéro de série hexadécimal + date).
type Revoked struct {
	Serial string
	At     time.Time
}

// GenerateCRL produit une liste de révocation (CRL) signée par la CA, au format PEM.
// number est un compteur monotone (CRL Number) ; validity fixe la fenêtre nextUpdate
// (les passerelles re-récupèrent la CRL via CDP à son expiration).
func GenerateCRL(caCertPEM, caKeyPEM []byte, revoked []Revoked, number int64, validity time.Duration) ([]byte, error) {
	caCert, caKey, err := parseCA(caCertPEM, caKeyPEM)
	if err != nil {
		return nil, err
	}
	entries := make([]x509.RevocationListEntry, 0, len(revoked))
	for _, r := range revoked {
		serial, ok := new(big.Int).SetString(r.Serial, 16)
		if !ok {
			return nil, fmt.Errorf("série invalide: %s", r.Serial)
		}
		at := r.At
		if at.IsZero() {
			at = time.Now()
		}
		entries = append(entries, x509.RevocationListEntry{SerialNumber: serial, RevocationTime: at.UTC()})
	}
	if validity <= 0 {
		validity = 7 * 24 * time.Hour
	}
	now := time.Now().Add(-5 * time.Minute)
	tmpl := &x509.RevocationList{
		Number:                    big.NewInt(number),
		ThisUpdate:                now,
		NextUpdate:                now.Add(validity),
		RevokedCertificateEntries: entries,
	}
	der, err := x509.CreateRevocationList(rand.Reader, tmpl, caCert, caKey)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "X509 CRL", Bytes: der}), nil
}
