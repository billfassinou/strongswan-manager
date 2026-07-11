package pki

import (
	"crypto/x509"
	"net"
	"testing"
)

func TestGenerateCA(t *testing.T) {
	certPEM, keyPEM, err := GenerateCA("Test Root CA")
	if err != nil {
		t.Fatal(err)
	}
	cert, err := ParseCert(certPEM)
	if err != nil {
		t.Fatal(err)
	}
	if !cert.IsCA {
		t.Fatal("le certificat CA devrait avoir IsCA=true")
	}
	if cert.Subject.CommonName != "Test Root CA" {
		t.Fatalf("CN = %q", cert.Subject.CommonName)
	}
	if len(keyPEM) == 0 {
		t.Fatal("clé CA vide")
	}
}

func TestIssueCertSignedByCAWithSAN(t *testing.T) {
	caCert, caKey, err := GenerateCA("Acme Root CA")
	if err != nil {
		t.Fatal(err)
	}
	iss, err := IssueCert(caCert, caKey, "gw-a", []string{"172.19.0.2", "gw-a.acme.io"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if iss.Serial == "" {
		t.Fatal("série vide")
	}

	leaf, err := ParseCert(iss.CertPEM)
	if err != nil {
		t.Fatal(err)
	}
	if leaf.IsCA {
		t.Fatal("un certificat feuille ne doit pas être CA")
	}
	// SAN IP présent
	foundIP := false
	for _, ip := range leaf.IPAddresses {
		if ip.Equal(net.ParseIP("172.19.0.2")) {
			foundIP = true
		}
	}
	if !foundIP {
		t.Fatalf("SAN IP manquant: %v", leaf.IPAddresses)
	}
	if len(leaf.DNSNames) != 1 || leaf.DNSNames[0] != "gw-a.acme.io" {
		t.Fatalf("SAN DNS manquant: %v", leaf.DNSNames)
	}

	// la chaîne se vérifie contre la CA
	roots := x509.NewCertPool()
	caParsed, _ := ParseCert(caCert)
	roots.AddCert(caParsed)
	if _, err := leaf.Verify(x509.VerifyOptions{Roots: roots, KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny}}); err != nil {
		t.Fatalf("vérification de chaîne échouée: %v", err)
	}
}

func TestIssueCertRejectsBadCA(t *testing.T) {
	if _, err := IssueCert([]byte("pas un pem"), []byte("pas un pem"), "x", nil, nil); err == nil {
		t.Fatal("une CA illisible aurait dû échouer")
	}
}
