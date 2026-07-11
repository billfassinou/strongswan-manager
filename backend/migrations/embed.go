// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

// Package migrations embarque les scripts SQL de migration dans le binaire,
// afin qu'aucun fichier externe ne soit requis à l'exécution (utile en air-gap).
package migrations

import "embed"

// FS contient tous les fichiers *.sql de migration.
//
//go:embed *.sql
var FS embed.FS
