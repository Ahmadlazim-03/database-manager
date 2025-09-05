package handlers

import (
	"db-manager-backend/config"
	"db-manager-backend/models"
	"db-manager-backend/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type DatabaseHandler struct {
	dbService *services.DatabaseService
}

type CreateConnectionRequest struct {
	Name     string `json:"name" validate:"required"`
	Type     string `json:"type" validate:"required"`
	Host     string `json:"host" validate:"required"`
	Port     int    `json:"port" validate:"required"`
	Database string `json:"database" validate:"required"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewDatabaseHandler() *DatabaseHandler {
	return &DatabaseHandler{
		dbService: services.NewDatabaseService(),
	}
}

func (h *DatabaseHandler) TestConnection(c *fiber.Ctx) error {
	var req services.ConnectionParams
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.dbService.TestConnection(req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Connection failed: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Connection successful",
	})
}

func (h *DatabaseHandler) CreateConnection(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	
	var req CreateConnectionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Test connection first
	params := services.ConnectionParams{
		Type:     req.Type,
		Host:     req.Host,
		Port:     req.Port,
		Database: req.Database,
		Username: req.Username,
		Password: req.Password,
	}

	if err := h.dbService.TestConnection(params); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Connection failed: " + err.Error(),
		})
	}

	// Create database connection record
	userUUID, _ := uuid.Parse(userID)
	dbConn := models.DatabaseConnection{
		UserID:   userUUID,
		Name:     req.Name,
		Type:     req.Type,
		Host:     req.Host,
		Port:     req.Port,
		Database: req.Database,
		Username: req.Username,
		Password: req.Password,
		Status:   "active",
	}

	if err := config.DB.Create(&dbConn).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save connection",
		})
	}

	return c.JSON(dbConn)
}

func (h *DatabaseHandler) GetConnections(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var connections []models.DatabaseConnection
	if err := config.DB.Where("user_id = ?", userID).Find(&connections).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch connections",
		})
	}

	return c.JSON(connections)
}

func (h *DatabaseHandler) GetDatabaseInfo(c *fiber.Ctx) error {
	connectionID := c.Params("id")
	userID := c.Locals("user_id").(string)

	var dbConn models.DatabaseConnection
	if err := config.DB.Where("id = ? AND user_id = ?", connectionID, userID).First(&dbConn).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Connection not found",
		})
	}

	params := services.ConnectionParams{
		Type:     dbConn.Type,
		Host:     dbConn.Host,
		Port:     dbConn.Port,
		Database: dbConn.Database,
		Username: dbConn.Username,
		Password: dbConn.Password,
	}

	info, err := h.dbService.GetDatabaseInfo(params)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get database info: " + err.Error(),
		})
	}

	return c.JSON(info)
}

func (h *DatabaseHandler) DeleteConnection(c *fiber.Ctx) error {
	connectionID := c.Params("id")
	userID := c.Locals("user_id").(string)

	if err := config.DB.Where("id = ? AND user_id = ?", connectionID, userID).Delete(&models.DatabaseConnection{}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete connection",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Connection deleted successfully",
	})
}
