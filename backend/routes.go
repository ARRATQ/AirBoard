package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"airboard/config"
	"airboard/handlers"
	"airboard/middleware"
	"airboard/services"
	"airboard/services/chat"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// newRouter construit le routeur Gin : middlewares globaux, handlers et routes.
func newRouter(cfg *config.Config, db *gorm.DB, storageService services.StorageService,
	gamificationService *services.GamificationService, chatHub *chat.Hub) *gin.Engine {
	// Initialisation des middlewares
	authMiddleware := middleware.NewAuthMiddleware(cfg, db)
	ssoMiddleware := middleware.NewSSOMiddleware(db, cfg)
	csrfManager := middleware.NewCSRFManager()

	mediaHandler := handlers.NewMediaHandler(db, storageService)

	// Initialisation des handlers
	authHandler := handlers.NewAuthHandler(db, authMiddleware, cfg.Server.SignupEnabled, cfg, gamificationService)
	dashboardHandler := handlers.NewDashboardHandler(db)
	adminHandler := handlers.NewAdminHandler(db, cfg, gamificationService)
	groupAdminHandler := handlers.NewGroupAdminHandler(db)
	settingsHandler := handlers.NewSettingsHandler(db)
	oauthHandler := handlers.NewOAuthHandler(db, authMiddleware, gamificationService)
	favoritesHandler := handlers.NewFavoritesHandler(db)
	analyticsHandler := handlers.NewAnalyticsHandler(db, gamificationService)
	reportsHandler := handlers.NewReportsHandler(db)
	announcementHandler := handlers.NewAnnouncementHandler(db)
	newsHandler := handlers.NewNewsHandler(db, cfg, gamificationService)
	eventsHandler := handlers.NewEventsHandler(db, gamificationService)
	homeHandler := handlers.NewHomeHandler(db)
	directoryHandler := handlers.NewDirectoryHandler(db)
	versionHandler := handlers.NewVersionHandler()
	emailHandler := handlers.NewEmailHandler(db, cfg)
	commentHandler := handlers.NewCommentHandler(db, gamificationService)
	feedbackHandler := handlers.NewFeedbackHandler(db)
	notificationHandler := handlers.NewNotificationHandler(db)
	pollsHandler := handlers.NewPollsHandler(db, gamificationService)
	suggestionsHandler := handlers.NewSuggestionsHandler(db, gamificationService)
	gamificationHandler := handlers.NewGamificationHandler(db, gamificationService)
	searchHandler := handlers.NewSearchHandler(db)
	chatbotHandler := handlers.NewChatbotHandler(db, gamificationService)
	chatHandler := handlers.NewChatHandler(db, chatHub, gamificationService)

	// Créer un routeur personnalisé avec configuration sécurisée
	router := gin.New()

	// Proxies de confiance : seuls eux peuvent fixer l'IP client via X-Forwarded-For.
	// En prod (Coolify), la chaîne est Traefik -> nginx -> backend, tous sur le réseau
	// Docker privé : il faut les faire tous confiance pour retrouver l'IP réelle, sinon
	// tous les utilisateurs partageraient l'IP de Traefik (et donc la même limite).
	// TRUSTED_PROXIES permet de restreindre la liste (CIDR ou IP, séparés par des virgules).
	trustedProxies := trustedProxiesFromEnv(cfg.Server.Mode)
	if err := router.SetTrustedProxies(trustedProxies); err != nil {
		log.Fatalf("TRUSTED_PROXIES invalide: %v", err)
	}
	log.Printf("Proxies de confiance: %v", trustedProxies)

	// Middleware de logging sécurisé avec vraie IP
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] \"%s %s %s\" %d %v %s %s %s\n",
			param.TimeStamp.Format("2006/01/02 15:04:05"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.ClientIP, // Utilise l'IP réelle après SetTrustedProxies
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	// Middleware de récupération d'erreurs
	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			log.Printf("Panic occurred: %s", err)
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	}))

	// Taille maximale des corps de requête
	router.Use(middleware.LimitRequestBody())

	// Middleware CORS
	router.Use(middleware.SetupCORS(cfg))

	// Middleware SSO (détection des headers Authentik)
	router.Use(ssoMiddleware.DetectSSO())

	// Serve uploaded files statically
	router.Static("/uploads", cfg.Storage.UploadDir)

	// Routes publiques
	api := router.Group("/api/v1")
	{
		// Gamification
		gamification := api.Group("/gamification")
		gamification.Use(authMiddleware.RequireAuth())
		{
			gamification.GET("/profile", gamificationHandler.GetMyProfile)
			gamification.GET("/achievements", gamificationHandler.GetMyAchievements)
			gamification.GET("/achievements/all", gamificationHandler.GetAllAchievements)
			gamification.GET("/leaderboard", gamificationHandler.GetLeaderboard)
			gamification.GET("/transactions", gamificationHandler.GetMyTransactions)
		}

		auth := api.Group("/auth")
		{
			// Limite par IP contre le brute-force (le verrouillage par compte est dans Login)
			auth.POST("/login", middleware.AuthRateLimit(), authHandler.Login)
			auth.POST("/register", middleware.AuthRateLimit(), authHandler.Register)
			auth.POST("/refresh", authHandler.RefreshToken)

			// Route pour vérifier si l'inscription est activée
			signup := auth.Group("/signup")
			{
				signup.GET("/status", authHandler.GetSignupStatus)
			}

			// Route SSO auto-login (accessible publiquement mais nécessite headers Authentik)
			sso := auth.Group("/sso")
			{
				sso.GET("/auto-login", authHandler.SSOAutoLogin)
			}

			// Routes OAuth publiques
			oauth := auth.Group("/oauth")
			{
				oauth.GET("/providers", oauthHandler.GetEnabledProviders)
				oauth.GET("/:provider/initiate", oauthHandler.InitiateOAuth)
				// Accepter GET et POST pour le callback (Microsoft redirige en GET, frontend peut POST)
				oauth.GET("/:provider/callback", oauthHandler.OAuthCallback)
				oauth.POST("/:provider/callback", oauthHandler.OAuthCallback)
			}
		}

		// Routes version (publiques)
		version := api.Group("/version")
		{
			version.GET("", versionHandler.GetVersion)
			version.GET("/check-updates", versionHandler.CheckForUpdates)
		}
	}

	// Routes protégées - Ordre correct: Auth d'abord, puis CSRF
	protected := api.Group("/")
	protected.Use(authMiddleware.RequireAuth())
	protected.Use(middleware.OptionalCSRFProtection(csrfManager))
	{
		// Route pour générer un token CSRF
		protected.POST("/auth/csrf-token", middleware.CSRFTokenHandler(csrfManager))

		// Profil utilisateur
		protected.GET("/auth/profile", authHandler.GetProfile)
		protected.PUT("/auth/profile", authHandler.UpdateProfile)
		protected.POST("/auth/change-password", authHandler.ChangePassword)
		protected.POST("/auth/avatar", authHandler.UploadAvatar)
		protected.DELETE("/auth/avatar", authHandler.DeleteAvatar)

		// Dashboard
		protected.GET("/dashboard", dashboardHandler.GetDashboard)

		// Home page
		protected.GET("/home", homeHandler.GetHomeData)

		// Paramètres publics (palette, branding — accessible à tous les utilisateurs connectés)
		protected.GET("/settings/public", settingsHandler.GetPublicSettings)

		// Routes favorites
		user := protected.Group("/user")
		{
			user.GET("/favorites", favoritesHandler.GetUserFavorites)
			user.POST("/favorites", favoritesHandler.AddFavorite)
			user.DELETE("/favorites/:id", favoritesHandler.RemoveFavorite)
			user.GET("/favorites/:id/check", favoritesHandler.IsFavorite)
		}

		// Routes analytics (tracking accessible à tous les utilisateurs connectés)
		analytics := protected.Group("/analytics")
		{
			analytics.POST("/track", analyticsHandler.TrackClick)
		}

		// Recherche globale
		protected.GET("/search", searchHandler.GlobalSearch)

		// Routes announcements (accessible à tous les utilisateurs connectés)
		protected.GET("/announcements", announcementHandler.GetActiveAnnouncements)

		// Chatbots actifs (accessible à tous les utilisateurs connectés)
		protected.GET("/chatbots/active", chatbotHandler.GetActiveChatbots)
		protected.POST("/chatbots/:id/interact", chatbotHandler.TrackInteraction)

		// Routes News Hub (accessible à tous les utilisateurs connectés)
		news := protected.Group("/news")
		{
			news.GET("", newsHandler.GetNews) // Liste des news avec filtres

			// Routes spécifiques d'abord (avant les routes avec paramètres)
			news.GET("/unread/count", newsHandler.GetUnreadCount) // Nombre de news non lues
			news.GET("/categories", newsHandler.GetCategories)    // Catégories (lecture seule)
			news.GET("/types", newsHandler.GetNewsTypes)          // Types d'articles (lecture seule)
			news.GET("/tags", newsHandler.GetTags)                // Tags (lecture seule)

			// Routes avec ID numérique
			news.POST("/:id/view", newsHandler.IncrementView)     // Incrémenter les vues
			news.GET("/:id/reactions", newsHandler.GetReactions)  // Récupérer les réactions
			news.POST("/:id/react", newsHandler.AddReaction)      // Ajouter une réaction
			news.DELETE("/:id/react", newsHandler.RemoveReaction) // Retirer une réaction

			// Route slug en dernier (greedy wildcard)
			news.GET("/article/:slug", newsHandler.GetNewsBySlug) // Récupérer une news par slug
		}

		// Routes Media (accessible à tous les utilisateurs connectés - editors et admins peuvent uploader)
		media := protected.Group("/media")
		{
			media.GET("", mediaHandler.GetMediaList)       // Liste des médias avec pagination et filtres
			media.GET("/:id", mediaHandler.GetMedia)       // Récupérer un média par ID
			media.DELETE("/:id", mediaHandler.DeleteMedia) // Supprimer un média (uploader ou admin)
		}

		// Routes Events (accessible à tous les utilisateurs connectés)
		events := protected.Group("/events")
		{
			events.GET("", eventsHandler.GetEvents)                // Liste des événements avec filtres
			events.GET("/calendar", eventsHandler.GetCalendarView) // Vue calendrier (expand récurrences)
			events.GET("/categories", eventsHandler.GetCategories) // Catégories (lecture seule)
			events.GET("/:slug", eventsHandler.GetEventBySlug)     // Récupérer un événement par slug
		}

		// Routes Commentaires (accessible à tous les utilisateurs connectés)
		comments := protected.Group("/comments")
		{
			comments.GET("", commentHandler.GetComments)                 // Récupérer les commentaires d'une entité
			comments.POST("", commentHandler.CreateComment)              // Créer un commentaire
			comments.PUT("/:id", commentHandler.UpdateComment)           // Modifier un commentaire
			comments.DELETE("/:id", commentHandler.DeleteComment)        // Supprimer un commentaire
			comments.GET("/settings", commentHandler.GetCommentSettings) // Récupérer les paramètres
		}

		// Routes Feedback (accessible à tous les utilisateurs connectés)
		feedback := protected.Group("/feedback")
		{
			feedback.GET("/stats", feedbackHandler.GetFeedbackStats) // Statistiques de feedback
			feedback.POST("", feedbackHandler.AddFeedback)           // Ajouter/modifier un feedback
			feedback.DELETE("", feedbackHandler.RemoveFeedback)      // Supprimer un feedback
		}

		// Routes Notifications (accessible à tous les utilisateurs connectés)
		notifications := protected.Group("/notifications")
		{
			notifications.GET("", notificationHandler.GetNotifications)            // Récupérer les notifications
			notifications.GET("/unread/count", notificationHandler.GetUnreadCount) // Nombre de notifications non lues
			notifications.GET("/stats", notificationHandler.GetNotificationStats)  // Statistiques
			notifications.PUT("/:id/read", notificationHandler.MarkAsRead)         // Marquer comme lue
			notifications.PUT("/read-all", notificationHandler.MarkAllAsRead)      // Tout marquer comme lu
			notifications.DELETE("/:id", notificationHandler.DeleteNotification)   // Supprimer une notification
			notifications.DELETE("/read/all", notificationHandler.DeleteAllRead)   // Supprimer toutes les notifications lues
		}

		// Routes Polls (accessible à tous les utilisateurs connectés)
		polls := protected.Group("/polls")
		{
			polls.GET("", pollsHandler.GetPolls)                   // Liste des sondages avec filtres
			polls.GET("/:id", pollsHandler.GetPollByID)            // Récupérer un sondage par ID
			polls.POST("/:id/vote", pollsHandler.Vote)             // Voter pour un sondage
			polls.DELETE("/:id/vote", pollsHandler.RemoveVote)     // Retirer son vote
			polls.GET("/:id/results", pollsHandler.GetPollResults) // Récupérer les résultats d'un sondage
		}

		// Trombinoscope (lecture pour tous les utilisateurs connectés)
		protected.GET("/directory", directoryHandler.GetDirectory)

		// Routes Suggestions (accessible à tous les utilisateurs connectés)
		suggestions := protected.Group("/suggestions")
		{
			suggestions.GET("/categories", suggestionsHandler.GetSuggestionCategories)
			suggestions.GET("", suggestionsHandler.GetSuggestions)
			suggestions.GET("/:id", suggestionsHandler.GetSuggestionByID)
			suggestions.POST("", suggestionsHandler.CreateSuggestion)
			suggestions.POST("/:id/vote", suggestionsHandler.VoteSuggestion)
		}

		// Routes Chat (accessible à tous les utilisateurs connectés)
		chatGroup := protected.Group("/chat")
		{
			chatGroup.GET("/contacts", chatHandler.GetContacts)
			chatGroup.GET("/history", chatHandler.GetHistory)
			chatGroup.DELETE("/messages/:id", chatHandler.DeleteMessage)
			chatGroup.DELETE("/history", chatHandler.ClearConversation)
		}

		// Route WebSocket (publique mais sécurisée par token en query param si nécessaire, ou gérée par middleware si header supporté)
		// Note : protected.Group use authMiddleware, which mostly looks for Authorization Header.
		// Native WebSockets in browsers don't support custom headers easily.
		// We might need a separate route group or accept query param auth in middleware.
		// For MVP, if AuthMiddleware supports checking Cookie or Query Token, it works.
		// Let's assume AuthMiddleware checks header. We might need to make WS explicit.
		// IMPORTANT: For now, we put it under protected, implying the client must find a way to pass auth (e.g. Protocol header or if we change middleware).
		// EASIER: Make it public but do manual check as we implemented in ServeWS.
		// But ServeWS relies on context UserID set by Middleware.
		// -> We will stick to protected and assume client sends token (e.g. via library that supports it or cookie).
		// If using browser native WebSocket, we usually use logic in ServeWS to parse query param "?token=..." if header missing.
		// Let's explicitly allow /ws to be outside main protected group if needed, but for now we try inside.
		// actually, our ServeWS implementation checks context "user_id". So it NEEDS the middleware.
		// We can tell middleware to look at query param "d_token" or "token".
		// For this implementation, let's keep it here.
		protected.GET("/ws", chatHandler.ServeWS)

		// Routes admin
		admin := protected.Group("/admin")
		admin.Use(authMiddleware.RequireAdmin())
		{
			// Gestion des groupes d'applications
			admin.GET("/app-groups", adminHandler.GetAppGroups)
			admin.POST("/app-groups", adminHandler.CreateAppGroup)
			admin.PUT("/app-groups/:id", adminHandler.UpdateAppGroup)
			admin.DELETE("/app-groups/:id", adminHandler.DeleteAppGroup)

			// Gestion des applications
			admin.GET("/applications", adminHandler.GetApplications)
			admin.POST("/applications", adminHandler.CreateApplication)
			admin.PUT("/applications/:id", adminHandler.UpdateApplication)
			admin.DELETE("/applications/:id", adminHandler.DeleteApplication)

			// Gestion des utilisateurs
			admin.GET("/users", adminHandler.GetUsers)
			admin.POST("/users", adminHandler.CreateUser)
			admin.PUT("/users/:id", adminHandler.UpdateUser)
			admin.DELETE("/users/:id", adminHandler.DeleteUser)
			admin.GET("/users/deleted", adminHandler.GetDeletedUsers)
			admin.POST("/users/:id/restore", adminHandler.RestoreUser)
			admin.DELETE("/users/:id/permanent", adminHandler.PermanentlyDeleteUser)

			// Gestion des groupes d'utilisateurs
			admin.GET("/groups", adminHandler.GetGroups)
			admin.POST("/groups", adminHandler.CreateGroup)
			admin.PUT("/groups/:id", adminHandler.UpdateGroup)
			admin.DELETE("/groups/:id", adminHandler.DeleteGroup)

			// Gestion des group admins (admin uniquement)
			admin.GET("/groups/:id/admins", adminHandler.GetGroupAdmins)
			admin.PUT("/groups/:id/admins", adminHandler.AssignGroupAdmins)

			// Gestion des paramètres de l'application
			admin.GET("/settings", settingsHandler.GetAppSettings)
			admin.PUT("/settings", settingsHandler.UpdateAppSettings)
			admin.POST("/settings/reset", settingsHandler.ResetAppSettings)

			// Gestion des messages Hero
			admin.GET("/settings/hero-messages", settingsHandler.GetHeroMessages)
			admin.POST("/settings/hero-messages", settingsHandler.CreateHeroMessage)
			admin.PUT("/settings/hero-messages/:id", settingsHandler.UpdateHeroMessage)
			admin.DELETE("/settings/hero-messages/:id", settingsHandler.DeleteHeroMessage)

			// Gestion des fournisseurs OAuth
			admin.GET("/oauth/providers", oauthHandler.GetAllProviders)
			admin.PUT("/oauth/providers/:id", oauthHandler.UpdateProvider)

			// Analytics (réservé aux admins)
			admin.GET("/analytics/dashboard", analyticsHandler.GetDashboard)
			admin.GET("/analytics/applications/:id", analyticsHandler.GetApplicationStats)
			admin.GET("/analytics/users/:id", analyticsHandler.GetUserStats)

			// Reporting (réservé aux admins)
			admin.GET("/reports/roles", reportsHandler.GetRoleReport)
			admin.GET("/reports/groups", reportsHandler.GetGroupReport)
			admin.GET("/reports/groups/:id", reportsHandler.GetGroupDetailReport)
			admin.GET("/reports/users", reportsHandler.GetUserReport)
			admin.GET("/reports/users/:id", reportsHandler.GetUserDetailReport)

			// Gestion des annonces (réservé aux admins)
			admin.GET("/announcements", announcementHandler.GetAllAnnouncements)
			admin.GET("/announcements/:id", announcementHandler.GetAnnouncement)
			admin.POST("/announcements", announcementHandler.CreateAnnouncement)
			admin.PUT("/announcements/:id", announcementHandler.UpdateAnnouncement)
			admin.DELETE("/announcements/:id", announcementHandler.DeleteAnnouncement)

			// Gestion des chatbots n8n (réservé aux admins)
			admin.GET("/chatbots", chatbotHandler.GetChatbots)
			admin.POST("/chatbots", chatbotHandler.CreateChatbot)
			admin.PUT("/chatbots/:id", chatbotHandler.UpdateChatbot)
			admin.DELETE("/chatbots/:id", chatbotHandler.DeleteChatbot)

			// Gestion de la base de données
			admin.POST("/database/reset", adminHandler.ResetDatabase)

			// Gestion des catégories de news (admin uniquement)
			admin.POST("/news/categories", newsHandler.CreateCategory)
			admin.PUT("/news/categories/:id", newsHandler.UpdateCategory)
			admin.DELETE("/news/categories/:id", newsHandler.DeleteCategory)

			// Gestion des types d'articles (admin uniquement)
			admin.POST("/news/types", newsHandler.CreateNewsType)
			admin.PUT("/news/types/:id", newsHandler.UpdateNewsType)
			admin.DELETE("/news/types/:id", newsHandler.DeleteNewsType)

			// Épingler des news (admin uniquement)
			admin.POST("/news/:id/pin", newsHandler.TogglePin)

			// Analytics News (admin uniquement)
			admin.GET("/news/analytics", newsHandler.GetAnalytics)

			// Gestion des événements (admin uniquement)
			admin.GET("/events", eventsHandler.ListEvents)
			admin.POST("/events", eventsHandler.CreateEvent)
			admin.PUT("/events/:id", eventsHandler.UpdateEvent)
			admin.DELETE("/events/:id", eventsHandler.DeleteEvent)

			// Gestion des catégories d'événements (admin uniquement)
			admin.POST("/events/categories", eventsHandler.CreateCategory)
			admin.PUT("/events/categories/:id", eventsHandler.UpdateCategory)
			admin.DELETE("/events/categories/:id", eventsHandler.DeleteCategory)

			// Analytics Events (admin uniquement)
			admin.GET("/events/analytics", eventsHandler.GetAnalytics)

			// Gestion des jours fériés (admin uniquement)
			admin.GET("/events/holidays/countries", eventsHandler.GetAvailableCountries)
			admin.GET("/events/holidays/preview", eventsHandler.PreviewHolidays)
			admin.POST("/events/holidays/import", eventsHandler.ImportHolidays)
			admin.DELETE("/events/holidays", eventsHandler.DeleteHolidays)

			// Gestion des emails et notifications
			admin.GET("/email/smtp", emailHandler.GetSMTPConfig)
			admin.PUT("/email/smtp", emailHandler.UpdateSMTPConfig)
			admin.POST("/email/smtp/test", emailHandler.TestSMTPConfig)
			admin.GET("/email/templates", emailHandler.GetEmailTemplates)
			admin.GET("/email/templates/variables", emailHandler.GetTemplateVariables)
			admin.GET("/email/templates/:type", emailHandler.GetEmailTemplate)
			admin.PUT("/email/templates/:type", emailHandler.UpdateEmailTemplate)
			admin.POST("/email/templates/:type/reset", emailHandler.ResetEmailTemplate)
			admin.GET("/email/templates/:type/preview", emailHandler.PreviewTemplate)
			admin.GET("/email/logs", emailHandler.GetEmailLogs)

			// OAuth 2.0 configuration for email (admin only)
			admin.GET("/email/oauth", emailHandler.GetOAuthConfig)
			admin.PUT("/email/oauth", emailHandler.UpdateOAuthConfig)
			admin.POST("/email/oauth/test", emailHandler.TestOAuthConnection)
			admin.POST("/email/oauth/refresh", emailHandler.RefreshOAuthToken)
			admin.GET("/email/health", emailHandler.GetEmailHealthStatus)

			// Gestion des commentaires (modération - admin uniquement)
			admin.GET("/comments/pending", commentHandler.GetPendingComments)     // Commentaires en attente
			admin.POST("/comments/moderate", commentHandler.ModerateComment)      // Modérer un commentaire
			admin.PUT("/comments/settings", commentHandler.UpdateCommentSettings) // Mettre à jour les paramètres

			// Gestion des feedbacks (admin uniquement)
			admin.GET("/feedback/all", feedbackHandler.GetAllFeedback) // Tous les feedbacks d'une entité

			// Gestion des sondages (admin uniquement)
			admin.POST("/polls", pollsHandler.CreatePoll)
			admin.PUT("/polls/:id", pollsHandler.UpdatePoll)
			admin.DELETE("/polls/:id", pollsHandler.DeletePoll)
			admin.POST("/polls/:id/close", pollsHandler.ClosePoll)
			admin.GET("/polls/analytics", pollsHandler.GetAnalytics)

			// Gestion de la gamification (admin)
			admin.GET("/gamification/rules", gamificationHandler.GetRules)
			admin.POST("/gamification/rules", gamificationHandler.UpsertRule)
			admin.GET("/gamification/achievements", gamificationHandler.GetAdminAchievements)
			admin.POST("/gamification/achievements", gamificationHandler.CreateAchievement)
			admin.PUT("/gamification/achievements/:id", gamificationHandler.UpdateAchievement)
			admin.DELETE("/gamification/achievements/:id", gamificationHandler.DeleteAchievement)

			// Gestion des suggestions (admin uniquement)
			admin.GET("/suggestions", suggestionsHandler.GetAdminSuggestions)
			admin.PATCH("/suggestions/:id/status", suggestionsHandler.UpdateSuggestionStatus)
			admin.PATCH("/suggestions/:id/archive", suggestionsHandler.UpdateSuggestionArchiveState)
			admin.GET("/suggestion-categories", suggestionsHandler.GetAdminSuggestionCategories)
			admin.POST("/suggestion-categories", suggestionsHandler.CreateSuggestionCategory)
			admin.PUT("/suggestion-categories/:id", suggestionsHandler.UpdateSuggestionCategory)
			admin.DELETE("/suggestion-categories/:id", suggestionsHandler.DeleteSuggestionCategory)

			// Gestion des médias (admin uniquement)
			admin.GET("/media", mediaHandler.GetMediaList)        // Liste des médias avec pagination et filtres
			admin.GET("/media/stats", mediaHandler.GetMediaStats) // Totaux de toute la médiathèque
			admin.GET("/media/:id", mediaHandler.GetMedia)        // Récupérer un média par ID
			admin.POST("/media/upload", mediaHandler.UploadMedia) // Uploader un média
			admin.PUT("/media/:id", mediaHandler.UpdateMedia)     // Mettre à jour les métadonnées d'un média
			admin.DELETE("/media/:id", mediaHandler.DeleteMedia)  // Supprimer un média

			// Trombinoscope : structure et champs dynamiques (admin uniquement)
			admin.POST("/directory/units", directoryHandler.CreateUnit)
			admin.PUT("/directory/units/:id", directoryHandler.UpdateUnit)
			admin.DELETE("/directory/units/:id", directoryHandler.DeleteUnit)
			admin.POST("/directory/positions", directoryHandler.CreatePosition)
			admin.PUT("/directory/positions/:id", directoryHandler.UpdatePosition)
			admin.DELETE("/directory/positions/:id", directoryHandler.DeletePosition)
			admin.POST("/directory/fields", directoryHandler.CreateField)
			admin.PUT("/directory/fields/:id", directoryHandler.UpdateField)
			admin.DELETE("/directory/fields/:id", directoryHandler.DeleteField)
		}

		// Routes editor (admin et editor peuvent créer/modifier des news et événements)
		editor := protected.Group("/editor")
		editor.Use(authMiddleware.RequireEditor())
		{
			// Gestion des news
			editor.POST("/news", newsHandler.CreateNews)
			editor.PUT("/news/:id", newsHandler.UpdateNews)
			editor.DELETE("/news/:id", newsHandler.DeleteNews)

			// Gestion des tags (editors peuvent créer des tags)
			editor.POST("/news/tags", newsHandler.CreateTag)
			editor.PUT("/news/tags/:id", newsHandler.UpdateTag)
			editor.DELETE("/news/tags/:id", newsHandler.DeleteTag)

			// Upload de médias (editors, group_admins et admins peuvent uploader)
			editor.POST("/media/upload", mediaHandler.UploadMedia)

			// Gestion des événements
			editor.POST("/events", eventsHandler.CreateEvent)
			editor.PUT("/events/:id", eventsHandler.UpdateEvent)
			editor.DELETE("/events/:id", eventsHandler.DeleteEvent)

			// Modération des commentaires (editors peuvent aussi modérer)
			editor.GET("/comments/pending", commentHandler.GetPendingComments)
			editor.POST("/comments/moderate", commentHandler.ModerateComment)

			// Gestion des sondages (editors peuvent créer/modifier/supprimer des sondages)
			editor.POST("/polls", pollsHandler.CreatePoll)
			editor.PUT("/polls/:id", pollsHandler.UpdatePoll)
			editor.DELETE("/polls/:id", pollsHandler.DeletePoll)
		}

		// Routes group-admin (gestion limitée au périmètre)
		groupAdmin := protected.Group("/group-admin")
		groupAdmin.Use(authMiddleware.RequireGroupAdmin())
		{
			// AppGroups (scoped)
			groupAdmin.GET("/app-groups", groupAdminHandler.GetAppGroups)
			groupAdmin.POST("/app-groups", adminHandler.CreateAppGroup)
			groupAdmin.PUT("/app-groups/:id", adminHandler.UpdateAppGroup)
			groupAdmin.DELETE("/app-groups/:id", adminHandler.DeleteAppGroup)

			// Applications (scoped)
			groupAdmin.GET("/applications", groupAdminHandler.GetApplications)
			groupAdmin.POST("/applications", groupAdminHandler.CreateApplication)
			groupAdmin.PUT("/applications/:id", groupAdminHandler.UpdateApplication)
			groupAdmin.DELETE("/applications/:id", groupAdminHandler.DeleteApplication)

			// News (scoped)
			groupAdmin.GET("/news", newsHandler.GetNews) // Liste des news avec filtrage automatique par rôle
			groupAdmin.POST("/news", newsHandler.CreateNews)
			groupAdmin.PUT("/news/:id", newsHandler.UpdateNews)
			groupAdmin.DELETE("/news/:id", newsHandler.DeleteNews)

			// Upload de médias
			groupAdmin.POST("/media/upload", mediaHandler.UploadMedia)

			// Tags : un admin de groupe peut en créer (en rédigeant un article ou un événement).
			// Modifier ou supprimer un tag partagé est réservé aux admins globaux.
			adminOnly := authMiddleware.RequireAdmin()
			groupAdmin.POST("/news/tags", newsHandler.CreateTag)
			groupAdmin.PUT("/news/tags/:id", adminOnly, newsHandler.UpdateTag)
			groupAdmin.DELETE("/news/tags/:id", adminOnly, newsHandler.DeleteTag)

			// Catégories : partagées par toute l'organisation, réservées aux admins globaux
			groupAdmin.POST("/news/categories", adminOnly, newsHandler.CreateCategory)
			groupAdmin.PUT("/news/categories/:id", adminOnly, newsHandler.UpdateCategory)
			groupAdmin.DELETE("/news/categories/:id", adminOnly, newsHandler.DeleteCategory)

			// Events (scoped)
			groupAdmin.GET("/events", eventsHandler.GetEvents) // Liste des événements avec filtrage automatique par rôle
			groupAdmin.POST("/events", eventsHandler.CreateEventGroupAdmin)
			groupAdmin.PUT("/events/:id", eventsHandler.UpdateEventGroupAdmin)
			groupAdmin.DELETE("/events/:id", eventsHandler.DeleteEventGroupAdmin)

			// Polls (scoped - group admin peut gérer les sondages de ses groupes)
			groupAdmin.GET("/polls", pollsHandler.GetPolls) // Liste des sondages avec filtrage automatique par rôle
			groupAdmin.POST("/polls", pollsHandler.CreatePoll)
			groupAdmin.PUT("/polls/:id", pollsHandler.UpdatePoll)
			groupAdmin.DELETE("/polls/:id", pollsHandler.DeletePoll)
			groupAdmin.POST("/polls/:id/close", pollsHandler.ClosePoll)

			// Info sur les groupes administrés
			groupAdmin.GET("/managed-groups", groupAdminHandler.GetManagedGroups)

			// Trombinoscope : personnes et affectations, réservées aux admins globaux
			// (l'annuaire est partagé par toute l'organisation, sans notion de groupe)
			groupAdmin.GET("/directory/users", adminOnly, directoryHandler.SearchUsers)
			groupAdmin.POST("/directory/people", adminOnly, directoryHandler.CreatePerson)
			groupAdmin.PUT("/directory/people/:id", adminOnly, directoryHandler.UpdatePerson)
			groupAdmin.DELETE("/directory/people/:id", adminOnly, directoryHandler.DeletePerson)
			groupAdmin.POST("/directory/assignments", adminOnly, directoryHandler.CreateAssignment)
			groupAdmin.PUT("/directory/assignments/:id", adminOnly, directoryHandler.UpdateAssignment)
			groupAdmin.DELETE("/directory/assignments/:id", adminOnly, directoryHandler.DeleteAssignment)
		}
	}

	// Route de santé
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Airboard API is running",
		})
	})

	// Documentation Swagger (optionnel)
	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}

// trustedProxiesFromEnv lit TRUSTED_PROXIES, sinon renvoie la liste par défaut :
// localhost, plus les réseaux privés (Docker) en mode release.
func trustedProxiesFromEnv(mode string) []string {
	if raw := strings.TrimSpace(os.Getenv("TRUSTED_PROXIES")); raw != "" {
		var proxies []string
		for _, p := range strings.Split(raw, ",") {
			if p = strings.TrimSpace(p); p != "" {
				proxies = append(proxies, p)
			}
		}
		return proxies
	}

	proxies := []string{"127.0.0.1", "::1"}
	if mode == "release" {
		proxies = append(proxies, "172.16.0.0/12", "192.168.0.0/16", "10.0.0.0/8")
	}
	return proxies
}
