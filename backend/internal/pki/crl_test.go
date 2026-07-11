// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package pki

import (
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"
)

func TestGenerateCRL(t *testing.T) {
	caCert, caKey, err := GenerateCA("Acme Root CA")
	if err != nil {
		t.Fatal(err)
	}
	iss, err := IssueCert(caCert, caKey, "gw-a", []string{"10.0.0.1"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	leaf, _ := ParseCert(iss.CertPEM)

	crlPEM, err := GenerateCRL(caCert, caKey, []Revoked{{Serial: iss.Serial, At: time.Now()}}, 1, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	block, _ := pem.Decode(crlPEM)
	if block == nil || block.Type != "X509 CRL" {
		t.Fatalf("PEM CRL invalide")
	}
	crl, err := x509.ParseRevocationList(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	// signée par la CA
	caParsed, _ := ParseCert(caCert)
	if err := crl.CheckSignatureFrom(caParsed); err != nil {
		t.Fatalf("signature CRL invalide: %v", err)
	}
	// la série révoquée est présente
	found := false
	for _, e := range crl.RevokedCertificateEntries {
		if e.SerialNumber.Cmp(leaf.SerialNumber) == 0 {
			found = true
		}
	}
	if !found {
		t.Fatal("le certificat révoqué n'apparaît pas dans la CRL")
	}
}

func TestGenerateEmptyCRL(t *testing.T) {
	caCert, caKey, _ := GenerateCA("CA")
	crlPEM, err := GenerateCRL(caCert, caKey, nil, 1, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	block, _ := pem.Decode(crlPEM)
	crl, err := x509.ParseRevocationList(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if len(crl.RevokedCertificateEntries) != 0 {
		t.Fatal("CRL vide attendue")
	}
}
