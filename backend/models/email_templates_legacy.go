package models

// legacyEmailTemplates : modèles par défaut d'avant la direction A+ (2026-11). Ils ne servent
// qu'à reconnaître, en base, un modèle jamais modifié, pour le passer au nouveau design.
func legacyEmailTemplates() []EmailTemplate {
	return []EmailTemplate{
		{
			Type:      "news",
			Name:      "Notification Nouvel Article",
			Subject:   "{{.AppName}} - Nouvel article : {{.Title}}",
			IsEnabled: true,
			HTMLBody: `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
.container { max-width: 600px; margin: 0 auto; background: white; }
.header { background: linear-gradient(135deg, #3B82F6 0%, #2563EB 100%); color: white; padding: 30px; text-align: center; }
.header h1 { margin: 0; font-size: 24px; font-weight: 600; }
.content { padding: 30px; }
.content h2 { color: #1f2937; margin-top: 0; font-size: 22px; }
.summary { background: #f8fafc; border-left: 4px solid #3B82F6; padding: 15px; margin: 20px 0; border-radius: 0 8px 8px 0; }
.meta { color: #6b7280; font-size: 14px; margin: 15px 0; }
.meta span { margin-right: 20px; }
.button { display: inline-block; padding: 12px 24px; background: #3B82F6; color: white; text-decoration: none; border-radius: 8px; font-weight: 500; margin-top: 20px; }
.button:hover { background: #2563EB; }
.footer { background: #f8fafc; padding: 20px; text-align: center; color: #6b7280; font-size: 12px; }
</style>
</head>
<body>
<div class="container">
<div class="header">
<h1>{{.AppName}}</h1>
</div>
<div class="content">
<h2>{{.Title}}</h2>
<div class="summary">{{.Summary}}</div>
<div class="meta">
<span>Par <strong>{{.Author}}</strong></span>
<span>Publié le {{.PublishedAt}}</span>
</div>
<a href="{{.Link}}" class="button">Lire l'article</a>
</div>
<div class="footer">
<p>Vous recevez cet email car vous êtes membre d'un groupe concerné par cet article.</p>
<p>© {{.AppName}}</p>
</div>
</div>
</body>
</html>`,
		},
		{
			Type:      "application",
			Name:      "Notification Nouvelle Application",
			Subject:   "{{.AppName}} - Nouvelle application : {{.Name}}",
			IsEnabled: true,
			HTMLBody: `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
.container { max-width: 600px; margin: 0 auto; background: white; }
.header { background: linear-gradient(135deg, #10B981 0%, #059669 100%); color: white; padding: 30px; text-align: center; }
.header h1 { margin: 0; font-size: 24px; font-weight: 600; }
.content { padding: 30px; }
.content h2 { color: #1f2937; margin-top: 0; font-size: 22px; }
.app-card { background: #f0fdf4; border: 1px solid #bbf7d0; border-radius: 12px; padding: 20px; margin: 20px 0; }
.app-card h3 { color: #166534; margin: 0 0 10px 0; }
.app-card p { color: #15803d; margin: 0; }
.category { display: inline-block; background: #dcfce7; color: #166534; padding: 4px 12px; border-radius: 20px; font-size: 13px; margin-top: 10px; }
.button { display: inline-block; padding: 12px 24px; background: #10B981; color: white; text-decoration: none; border-radius: 8px; font-weight: 500; margin-top: 20px; }
.button:hover { background: #059669; }
.footer { background: #f8fafc; padding: 20px; text-align: center; color: #6b7280; font-size: 12px; }
</style>
</head>
<body>
<div class="container">
<div class="header">
<h1>{{.AppName}}</h1>
</div>
<div class="content">
<h2>Nouvelle Application Disponible</h2>
<div class="app-card">
<h3>{{.Name}}</h3>
<p>{{.Description}}</p>
<span class="category">{{.AppGroup}}</span>
</div>
<a href="{{.URL}}" class="button">Ouvrir l'application</a>
</div>
<div class="footer">
<p>Vous recevez cet email car vous avez accès à cette catégorie d'applications.</p>
<p>© {{.AppName}}</p>
</div>
</div>
</body>
</html>`,
		},
		{
			Type:      "event",
			Name:      "Notification Nouvel Événement",
			Subject:   "{{.AppName}} - Nouvel événement : {{.Title}}",
			IsEnabled: true,
			HTMLBody: `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
.container { max-width: 600px; margin: 0 auto; background: white; }
.header { background: linear-gradient(135deg, #F59E0B 0%, #D97706 100%); color: white; padding: 30px; text-align: center; }
.header h1 { margin: 0; font-size: 24px; font-weight: 600; }
.content { padding: 30px; }
.content h2 { color: #1f2937; margin-top: 0; font-size: 22px; }
.event-details { background: #fffbeb; border: 1px solid #fcd34d; border-radius: 12px; padding: 20px; margin: 20px 0; }
.event-details .detail { display: flex; align-items: center; margin: 10px 0; }
.event-details .icon { width: 24px; margin-right: 10px; color: #d97706; }
.event-details .label { color: #92400e; font-weight: 500; }
.event-details .value { color: #78350f; margin-left: 5px; }
.description { color: #4b5563; margin: 20px 0; }
.button { display: inline-block; padding: 12px 24px; background: #F59E0B; color: white; text-decoration: none; border-radius: 8px; font-weight: 500; margin-top: 20px; }
.button:hover { background: #D97706; }
.footer { background: #f8fafc; padding: 20px; text-align: center; color: #6b7280; font-size: 12px; }
</style>
</head>
<body>
<div class="container">
<div class="header">
<h1>{{.AppName}}</h1>
</div>
<div class="content">
<h2>{{.Title}}</h2>
<div class="event-details">
<div class="detail">
<span class="label">Date de début :</span>
<span class="value">{{.StartDate}}</span>
</div>
{{if .EndDate}}
<div class="detail">
<span class="label">Date de fin :</span>
<span class="value">{{.EndDate}}</span>
</div>
{{end}}
{{if .Location}}
<div class="detail">
<span class="label">Lieu :</span>
<span class="value">{{.Location}}</span>
</div>
{{end}}
</div>
<p class="description">{{.Description}}</p>
<a href="{{.Link}}" class="button">Voir les détails</a>
</div>
<div class="footer">
<p>Vous recevez cet email car vous êtes membre d'un groupe concerné par cet événement.</p>
<p>© {{.AppName}}</p>
</div>
</div>
</body>
</html>`,
		},
		{
			Type:      "announcement",
			Name:      "Notification Annonce",
			Subject:   "{{.AppName}} - {{.Type}} : {{.Title}}",
			IsEnabled: true,
			HTMLBody: `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
.container { max-width: 600px; margin: 0 auto; background: white; }
.header { background: linear-gradient(135deg, #EF4444 0%, #DC2626 100%); color: white; padding: 30px; text-align: center; }
.header h1 { margin: 0; font-size: 24px; font-weight: 600; }
.header.info { background: linear-gradient(135deg, #3B82F6 0%, #2563EB 100%); }
.header.warning { background: linear-gradient(135deg, #F59E0B 0%, #D97706 100%); }
.header.success { background: linear-gradient(135deg, #10B981 0%, #059669 100%); }
.content { padding: 30px; }
.content h2 { color: #1f2937; margin-top: 0; font-size: 22px; }
.announcement { background: #fef2f2; border-left: 4px solid #EF4444; padding: 20px; margin: 20px 0; border-radius: 0 8px 8px 0; }
.announcement.info { background: #eff6ff; border-left-color: #3B82F6; }
.announcement.warning { background: #fffbeb; border-left-color: #F59E0B; }
.announcement.success { background: #f0fdf4; border-left-color: #10B981; }
.type-badge { display: inline-block; padding: 4px 12px; border-radius: 20px; font-size: 12px; font-weight: 600; text-transform: uppercase; margin-bottom: 15px; }
.type-badge.error { background: #fee2e2; color: #991b1b; }
.type-badge.info { background: #dbeafe; color: #1e40af; }
.type-badge.warning { background: #fef3c7; color: #92400e; }
.type-badge.success { background: #dcfce7; color: #166534; }
.footer { background: #f8fafc; padding: 20px; text-align: center; color: #6b7280; font-size: 12px; }
</style>
</head>
<body>
<div class="container">
<div class="header {{.Type}}">
<h1>{{.AppName}}</h1>
</div>
<div class="content">
<span class="type-badge {{.Type}}">{{.Type}}</span>
<h2>{{.Title}}</h2>
<div class="announcement {{.Type}}">
{{.Content}}
</div>
</div>
<div class="footer">
<p>Vous recevez cet email car cette annonce concerne tous les utilisateurs.</p>
<p>© {{.AppName}}</p>
</div>
</div>
</body>
</html>`,
		},
		{
			Type:      "chat_message",
			Name:      "Notification Message Chat",
			Subject:   "{{.AppName}} - Nouveau message de {{.SenderName}}",
			IsEnabled: true,
			HTMLBody: `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
.container { max-width: 600px; margin: 0 auto; background: white; }
.header { background: linear-gradient(135deg, #8B5CF6 0%, #7C3AED 100%); color: white; padding: 30px; text-align: center; }
.header h1 { margin: 0; font-size: 24px; font-weight: 600; }
.content { padding: 30px; }
.content h2 { color: #1f2937; margin-top: 0; font-size: 22px; }
.message-card { background: #f5f3ff; border-left: 4px solid #8B5CF6; padding: 20px; margin: 20px 0; border-radius: 0 8px 8px 0; }
.message-card .sender { font-weight: 600; color: #6D28D9; margin-bottom: 8px; }
.message-card .preview { color: #4B5563; font-style: italic; }
.meta { color: #6b7280; font-size: 14px; margin: 15px 0; }
.button { display: inline-block; padding: 12px 24px; background: #8B5CF6; color: white; text-decoration: none; border-radius: 8px; font-weight: 500; margin-top: 20px; }
.button:hover { background: #7C3AED; }
.footer { background: #f8fafc; padding: 20px; text-align: center; color: #6b7280; font-size: 12px; }
</style>
</head>
<body>
<div class="container">
<div class="header">
<h1>{{.AppName}}</h1>
</div>
<div class="content">
<h2>Nouveau message</h2>
<div class="message-card">
<div class="sender">{{.SenderName}}</div>
<div class="preview">{{.MessagePreview}}</div>
</div>
<div class="meta">
<span>Envoy&eacute; le {{.SentAt}}</span>
</div>
<a href="{{.ChatLink}}" class="button">Ouvrir la conversation</a>
</div>
<div class="footer">
<p>Vous recevez cet email car vous avez re&ccedil;u un message hors ligne sur l'intranet.</p>
<p>&copy; {{.AppName}}</p>
</div>
</div>
</body>
</html>`,
		},
	}
}
