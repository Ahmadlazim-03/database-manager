package handlers

import (
	"db-manager-backend/config"
	"db-manager-backend/models"
	"db-manager-backend/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type APIHandler struct{}

type CreateAPIKeyRequest struct {
	DatabaseID string `json:"database_id" validate:"required"`
	Name       string `json:"name" validate:"required"`
}

type CreateEndpointRequest struct {
	DatabaseID string `json:"database_id" validate:"required"`
	Collection string `json:"collection" validate:"required"`
	Method     string `json:"method" validate:"required"`
}

func NewAPIHandler() *APIHandler {
	return &APIHandler{}
}

func (h *APIHandler) CreateAPIKey(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	
	var req CreateAPIKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Verify database belongs to user
	var dbConn models.DatabaseConnection
	if err := config.DB.Where("id = ? AND user_id = ?", req.DatabaseID, userID).First(&dbConn).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	userUUID, _ := uuid.Parse(userID)
	dbUUID, _ := uuid.Parse(req.DatabaseID)

	apiKey := models.APIKey{
		UserID:     userUUID,
		DatabaseID: dbUUID,
		Name:       req.Name,
		Key:        utils.GenerateAPIKey(),
		IsActive:   true,
	}

	if err := config.DB.Create(&apiKey).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create API key",
		})
	}

	return c.JSON(apiKey)
}

func (h *APIHandler) GetAPIKeys(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var apiKeys []models.APIKey
	if err := config.DB.Preload("Database").Where("user_id = ?", userID).Find(&apiKeys).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch API keys",
		})
	}

	return c.JSON(apiKeys)
}

func (h *APIHandler) ToggleAPIKey(c *fiber.Ctx) error {
	keyID := c.Params("id")
	userID := c.Locals("user_id").(string)

	var apiKey models.APIKey
	if err := config.DB.Where("id = ? AND user_id = ?", keyID, userID).First(&apiKey).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "API key not found",
		})
	}

	apiKey.IsActive = !apiKey.IsActive
	if err := config.DB.Save(&apiKey).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update API key",
		})
	}

	return c.JSON(apiKey)
}

func (h *APIHandler) CreateEndpoint(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	
	var req CreateEndpointRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Verify database belongs to user
	var dbConn models.DatabaseConnection
	if err := config.DB.Where("id = ? AND user_id = ?", req.DatabaseID, userID).First(&dbConn).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	dbUUID, _ := uuid.Parse(req.DatabaseID)
	
	endpoint := models.APIEndpoint{
		DatabaseID: dbUUID,
		Collection: req.Collection,
		Path:       "/api/" + req.Collection,
		Method:     req.Method,
		IsActive:   true,
	}

	if err := config.DB.Create(&endpoint).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create endpoint",
		})
	}

	return c.JSON(endpoint)
}

func (h *APIHandler) GetEndpoints(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	databaseID := c.Query("database_id")

	query := config.DB.Preload("Database")
	if databaseID != "" {
		query = query.Joins("JOIN database_connections ON api_endpoints.database_id = database_connections.id").
			Where("database_connections.user_id = ? AND api_endpoints.database_id = ?", userID, databaseID)
	} else {
		query = query.Joins("JOIN database_connections ON api_endpoints.database_id = database_connections.id").
			Where("database_connections.user_id = ?", userID)
	}

	var endpoints []models.APIEndpoint
	if err := query.Find(&endpoints).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch endpoints",
		})
	}

	return c.JSON(endpoints)
}

func (h *APIHandler) ToggleEndpoint(c *fiber.Ctx) error {
	endpointID := c.Params("id")
	userID := c.Locals("user_id").(string)

	var endpoint models.APIEndpoint
	if err := config.DB.Joins("JOIN database_connections ON api_endpoints.database_id = database_connections.id").
		Where("api_endpoints.id = ? AND database_connections.user_id = ?", endpointID, userID).
		First(&endpoint).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Endpoint not found",
		})
	}

	endpoint.IsActive = !endpoint.IsActive
	if err := config.DB.Save(&endpoint).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update endpoint",
		})
	}

	return c.JSON(endpoint)
}

func (h *APIHandler) GetLogs(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	
	var logs []models.APILog
	if err := config.DB.Preload("APIKey").Preload("Endpoint").
		Joins("JOIN api_keys ON api_logs.api_key_id = api_keys.id").
		Where("api_keys.user_id = ?", userID).
		Order("api_logs.created_at DESC").
		Limit(100).
		Find(&logs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch logs",
		})
	}

	return c.JSON(logs)
}

func (h *APIHandler) DeleteAPIKey(c *fiber.Ctx) error {
	keyID := c.Params("id")
	userID := c.Locals("user_id").(string)

	var apiKey models.APIKey
	if err := config.DB.Where("id = ? AND user_id = ?", keyID, userID).First(&apiKey).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "API key not found",
		})
	}

	if err := config.DB.Delete(&apiKey).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete API key",
		})
	}

	return c.JSON(fiber.Map{
		"message": "API key deleted successfully",
	})
}

func (h *APIHandler) DeleteEndpoint(c *fiber.Ctx) error {
	endpointID := c.Params("id")
	userID := c.Locals("user_id").(string)

	var endpoint models.APIEndpoint
	if err := config.DB.Joins("JOIN database_connections ON api_endpoints.database_id = database_connections.id").
		Where("api_endpoints.id = ? AND database_connections.user_id = ?", endpointID, userID).
		First(&endpoint).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Endpoint not found",
		})
	}

	if err := config.DB.Delete(&endpoint).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete endpoint",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Endpoint deleted successfully",
	})
}

func (h *APIHandler) ClearLogs(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	
	// Delete all logs for user's API keys
	if err := config.DB.Exec(`
		DELETE FROM api_logs 
		WHERE api_key_id IN (
			SELECT id FROM api_keys WHERE user_id = ?
		)
	`, userID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to clear logs",
		})
	}

	return c.JSON(fiber.Map{
		"message": "All logs cleared successfully",
	})
}
