package models

import (
	"strings"
	"testing"
)

func TestUpgradeLegacyEmailTemplate(t *testing.T) {
	for _, legacy := range legacyEmailTemplates() {
		current, _ := templateOfType(GetDefaultEmailTemplates(), legacy.Type)

		// Modèle livré, tel quel : passe au nouveau design
		stored := legacy
		stored.ID = 42
		stored.IsEnabled = false
		got, ok := UpgradeLegacyEmailTemplate(stored)
		if !ok || got.HTMLBody != current.HTMLBody || got.Name != current.Name || got.Subject != current.Subject {
			t.Errorf("%s : modèle livré non mis à jour", legacy.Type)
		}
		if got.ID != 42 || got.IsEnabled {
			t.Errorf("%s : l'identifiant et l'activation doivent être conservés", legacy.Type)
		}

		// Espaces et retours à la ligne différents (éditeur, base) : toujours reconnu
		reflowed := legacy
		reflowed.HTMLBody = "\n  " + strings.ReplaceAll(legacy.HTMLBody, "\n", "\r\n    ") + "\n"
		if _, ok := UpgradeLegacyEmailTemplate(reflowed); !ok {
			t.Errorf("%s : différence d'espaces non tolérée", legacy.Type)
		}

		// Sujet personnalisé : corps mis à jour, sujet conservé
		custom := legacy
		custom.Subject = "Mon sujet"
		if got, ok := UpgradeLegacyEmailTemplate(custom); !ok || got.Subject != "Mon sujet" {
			t.Errorf("%s : le sujet personnalisé doit être conservé", legacy.Type)
		}

		// Corps modifié par l'admin : on n'y touche pas
		edited := legacy
		edited.HTMLBody = legacy.HTMLBody + "<p>Signature</p>"
		if got, ok := UpgradeLegacyEmailTemplate(edited); ok || got.HTMLBody != edited.HTMLBody {
			t.Errorf("%s : un modèle personnalisé ne doit pas être remplacé", legacy.Type)
		}

		// Déjà au nouveau design : rien à faire
		if _, ok := UpgradeLegacyEmailTemplate(current); ok {
			t.Errorf("%s : modèle déjà à jour traité à nouveau", legacy.Type)
		}
	}

	if _, ok := UpgradeLegacyEmailTemplate(EmailTemplate{Type: "inconnu", HTMLBody: "x"}); ok {
		t.Error("type inconnu : aucune mise à jour attendue")
	}
}

func TestDefaultEmailTemplatesCoverEveryLegacyType(t *testing.T) {
	for _, legacy := range legacyEmailTemplates() {
		if _, ok := templateOfType(GetDefaultEmailTemplates(), legacy.Type); !ok {
			t.Errorf("aucun nouveau modèle pour %s", legacy.Type)
		}
	}
}
