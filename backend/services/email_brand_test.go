package services

import (
	"strings"
	"testing"

	"airboard/models"
)

// Valeurs de référence calculées par inkForLight (frontend/src/config/tokens.js).
func TestInkForLightMatchesFrontend(t *testing.T) {
	cases := map[string]string{
		"#d97757": "#c2512c",
		"#c8853a": "#9e682c",
		"#5c6ef8": "#5265f8",
		"#10b981": "#0c855d",
		"#e84f7a": "#e22258",
		"#ffcc00": "#8f7200",
		"#3366ff": "#3366ff",
		"#0a0a0a": "#0a0a0a",
		"#7fdbff": "#007cad",
	}
	for accent, want := range cases {
		if got := inkForLight(accent); got != want {
			t.Errorf("inkForLight(%s) = %s, attendu %s", accent, got, want)
		}
	}
}

func TestBrandForFollowsChosenPalette(t *testing.T) {
	cases := []struct {
		name        string
		settings    models.AppSettings
		accent, ink string
	}{
		{"palette livrée avec encre", models.AppSettings{ColorPalette: "indigo"}, "#5c6ef8", "#3d4ed4"},
		{"palette livrée, encre calculée", models.AppSettings{ColorPalette: "ambre"}, "#c8853a", "#9e682c"},
		{"couleur personnalisée", models.AppSettings{ColorPalette: "custom", CustomAccentColor: "#FFCC00"}, "#ffcc00", "#8f7200"},
		{"palette personnalisée", models.AppSettings{ColorPalette: "cp_7", CustomPalettes: `[{"id":"cp_7","accent":"#7fdbff","bg":"#ffffff","auto_bg":true}]`}, "#7fdbff", "#007cad"},
		{"palette personnalisée introuvable", models.AppSettings{ColorPalette: "cp_9", CustomPalettes: `[]`}, "#d97757", "#a8462f"},
		{"couleur personnalisée invalide", models.AppSettings{ColorPalette: "custom", CustomAccentColor: "red;}"}, "#d97757", "#a8462f"},
		{"palette inconnue", models.AppSettings{ColorPalette: "néon"}, "#d97757", "#a8462f"},
		{"aucune palette", models.AppSettings{}, "#d97757", "#a8462f"},
	}
	for _, c := range cases {
		b := brandFor(c.settings, "https://portail.example/")
		if b.Accent != c.accent || b.Ink != c.ink {
			t.Errorf("%s : %s / %s, attendu %s / %s", c.name, b.Accent, b.Ink, c.accent, c.ink)
		}
		if b.PortalURL != "https://portail.example" {
			t.Errorf("%s : adresse %q", c.name, b.PortalURL)
		}
	}
}

func TestDefaultEmailTemplatesRender(t *testing.T) {
	brand := brandFor(models.AppSettings{ColorPalette: "emeraude"}, "https://portail.example")
	label, color := announcementType("warning")
	data := map[string]interface{}{
		"news":         NewsEmailData{EmailBrand: brand, Title: "Titre <b>gras</b>", Summary: "Résumé", Author: "Jean", Link: "https://portail.example/news/x", AppName: "Intranet", PublishedAt: "23/11/2026 à 10:00"},
		"application":  ApplicationEmailData{EmailBrand: brand, Name: "Outil", Description: "Desc", URL: "https://outil.example", AppGroup: "RH", AppName: "Intranet"},
		"event":        EventEmailData{EmailBrand: brand, Title: "Réunion", Description: "Ordre du jour", StartDate: "24/11/2026 à 09:00", Location: "Salle A", Link: "https://portail.example/events/r", AppName: "Intranet"},
		"announcement": AnnouncementEmailData{EmailBrand: brand, Title: "Coupure", Content: "Ligne 1\nLigne 2", Type: "warning", TypeLabel: label, TypeColor: color, AppName: "Intranet"},
		"chat_message": ChatMessageEmailData{EmailBrand: brand, SenderName: "Salma", MessagePreview: "Bonjour", SentAt: "23/11/2026 à 11:00", ChatLink: "https://portail.example/chat", AppName: "Intranet"},
	}
	s := &EmailService{}
	rendered := map[string]string{}
	for _, tpl := range models.GetDefaultEmailTemplates() {
		d, ok := data[tpl.Type]
		if !ok {
			t.Fatalf("pas de données d'essai pour %s", tpl.Type)
		}
		html, err := s.ExecuteTemplate(tpl.HTMLBody, d)
		if err != nil {
			t.Fatalf("%s : %v", tpl.Type, err)
		}
		rendered[tpl.Type] = html
		subject, err := s.ExecuteTemplate(tpl.Subject, d)
		if err != nil || !strings.HasPrefix(subject, "Intranet - ") {
			t.Fatalf("%s : sujet %q (%v)", tpl.Type, subject, err)
		}
		for _, want := range []string{"#047857", "#10b981", "IBM Plex Sans", "https://portail.example"} {
			if !strings.Contains(html, want) {
				t.Errorf("%s : %q absent du rendu", tpl.Type, want)
			}
		}
		// Une valeur refusée par html/template laisse ce marqueur à la place
		if strings.Contains(html, "ZgotmplZ") {
			t.Errorf("%s : valeur rejetée par html/template", tpl.Type)
		}
	}

	if strings.Contains(rendered["news"], "<b>gras</b>") || !strings.Contains(rendered["news"], "&lt;b&gt;gras&lt;/b&gt;") {
		t.Error("le titre d'un article doit être échappé")
	}
	if !strings.Contains(rendered["announcement"], "Avertissement") || !strings.Contains(rendered["announcement"], "#b45309") {
		t.Error("l'annonce doit montrer le libellé et la couleur de son type")
	}
}
