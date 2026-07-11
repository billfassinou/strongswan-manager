package secrets

import (
	"bytes"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	c, err := NewCipher("passphrase-de-test")
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte("mon-psk-super-secret")
	enc, err := c.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(enc, plain) {
		t.Fatal("le texte chiffré contient le clair")
	}
	dec, err := c.Decrypt(enc)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(dec, plain) {
		t.Fatalf("round-trip: %q != %q", dec, plain)
	}
}

func TestEncryptNonDeterministic(t *testing.T) {
	c, _ := NewCipher("k")
	a, _ := c.Encrypt([]byte("x"))
	b, _ := c.Encrypt([]byte("x"))
	if bytes.Equal(a, b) {
		t.Fatal("deux chiffrés identiques (nonce non aléatoire ?)")
	}
}

func TestDecryptWrongKeyFails(t *testing.T) {
	c1, _ := NewCipher("k1")
	c2, _ := NewCipher("k2")
	enc, _ := c1.Encrypt([]byte("secret"))
	if _, err := c2.Decrypt(enc); err == nil {
		t.Fatal("déchiffrement avec une mauvaise clé aurait dû échouer")
	}
}

func TestDecryptTruncated(t *testing.T) {
	c, _ := NewCipher("k")
	if _, err := c.Decrypt([]byte("court")); err == nil {
		t.Fatal("données tronquées auraient dû échouer")
	}
}
