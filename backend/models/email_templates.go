package models

import (
	"regexp"
	"strings"
)

// Modèles d'email par défaut, dans la direction A+ du portail (plans/refonte-airboard.md) :
// IBM Plex Sans / Mono, bordures 1 px, rayons 3-6 px, bandeau sombre façon tableau d'affichage,
// couleur de la palette choisie dans Réglages → Apparence ({{.Accent}}, {{.Ink}}).
//
// Les modèles sont enregistrés en base et modifiables par l'admin. Le HTML est en styles en ligne
// (Gmail, Outlook) ; un bloc <style> ajoute le mode sombre des clients qui le gèrent (Apple Mail).

const (
	emailSans = `'IBM Plex Sans', -apple-system, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif`
	emailMono = `'IBM Plex Mono', SFMono-Regular, Menlo, Consolas, monospace`

	// Tokens neutres de la direction A+ (clair)
	emailChrome = "#f3f4f6"
	emailLine   = "#e2e4e8"
	emailSunken = "#f5f6f7"
	emailText   = "#121518"
	emailMuted  = "#596068"
	emailBoard  = "#1b1e22"
	emailBoardT = "#f4f1e8"
)

// emailPart : morceaux de mise en page réutilisés par les modèles.
type emailPart struct {
	eyebrow   string // surtitre du bandeau (capitales mono)
	preheader string // texte d'aperçu dans la boîte de réception
	body      string
	footer    string
}

func emailLayout(p emailPart) string {
	r := strings.NewReplacer(
		"%SANS%", emailSans, "%MONO%", emailMono,
		"%CHROME%", emailChrome, "%LINE%", emailLine, "%TEXT%", emailText, "%MUTED%", emailMuted,
		"%BOARD%", emailBoard, "%BOARDT%", emailBoardT,
		"%EYEBROW%", p.eyebrow, "%PREHEADER%", p.preheader, "%BODY%", p.body, "%FOOTER%", p.footer,
	)
	return r.Replace(`<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<meta name="color-scheme" content="light dark">
<meta name="supported-color-schemes" content="light dark">
<title>{{.AppName}}</title>
<link href="https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@500&family=IBM+Plex+Sans:wght@400;600&display=swap" rel="stylesheet">
<style>
body { margin: 0; padding: 0; -webkit-text-size-adjust: 100%; }
@media (max-width: 620px) {
  .ab-pad { padding-left: 20px !important; padding-right: 20px !important; }
}
@media (prefers-color-scheme: dark) {
  .ab-page { background: #0b0c0e !important; }
  .ab-card { background: #141619 !important; border-color: #25292e !important; }
  .ab-sunken { background: #101215 !important; border-color: #25292e !important; }
  .ab-line { border-color: #25292e !important; }
  .ab-text { color: #e6e9ec !important; }
  .ab-muted { color: #959da5 !important; }
  .ab-badge { background: #f5f6f7 !important; }
}
</style>
</head>
<body class="ab-page" style="margin:0;padding:0;background:%CHROME%;">
<div style="display:none;max-height:0;overflow:hidden;opacity:0;mso-hide:all;">%PREHEADER%</div>
<table role="presentation" class="ab-page" width="100%" cellpadding="0" cellspacing="0" border="0" style="background:%CHROME%;">
<tr><td align="center" style="padding:32px 12px;">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="max-width:600px;">
<tr><td class="ab-text" style="padding:0 4px 12px;font-family:%SANS%;font-size:15px;font-weight:600;color:%TEXT%;">
<span style="display:inline-block;width:10px;height:10px;margin-inline-end:8px;border-radius:2px;background:{{.Accent}};vertical-align:middle;"></span>{{.AppName}}
</td></tr>
<tr><td class="ab-card" style="background:#ffffff;border:1px solid %LINE%;border-radius:6px;overflow:hidden;">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
<tr><td class="ab-pad" style="background:%BOARD%;border-top:3px solid {{.Accent}};padding:12px 32px;font-family:%MONO%;font-size:11px;font-weight:500;letter-spacing:0.1em;text-transform:uppercase;color:%BOARDT%;">%EYEBROW%</td></tr>
<tr><td class="ab-pad ab-text" style="padding:28px 32px 32px;font-family:%SANS%;font-size:15px;line-height:1.6;color:%TEXT%;">
%BODY%
</td></tr>
</table>
</td></tr>
<tr><td class="ab-muted" style="padding:16px 4px 0;font-family:%SANS%;font-size:12px;line-height:1.5;color:%MUTED%;">
%FOOTER%<br>
<a href="{{.PortalURL}}" class="ab-muted" style="color:%MUTED%;">{{.AppName}}</a>
</td></tr>
</table>
</td></tr>
</table>
</body>
</html>`)
}

func emailTitle(text string) string {
	return `<h1 class="ab-text" style="margin:0 0 12px;font-family:` + emailSans + `;font-size:22px;line-height:1.3;font-weight:600;color:` + emailText + `;">` + text + `</h1>`
}

func emailParagraph(text string) string {
	return `<p class="ab-text" style="margin:0 0 20px;white-space:pre-line;color:` + emailText + `;">` + text + `</p>`
}

func emailMeta(text string) string {
	return `<p class="ab-muted" style="margin:0 0 24px;font-size:13px;color:` + emailMuted + `;">` + text + `</p>`
}

// emailDetail : ligne étiquette / valeur, comme une ligne de tableau d'affichage.
func emailDetail(label, value string) string {
	return `<tr><td class="ab-line ab-muted" style="padding:10px 0;border-bottom:1px solid ` + emailLine + `;width:96px;vertical-align:top;font-family:` + emailMono + `;font-size:11px;letter-spacing:0.08em;text-transform:uppercase;color:` + emailMuted + `;">` + label + `</td>` +
		`<td class="ab-line ab-text" style="padding:10px 0;border-bottom:1px solid ` + emailLine + `;font-size:14px;color:` + emailText + `;">` + value + `</td></tr>`
}

func emailDetails(rows string) string {
	return `<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" class="ab-line" style="margin:4px 0 24px;border-top:1px solid ` + emailLine + `;">` + rows + `</table>`
}

func emailButton(href, label string) string {
	return `<table role="presentation" cellpadding="0" cellspacing="0" border="0"><tr><td style="border-radius:4px;background:{{.Ink}};">` +
		`<a href="` + href + `" style="display:inline-block;padding:11px 20px;border-radius:4px;font-family:` + emailSans + `;font-size:14px;font-weight:600;line-height:1.2;color:#ffffff;text-decoration:none;">` + label + `</a>` +
		`</td></tr></table>`
}

// emailBadge : étiquette de statut du portail (capitales mono, bordure 1 px de sa couleur).
func emailBadge(color, label string) string {
	return `<p style="margin:0 0 14px;"><span class="ab-badge" style="display:inline-block;padding:2px 7px;border:1px solid ` + color + `;border-radius:3px;font-family:` + emailMono + `;font-size:11px;font-weight:500;letter-spacing:0.08em;text-transform:uppercase;color:` + color + `;">` + label + `</span></p>`
}

func emailQuote(body string) string {
	return `<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="margin:0 0 16px;"><tr>` +
		`<td class="ab-sunken" style="padding:14px 16px;background:` + emailSunken + `;border:1px solid ` + emailLine + `;border-inline-start:3px solid {{.Accent}};border-radius:4px;">` + body + `</td></tr></table>`
}

// GetDefaultEmailTemplates retourne les modèles par défaut (direction A+).
func GetDefaultEmailTemplates() []EmailTemplate {
	return []EmailTemplate{
		{
			Type:      "news",
			Name:      "Notification Nouvel Article",
			Subject:   "{{.AppName}} - Nouvel article : {{.Title}}",
			IsEnabled: true,
			HTMLBody: emailLayout(emailPart{
				eyebrow:   "Actualité",
				preheader: "{{.Summary}}",
				body: emailTitle("{{.Title}}") +
					"{{if .Summary}}" + emailParagraph("{{.Summary}}") + "{{end}}" +
					emailMeta("Par <strong>{{.Author}}</strong> · publié le {{.PublishedAt}}") +
					emailButton("{{.Link}}", "Lire l'article"),
				footer: "Vous recevez cet email car vous êtes membre d'un groupe concerné par cet article.",
			}),
		},
		{
			Type:      "application",
			Name:      "Notification Nouvelle Ressource",
			Subject:   "{{.AppName}} - Nouvelle ressource : {{.Name}}",
			IsEnabled: true,
			HTMLBody: emailLayout(emailPart{
				eyebrow:   "Nouvelle ressource",
				preheader: "{{.Description}}",
				body: emailTitle("{{.Name}}") +
					"{{if .AppGroup}}" + emailMeta("Catégorie : <strong>{{.AppGroup}}</strong>") + "{{end}}" +
					"{{if .Description}}" + emailParagraph("{{.Description}}") + "{{end}}" +
					emailButton("{{.URL}}", "Ouvrir la ressource"),
				footer: "Vous recevez cet email car vous avez accès à cette catégorie de ressources.",
			}),
		},
		{
			Type:      "event",
			Name:      "Notification Nouvel Événement",
			Subject:   "{{.AppName}} - Nouvel événement : {{.Title}}",
			IsEnabled: true,
			HTMLBody: emailLayout(emailPart{
				eyebrow:   "Événement",
				preheader: "{{.StartDate}}{{if .Location}} · {{.Location}}{{end}}",
				body: emailTitle("{{.Title}}") +
					emailDetails(
						emailDetail("Début", "{{.StartDate}}")+
							"{{if .EndDate}}"+emailDetail("Fin", "{{.EndDate}}")+"{{end}}"+
							"{{if .Location}}"+emailDetail("Lieu", "{{.Location}}")+"{{end}}",
					) +
					"{{if .Description}}" + emailParagraph("{{.Description}}") + "{{end}}" +
					emailButton("{{.Link}}", "Voir l'événement"),
				footer: "Vous recevez cet email car vous êtes membre d'un groupe concerné par cet événement.",
			}),
		},
		{
			Type:      "announcement",
			Name:      "Notification Annonce",
			Subject:   "{{.AppName}} - {{.TypeLabel}} : {{.Title}}",
			IsEnabled: true,
			HTMLBody: emailLayout(emailPart{
				eyebrow:   "Annonce",
				preheader: "{{.Content}}",
				body: emailBadge("{{.TypeColor}}", "{{.TypeLabel}}") +
					emailTitle("{{.Title}}") +
					emailParagraph("{{.Content}}") +
					emailButton("{{.PortalURL}}", "Ouvrir {{.AppName}}"),
				footer: "Vous recevez cet email car cette annonce concerne tous les utilisateurs.",
			}),
		},
		{
			Type:      "chat_message",
			Name:      "Notification Message Chat",
			Subject:   "{{.AppName}} - Nouveau message de {{.SenderName}}",
			IsEnabled: true,
			HTMLBody: emailLayout(emailPart{
				eyebrow:   "Messagerie",
				preheader: "{{.SenderName}} : {{.MessagePreview}}",
				body: emailTitle("Nouveau message") +
					emailQuote(`<p class="ab-text" style="margin:0 0 4px;font-weight:600;color:`+emailText+`;">{{.SenderName}}</p>`+
						`<p class="ab-text" style="margin:0;color:`+emailText+`;">{{.MessagePreview}}</p>`) +
					emailMeta("Envoyé le {{.SentAt}}") +
					emailButton("{{.ChatLink}}", "Ouvrir la conversation"),
				footer: "Vous recevez cet email car vous avez reçu un message pendant votre absence.",
			}),
		},
	}
}

var emailSpaces = regexp.MustCompile(`\s+`)

func sameTemplateText(a, b string) bool {
	return emailSpaces.ReplaceAllString(strings.TrimSpace(a), " ") == emailSpaces.ReplaceAllString(strings.TrimSpace(b), " ")
}

func templateOfType(list []EmailTemplate, templateType string) (EmailTemplate, bool) {
	for _, t := range list {
		if t.Type == templateType {
			return t, true
		}
	}
	return EmailTemplate{}, false
}

// UpgradeLegacyEmailTemplate passe au nouveau design un modèle resté tel que livré avant la
// direction A+ (corps identique à l'ancien modèle par défaut, aux espaces près). Un modèle modifié
// par l'admin est laissé tel quel ; il peut le réinitialiser depuis Réglages → Email.
// Le sujet n'est remplacé que s'il n'a pas été modifié non plus.
func UpgradeLegacyEmailTemplate(t EmailTemplate) (EmailTemplate, bool) {
	legacy, okLegacy := templateOfType(legacyEmailTemplates(), t.Type)
	current, okCurrent := templateOfType(GetDefaultEmailTemplates(), t.Type)
	if !okLegacy || !okCurrent || !sameTemplateText(t.HTMLBody, legacy.HTMLBody) {
		return t, false
	}
	upgraded := t
	upgraded.HTMLBody = current.HTMLBody
	upgraded.Name = current.Name
	if sameTemplateText(t.Subject, legacy.Subject) {
		upgraded.Subject = current.Subject
	}
	return upgraded, true
}
