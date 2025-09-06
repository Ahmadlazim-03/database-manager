package main

import (
	"log"

	"db-manager-backend/config"
	"db-manager-backend/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Load environment variables
	config.LoadEnv()

	// Connect to database
	config.ConnectDB()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB limit for request body
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			log.Printf("Error: %v", err)
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Initialize handlers
	authHandler := handlers.NewAuthHandler()
	dbHandler := handlers.NewDatabaseHandler()
	apiHandler := handlers.NewAPIHandler()
	dbManagementHandler := handlers.NewDatabaseManagementHandler()
	dynamicAPIHandler := handlers.NewDynamicAPIHandlerOptimized() // Use optimized version
	sharingHandler := handlers.NewSharingHandler()

	// Routes
	api := app.Group("/api")

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Get("/profile", handlers.JWTMiddleware, authHandler.GetProfile)

	// Database routes (protected)
	database := api.Group("/database", handlers.JWTMiddleware)
	database.Post("/test", dbHandler.TestConnection)
	database.Post("/", dbHandler.CreateConnection)
	database.Get("/", dbHandler.GetConnections)
	database.Get("/:id/info", dbHandler.GetDatabaseInfo)
	database.Delete("/:id", dbHandler.DeleteConnection)

	// Database Management routes (protected)
	dbManagement := api.Group("/database-management", handlers.JWTMiddleware)
	dbManagement.Get("/collections", dbManagementHandler.GetCollections)
	dbManagement.Get("/collections/:collection/schema", dbManagementHandler.GetCollectionSchema)
	dbManagement.Get("/collections/:collection/documents", dbManagementHandler.GetDocuments)
	dbManagement.Post("/collections/:collection/documents", dbManagementHandler.CreateDocument)
	dbManagement.Put("/collections/:collection/documents/:id", dbManagementHandler.UpdateDocument)
	dbManagement.Delete("/collections/:collection/documents/:id", dbManagementHandler.DeleteDocument)

	// API management routes (protected)
	apiGroup := api.Group("/api-management", handlers.JWTMiddleware)
	apiGroup.Post("/keys", apiHandler.CreateAPIKey)
	apiGroup.Get("/keys", apiHandler.GetAPIKeys)
	apiGroup.Put("/keys/:id/toggle", apiHandler.ToggleAPIKey)
	apiGroup.Delete("/keys/:id", apiHandler.DeleteAPIKey)
	apiGroup.Post("/endpoints", apiHandler.CreateEndpoint)
	apiGroup.Get("/endpoints", apiHandler.GetEndpoints)
	apiGroup.Put("/endpoints/:id/toggle", apiHandler.ToggleEndpoint)
	apiGroup.Delete("/endpoints/:id", apiHandler.DeleteEndpoint)
	apiGroup.Get("/logs", apiHandler.GetLogs)
	apiGroup.Delete("/logs", apiHandler.ClearLogs)

	// Memory monitoring endpoint (protected)
	apiGroup.Get("/memory-stats", dynamicAPIHandler.GetMemoryStats)

	// Database sharing routes (protected)
	sharing := api.Group("/sharing", handlers.JWTMiddleware)
	sharing.Post("/invitations", sharingHandler.CreateInvitation)
	sharing.Get("/invitations/database/:databaseId", sharingHandler.GetDatabaseInvitations)
	sharing.Get("/invitations/:token", sharingHandler.GetInvitation)
	sharing.Post("/invitations/:token/accept", sharingHandler.AcceptInvitation)
	sharing.Get("/shared-databases", sharingHandler.GetSharedDatabases)
	sharing.Get("/pending-invitations", sharingHandler.GetPendingInvitations)
	sharing.Get("/database-access/:databaseId", sharingHandler.GetDatabaseAccess)
	sharing.Delete("/access", sharingHandler.RevokeAccess)
	sharing.Delete("/invitations/:invitationId", sharingHandler.RevokeInvitation)
	sharing.Delete("/leave", sharingHandler.LeaveSharedDatabase)

	// Dynamic API routes (public with API key)
	dynamicAPI := api.Group("/:collection", 
		dynamicAPIHandler.MemoryMonitor,
		dynamicAPIHandler.ValidateAPIKey, 
		dynamicAPIHandler.ValidateEndpoint,
		dynamicAPIHandler.LogRequest)
	dynamicAPI.Get("/", dynamicAPIHandler.HandleGET)
	dynamicAPI.Get("/:id", dynamicAPIHandler.HandleGET)
	dynamicAPI.Post("/", dynamicAPIHandler.HandlePOST)
	dynamicAPI.Put("/:id", dynamicAPIHandler.HandlePUT)
	dynamicAPI.Delete("/:id", dynamicAPIHandler.HandleDELETE)

	// Start server
	port := config.GetEnv("PORT", "8080")
	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
