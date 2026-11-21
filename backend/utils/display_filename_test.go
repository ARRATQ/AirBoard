package utils

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestDisplayFilename(t *testing.T) {
	tests := []struct {
		name, in, want string
	}{
		{"nom simple conservé avec son extension", "Rapport annuel.pdf", "Rapport annuel.pdf"},
		{"accents conservés", "Présentation été.pptx", "Présentation été.pptx"},
		{"chemin Unix retiré", "../../etc/photo.png", "photo.png"},
		{"chemin Windows retiré", `C:\Users\moi\photo.png`, "photo.png"},
		{"caractères de contrôle retirés", "pho\x00to\n.png", "photo.png"},
		{"espaces autour retirés", "  photo.png  ", "photo.png"},
		{"vide", "   ", ""},
		{"chemin sans nom", "dossier/", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DisplayFilename(tt.in); got != tt.want {
				t.Errorf("DisplayFilename(%q) = %q, attendu %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestDisplayFilenameKeepsExtensionWhenTruncating(t *testing.T) {
	got := DisplayFilename(strings.Repeat("é", 400) + ".pdf")
	if utf8.RuneCountInString(got) != maxDisplayFilenameRunes {
		t.Fatalf("longueur = %d, attendu %d", utf8.RuneCountInString(got), maxDisplayFilenameRunes)
	}
	if !strings.HasSuffix(got, ".pdf") {
		t.Errorf("l'extension doit être conservée : %q", got)
	}
}
