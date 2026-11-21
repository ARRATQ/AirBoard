// Package migrations embarque les migrations SQL versionnées du schéma, appliquées
// par goose au démarrage (voir database.Migrate et backend/MIGRATIONS.md).
package migrations

import "embed"

// FS contient les fichiers NNNNN_description.sql, appliqués dans l'ordre des numéros.
//
//go:embed *.sql
var FS embed.FS
