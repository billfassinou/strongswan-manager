// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

// Package vici abstrait la communication avec StrongSwan via l'interface VICI.
// L'interface Adapter permet deux implémentations : le vrai client govici et un
// mock en mémoire pour les tests unitaires — et facilite l'externalisation
// ultérieure vers un agent distant (mTLS) sans toucher au reste du code.
package vici

import (
	"context"

	"strongswan-manager/internal/domain"
)

// Adapter est le contrat d'accès à un démon charon (une passerelle).
type Adapter interface {
	// Version renvoie la version de StrongSwan (pour la détection de compatibilité, V1).
	Version(ctx context.Context) (string, error)
	// LoadConn charge/rafraîchit une connexion (équivalent structuré de swanctl --load-conns).
	LoadConn(ctx context.Context, name string, conn map[string]any) error
	// UnloadConn retire une connexion.
	UnloadConn(ctx context.Context, name string) error
	// LoadShared charge un secret partagé (PSK/EAP) pour des identités IKE données.
	LoadShared(ctx context.Context, id, secretType string, ikeIDs []string, data string) error
	// LoadCert charge un certificat X.509 (flag "" pour une feuille, "ca" pour une autorité).
	LoadCert(ctx context.Context, pem string, flag string) error
	// LoadKey charge une clé privée (PEM).
	LoadKey(ctx context.Context, pem string) error
	// ListSAs renvoie l'état temps réel des SA (swanctl --list-sas).
	ListSAs(ctx context.Context) ([]domain.SAState, error)
	// Initiate/Terminate/Rekey pilotent une CHILD_SA (EF-25 / EF-02).
	Initiate(ctx context.Context, child string) error
	Terminate(ctx context.Context, child string) error
	Rekey(ctx context.Context, child string) error
	// Close ferme la session.
	Close() error
}
