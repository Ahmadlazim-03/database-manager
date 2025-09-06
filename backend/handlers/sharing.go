package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"db-manager-backend/config"
	"db-manager-backend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type SharingHandler struct{}

func NewSharingHandler() *SharingHandler {
	return &SharingHandler{}
}

// CreateInvitation creates a new database invitation
func (h *SharingHandler) CreateInvitation(c *fiber.Ctx) error {
	type InvitationRequest struct {
		DatabaseID      string `json:"database_id" validate:"required"`
		InviteeEmail    string `json:"invitee_email" validate:"required,email"`
		PermissionLevel string `json:"permission_level" validate:"required,oneof=read write admin"`
	}

	var req InvitationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get current user
	userIDStr := c.Locals("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Parse database ID
	databaseID, err := uuid.Parse(req.DatabaseID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid database ID",
		})
	}

	// Check if user owns the database
	var database models.DatabaseConnection
	if err := config.DB.Where("id = ? AND user_id = ?", databaseID, userID).First(&database).Error; err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Database not found or access denied",
		})
	}

	// Check if invitation already exists
	var existingInvitation models.DatabaseInvitation
	if err := config.DB.Where("database_id = ? AND invitee_email = ? AND status = ?", 
		databaseID, req.InviteeEmail, "pending").First(&existingInvitation).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invitation already sent to this email",
		})
	}

	// Generate invitation token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate invitation token",
		})
	}
	invitationToken := hex.EncodeToString(tokenBytes)

	// Create invitation
	invitation := models.DatabaseInvitation{
		DatabaseID:      databaseID,
		InviterID:       userID,
		InviteeEmail:    req.InviteeEmail,
		InvitationToken: invitationToken,
		PermissionLevel: req.PermissionLevel,
		Status:          "pending",
		ExpiresAt:       time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := config.DB.Create(&invitation).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create invitation",
		})
	}

	return c.JSON(fiber.Map{
		"message":         "Invitation sent successfully",
		"invitation":      invitation,
		"invitation_link": "http://localhost:5173/join/" + invitationToken,
	})
}

// GetDatabaseInvitations gets all invitations for a database
func (h *SharingHandler) GetDatabaseInvitations(c *fiber.Ctx) error {
	databaseIDStr := c.Params("databaseId")
	userIDStr := c.Locals("user_id").(string)
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	databaseID, err := uuid.Parse(databaseIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid database ID",
		})
	}

	// Check if user owns the database
	var database models.DatabaseConnection
	if err := config.DB.Where("id = ? AND user_id = ?", databaseID, userID).First(&database).Error; err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Database not found or access denied",
		})
	}

	var invitations []models.DatabaseInvitation
	if err := config.DB.Preload("Invitee").Where("database_id = ?", databaseID).
		Order("created_at DESC").Find(&invitations).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get invitations",
		})
	}

	return c.JSON(invitations)
}

// GetInvitation gets invitation details by token
func (h *SharingHandler) GetInvitation(c *fiber.Ctx) error {
	token := c.Params("token")

	var invitation models.DatabaseInvitation
	if err := config.DB.Preload("Database").Preload("Inviter").
		Where("invitation_token = ? AND status = ? AND expires_at > ?", 
		token, "pending", time.Now()).First(&invitation).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Invalid or expired invitation",
		})
	}

	return c.JSON(invitation)
}

// AcceptInvitation accepts a database invitation
func (h *SharingHandler) AcceptInvitation(c *fiber.Ctx) error {
	token := c.Params("token")
	userIDStr := c.Locals("user_id").(string)
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var invitation models.DatabaseInvitation
	if err := config.DB.Where("invitation_token = ? AND status = ? AND expires_at > ?", 
		token, "pending", time.Now()).First(&invitation).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Invalid or expired invitation",
		})
	}

	// Check if user already has access
	var existingAccess models.DatabaseAccess
	if err := config.DB.Where("database_id = ? AND user_id = ?", 
		invitation.DatabaseID, userID).First(&existingAccess).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "You already have access to this database",
		})
	}

	// Create database access
	access := models.DatabaseAccess{
		DatabaseID:      invitation.DatabaseID,
		UserID:          userID,
		PermissionLevel: invitation.PermissionLevel,
		GrantedBy:       invitation.InviterID,
	}

	if err := config.DB.Create(&access).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to grant access",
		})
	}

	// Update invitation status
	now := time.Now()
	invitation.Status = "accepted"
	invitation.AcceptedAt = &now
	invitation.InviteeID = &userID

	if err := config.DB.Save(&invitation).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update invitation",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Invitation accepted successfully",
		"access":  access,
	})
}

// GetSharedDatabases gets databases shared with the user
func (h *SharingHandler) GetSharedDatabases(c *fiber.Ctx) error {
	userIDStr := c.Locals("user_id").(string)
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var accesses []models.DatabaseAccess
	if err := config.DB.Preload("Database").Preload("Grantor").
		Where("user_id = ?", userID).Find(&accesses).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get shared databases",
		})
	}

	return c.JSON(accesses)
}

// GetPendingInvitations gets pending invitations for the current user
func (h *SharingHandler) GetPendingInvitations(c *fiber.Ctx) error {
	userIDStr := c.Locals("user_id").(string)
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Get user email
	var user models.User
	if err := config.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user data",
		})
	}

	var invitations []models.DatabaseInvitation
	if err := config.DB.Preload("Database").Preload("Inviter").
		Where("invitee_email = ? AND status = ? AND expires_at > ?", 
		user.Email, "pending", time.Now()).Find(&invitations).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get pending invitations",
		})
	}

	return c.JSON(invitations)
}

// GetDatabaseAccess gets all users with access to a database
func (h *SharingHandler) GetDatabaseAccess(c *fiber.Ctx) error {
	databaseIDStr := c.Params("databaseId")
	userIDStr := c.Locals("user_id").(string)
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	databaseID, err := uuid.Parse(databaseIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid database ID",
		})
	}

	// Check if user owns the database
	var database models.DatabaseConnection
	if err := config.DB.Where("id = ? AND user_id = ?", databaseID, userID).First(&database).Error; err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Database not found or access denied",
		})
	}

	var accesses []models.DatabaseAccess
	if err := config.DB.Preload("User").Preload("Grantor").
		Where("database_id = ?", databaseID).Find(&accesses).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get database access",
		})
	}

	return c.JSON(accesses)
}

// RevokeAccess revokes user access to a database
func (h *SharingHandler) RevokeAccess(c *fiber.Ctx) error {
	type RevokeRequest struct {
		DatabaseID string `json:"database_id" validate:"required"`
		UserID     string `json:"user_id" validate:"required"`
	}

	var req RevokeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	currentUserIDStr := c.Locals("user_id").(string)
	
	currentUserID, err := uuid.Parse(currentUserIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	databaseID, err := uuid.Parse(req.DatabaseID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid database ID",
		})
	}

	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Check if current user owns the database
	var database models.DatabaseConnection
	if err := config.DB.Where("id = ? AND user_id = ?", databaseID, currentUserID).First(&database).Error; err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Database not found or access denied",
		})
	}

	// Delete access
	if err := config.DB.Where("database_id = ? AND user_id = ?", 
		databaseID, targetUserID).Delete(&models.DatabaseAccess{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to revoke access",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Access revoked successfully",
	})
}

// RevokeInvitation revokes a pending invitation
func (h *SharingHandler) RevokeInvitation(c *fiber.Ctx) error {
	invitationIDStr := c.Params("invitationId")
	userIDStr := c.Locals("user_id").(string)
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	invitationID, err := uuid.Parse(invitationIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid invitation ID",
		})
	}

	// Check if user has permission to revoke
	var invitation models.DatabaseInvitation
	if err := config.DB.Joins("JOIN database_connections ON database_invitations.database_id = database_connections.id").
		Where("database_invitations.id = ? AND database_connections.user_id = ?", 
		invitationID, userID).First(&invitation).Error; err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Invitation not found or access denied",
		})
	}

	if err := config.DB.Delete(&invitation).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to revoke invitation",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Invitation revoked successfully",
	})
}
