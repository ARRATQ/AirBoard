package utils

import (
	"path/filepath"
	"strings"
	"unicode"
)

// maxDisplayFilenameRunes borne le nom affiché d'un fichier (colonne texte, affichage, téléchargement).
const maxDisplayFilenameRunes = 255

// DisplayFilename renvoie le nom d'origine d'un fichier envoyé, pour l'affichage et le
// téléchargement : sans chemin ni caractère de contrôle, extension conservée. Le nom de
// stockage, lui, reste celui de SanitizeFilename. Renvoie "" s'il ne reste aucun nom.
func DisplayFilename(original string) string {
	name := original
	if i := strings.LastIndexAny(name, `/\`); i >= 0 {
		name = name[i+1:]
	}
	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, name)
	name = strings.TrimSpace(name)

	runes := []rune(name)
	if len(runes) <= maxDisplayFilenameRunes {
		return name
	}
	ext := []rune(filepath.Ext(name))
	if len(ext) >= maxDisplayFilenameRunes {
		ext = nil
	}
	return string(runes[:maxDisplayFilenameRunes-len(ext)]) + string(ext)
}
