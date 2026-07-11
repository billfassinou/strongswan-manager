// Package migrations embarque les scripts SQL de migration dans le binaire,
// afin qu'aucun fichier externe ne soit requis à l'exécution (utile en air-gap).
package migrations

import "embed"

// FS contient tous les fichiers *.sql de migration.
//
//go:embed *.sql
var FS embed.FS
