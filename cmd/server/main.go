package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"

	"jia/server/internal/config"
	"jia/server/internal/database"
	"jia/server/internal/handlers"
	"jia/server/internal/middleware"
	"jia/server/internal/repositories"
	"jia/server/internal/services"
	"jia/server/internal/ws"
)

func main() {
	// 1. Initialize configuration & database
	config.LoadConfig()
	database.ConnectDB()

	// 2. Instantiate Repositories
	settingsRepo := repositories.NewSettingsRepository()
	userRepo := repositories.NewUserRepository()
	keyRepo := repositories.NewKeyRepository()
	convRepo := repositories.NewConversationRepository()
	msgRepo := repositories.NewMessageRepository()
	pushRepo := repositories.NewPushRepository()
	inviteRepo := repositories.NewInviteRepository()
	contactRepo := repositories.NewContactRepository()
	sessionRepo := repositories.NewSessionRepository()

	// 3. Instantiate Services
	setupService := services.NewSetupService(settingsRepo, userRepo)
	authService := services.NewAuthService(userRepo, sessionRepo, settingsRepo, inviteRepo)
	keyService := services.NewKeyService(keyRepo)
	storageService := services.NewStorageService(settingsRepo)
	pushService := services.NewPushService(pushRepo, settingsRepo)
	adminService := services.NewAdminService(userRepo, settingsRepo, inviteRepo, msgRepo)
	convService := services.NewConversationService(convRepo, userRepo)
	msgService := services.NewMessageService(msgRepo, convRepo)
	contactService := services.NewContactService(contactRepo, userRepo)

	// 4. Instantiate WebSocket structures
	hub := ws.NewHub(convRepo)
	go hub.Run()

	// Connect hooks between MessageService and WebSocket Hub for instant broadcasts & push alerts
	hub.BindMessageServiceHooks(msgService, pushService)

	// Wire reload hook so that services update dynamically when settings change (e.g. S3 credentials or FCM tokens)
	services.InitializeDynamicServices = func() {
		log.Println("Settings changed! Reloading dynamic configuration for S3 and Push notification services...")
		storageService.ReloadConfig()
		pushService.ReloadConfig()
	}

	// 5. Instantiate Handlers
	setupHandler := handlers.NewSetupHandler(setupService)
	authHandler := handlers.NewAuthHandler(authService)
	adminHandler := handlers.NewAdminHandler(adminService)
	userHandler := handlers.NewUserHandler(userRepo)
	keyHandler := handlers.NewKeyHandler(keyService)
	convHandler := handlers.NewConversationHandler(convService)
	msgHandler := handlers.NewMessageHandler(msgService, storageService)
	pushHandler := handlers.NewPushHandler(pushService)
	attachmentHandler := handlers.NewAttachmentHandler(msgRepo, convRepo, storageService)
	contactHandler := handlers.NewContactHandler(contactService)
	wsHandler := ws.NewWSHandler(hub, msgService, convService)

	// 6. Bootstrap Fiber Application
	app := fiber.New(fiber.Config{
		AppName: "Jia Messenger Server v1.0",
	})

	// Global Middlewares
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods: []string{"GET", "POST", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"},
	}))

	// 7. Route Registrations

	// Unauthenticated Setup gates
	setupGate := app.Group("/api/setup")
	setupGate.Use(middleware.SetupIncompleteRequired(setupService))
	setupGate.Post("/", setupHandler.Setup)

	app.Get("/api/setup/status", setupHandler.GetStatus)

	// Normal API operations (Must require setup completion)
	api := app.Group("/api")
	api.Use(middleware.SetupCompletedRequired(setupService))

	// Auth operations
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)
	auth.Post("/logout", authHandler.Logout)

	// Authenticated routes
	authRequired := api.Group("/")
	authRequired.Use(middleware.AuthRequired)

	// WebSocket Upgrade
	authRequired.Get("/ws", wsHandler.Upgrade)

	// Users endpoints
	users := authRequired.Group("/users")
	users.Get("/me", userHandler.GetMe)
	users.Patch("/me", userHandler.UpdateMe)
	users.Get("/search", userHandler.Search)
	users.Get("/:id", userHandler.GetByID)

	// Keys endpoints (E2E)
	keys := authRequired.Group("/keys")
	keys.Post("/bundle", keyHandler.UploadBundle)
	keys.Get("/:userId", keyHandler.GetBundle)
	keys.Post("/prekeys", keyHandler.ReplenishPrekeys)

	// Conversations endpoints
	conversations := authRequired.Group("/conversations")
	conversations.Get("/", convHandler.List)
	conversations.Post("/", convHandler.Create)
	conversations.Get("/:id", convHandler.GetDetails)
	conversations.Patch("/:id", convHandler.Update)
	conversations.Delete("/:id", convHandler.Leave)
	conversations.Post("/:id/participants", convHandler.AddParticipants)
	conversations.Delete("/:id/participants/:userId", convHandler.RemoveParticipant)
	conversations.Patch("/:id/read", convHandler.MarkRead)

	// Messages endpoints
	conversations.Get("/:id/messages", msgHandler.GetHistory)
	conversations.Post("/:id/messages", msgHandler.Send)

	messages := authRequired.Group("/messages")
	messages.Patch("/:id", msgHandler.Edit)
	messages.Delete("/:id", msgHandler.Delete)
	messages.Post("/:id/reactions", msgHandler.AddReaction)
	messages.Delete("/:id/reactions/:emoji", msgHandler.RemoveReaction)

	// Attachments endpoints
	authRequired.Get("/attachments/:id/url", attachmentHandler.GetPresignedURL)

	// Push endpoints
	push := authRequired.Group("/push")
	push.Post("/subscribe", pushHandler.Subscribe)
	push.Delete("/subscribe", pushHandler.Unsubscribe)

	// Contacts endpoints
	contacts := authRequired.Group("/contacts")
	contacts.Get("/", contactHandler.List)
	contacts.Post("/", contactHandler.Add)
	contacts.Delete("/:id", contactHandler.Remove)
	contacts.Patch("/:id", contactHandler.Update)

	// Admin endpoints (AdminRequired)
	admin := api.Group("/admin")
	admin.Use(middleware.AuthRequired)
	admin.Use(middleware.AdminRequired)

	admin.Get("/stats", adminHandler.GetStats)
	admin.Get("/settings", adminHandler.GetSettings)
	admin.Patch("/settings", adminHandler.UpdateSettings)

	admin.Get("/users", adminHandler.ListUsers)
	admin.Patch("/users/:id", adminHandler.UpdateUserRole)
	admin.Delete("/users/:id", adminHandler.DeleteUser)

	admin.Get("/invites", adminHandler.ListInvites)
	admin.Post("/invites", adminHandler.CreateInvite)
	admin.Delete("/invites/:id", adminHandler.RevokeInvite)

	// 8. Start HTTP Server
	port := config.AppConfig.Port
	log.Printf("Server starting on port %s...", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
