// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

// Package secrets fournit le chiffrement au repos des secrets (PSK, EAP, XAuth…).
// Les valeurs ne sont jamais stockées ni renvoyées en clair (EF-05 / §9).
package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
)

// Cipher chiffre/déchiffre avec AES-256-GCM. La clé est dérivée (SHA-256) de la
// passphrase applicative (config SECRETS_KEY). En production, préférer Vault/KMS.
type Cipher struct{ gcm cipher.AEAD }

// NewCipher construit un Cipher à partir d'une passphrase.
func NewCipher(passphrase string) (*Cipher, error) {
	key := sha256.Sum256([]byte(passphrase))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Cipher{gcm: gcm}, nil
}

// Encrypt renvoie nonce||ciphertext.
func (c *Cipher) Encrypt(plain []byte) ([]byte, error) {
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return c.gcm.Seal(nonce, nonce, plain, nil), nil
}

// Decrypt inverse Encrypt.
func (c *Cipher) Decrypt(data []byte) ([]byte, error) {
	ns := c.gcm.NonceSize()
	if len(data) < ns {
		return nil, errors.New("données chiffrées tronquées")
	}
	nonce, ct := data[:ns], data[ns:]
	return c.gcm.Open(nil, nonce, ct, nil)
}
