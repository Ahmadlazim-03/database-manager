package handlers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"db-manager-backend/config"
	"db-manager-backend/models"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DynamicAPIHandler struct{}

func NewDynamicAPIHandler() *DynamicAPIHandler {
	return &DynamicAPIHandler{}
}

// Middleware untuk validasi API key
func (h *DynamicAPIHandler) ValidateAPIKey(c *fiber.Ctx) error {
	apiKey := c.Get("X-API-Key")
	if apiKey == "" {
		return c.Status(401).JSON(fiber.Map{
			"error": "Missing API key",
		})
	}

	var key models.APIKey
	if err := config.DB.Preload("Database").Where("key = ? AND is_active = ?", apiKey, true).First(&key).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid API key",
		})
	}

	c.Locals("apiKey", key)
	c.Locals("database", key.Database)
	return c.Next()
}

// Log API request
func (h *DynamicAPIHandler) LogRequest(c *fiber.Ctx) error {
	start := time.Now()

	// Continue with the request
	err := c.Next()

	// Get response data before goroutine
	statusCode := c.Response().StatusCode()
	method := c.Method()
	path := c.Path()
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")

	// Log the request after completion
	go func() {
		apiKey, ok := c.Locals("apiKey").(models.APIKey)
		if !ok {
			return
		}

		// Find the endpoint
		var endpoint models.APIEndpoint
		config.DB.Where("path = ? AND method = ? AND database_id = ?", 
			path, method, apiKey.DatabaseID).First(&endpoint)

		logEntry := models.APILog{
			APIKeyID:     apiKey.ID,
			EndpointID:   endpoint.ID,
			Method:       method,
			Path:         path,
			StatusCode:   statusCode,
			ResponseTime: time.Since(start).Milliseconds(),
			IPAddress:    ipAddress,
			UserAgent:    userAgent,
		}

		config.DB.Create(&logEntry)
	}()

	return err
}

// Generic CRUD operations
func (h *DynamicAPIHandler) HandleGET(c *fiber.Ctx) error {
	database := c.Locals("database").(models.DatabaseConnection)
	collection := c.Params("collection")
	id := c.Params("id", "")

	switch database.Type {
	case "mongodb":
		return h.handleMongoGET(c, database, collection, id)
	case "mysql", "postgres":
		return h.handleSQLGET(c, database, collection, id)
	default:
		return c.Status(400).JSON(fiber.Map{
			"error": "Unsupported database type",
		})
	}
}

func (h *DynamicAPIHandler) HandlePOST(c *fiber.Ctx) error {
	database := c.Locals("database").(models.DatabaseConnection)
	collection := c.Params("collection")

	switch database.Type {
	case "mongodb":
		return h.handleMongoPOST(c, database, collection)
	case "mysql", "postgres":
		return h.handleSQLPOST(c, database, collection)
	default:
		return c.Status(400).JSON(fiber.Map{
			"error": "Unsupported database type",
		})
	}
}

func (h *DynamicAPIHandler) HandlePUT(c *fiber.Ctx) error {
	database := c.Locals("database").(models.DatabaseConnection)
	collection := c.Params("collection")
	id := c.Params("id")

	switch database.Type {
	case "mongodb":
		return h.handleMongoPUT(c, database, collection, id)
	case "mysql", "postgres":
		return h.handleSQLPUT(c, database, collection, id)
	default:
		return c.Status(400).JSON(fiber.Map{
			"error": "Unsupported database type",
		})
	}
}

func (h *DynamicAPIHandler) HandleDELETE(c *fiber.Ctx) error {
	database := c.Locals("database").(models.DatabaseConnection)
	collection := c.Params("collection")
	id := c.Params("id")

	switch database.Type {
	case "mongodb":
		return h.handleMongoDELETE(c, database, collection, id)
	case "mysql", "postgres":
		return h.handleSQLDELETE(c, database, collection, id)
	default:
		return c.Status(400).JSON(fiber.Map{
			"error": "Unsupported database type",
		})
	}
}

// MongoDB handlers
func (h *DynamicAPIHandler) handleMongoGET(c *fiber.Ctx, database models.DatabaseConnection, collection, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var uri string
	if database.Username != "" && database.Password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
			database.Username, database.Password, database.Host, database.Port, database.Database)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%d/%s", database.Host, database.Port, database.Database)
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database connection failed"})
	}
	defer client.Disconnect(ctx)

	db := client.Database(database.Database)
	coll := db.Collection(collection)

	if id != "" {
		// Get single document
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid ID format"})
		}

		var result bson.M
		err = coll.FindOne(ctx, bson.M{"_id": objectID}).Decode(&result)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return c.Status(404).JSON(fiber.Map{"error": "Document not found"})
			}
			return c.Status(500).JSON(fiber.Map{"error": "Database query failed"})
		}

		return c.JSON(result)
	} else {
		// Get all documents with pagination
		page, _ := strconv.Atoi(c.Query("page", "1"))
		limit, _ := strconv.Atoi(c.Query("limit", "10"))
		skip := (page - 1) * limit

		cursor, err := coll.Find(ctx, bson.M{}, options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Database query failed"})
		}
		defer cursor.Close(ctx)

		var results []bson.M
		if err = cursor.All(ctx, &results); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to decode results"})
		}

		total, _ := coll.CountDocuments(ctx, bson.M{})

		return c.JSON(fiber.Map{
			"data":  results,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}

func (h *DynamicAPIHandler) handleMongoPOST(c *fiber.Ctx, database models.DatabaseConnection, collection string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var uri string
	if database.Username != "" && database.Password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
			database.Username, database.Password, database.Host, database.Port, database.Database)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%d/%s", database.Host, database.Port, database.Database)
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database connection failed"})
	}
	defer client.Disconnect(ctx)

	var document bson.M
	if err := c.BodyParser(&document); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	db := client.Database(database.Database)
	coll := db.Collection(collection)

	result, err := coll.InsertOne(ctx, document)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to insert document"})
	}

	return c.Status(201).JSON(fiber.Map{
		"id":      result.InsertedID,
		"message": "Document created successfully",
	})
}

func (h *DynamicAPIHandler) handleMongoPUT(c *fiber.Ctx, database models.DatabaseConnection, collection, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	var uri string
	if database.Username != "" && database.Password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
			database.Username, database.Password, database.Host, database.Port, database.Database)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%d/%s", database.Host, database.Port, database.Database)
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database connection failed"})
	}
	defer client.Disconnect(ctx)

	var update bson.M
	if err := c.BodyParser(&update); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	db := client.Database(database.Database)
	coll := db.Collection(collection)

	result, err := coll.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": update})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update document"})
	}

	if result.MatchedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Document not found"})
	}

	return c.JSON(fiber.Map{
		"message": "Document updated successfully",
	})
}

func (h *DynamicAPIHandler) handleMongoDELETE(c *fiber.Ctx, database models.DatabaseConnection, collection, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	var uri string
	if database.Username != "" && database.Password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
			database.Username, database.Password, database.Host, database.Port, database.Database)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%d/%s", database.Host, database.Port, database.Database)
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database connection failed"})
	}
	defer client.Disconnect(ctx)

	db := client.Database(database.Database)
	coll := db.Collection(collection)

	result, err := coll.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete document"})
	}

	if result.DeletedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Document not found"})
	}

	return c.JSON(fiber.Map{
		"message": "Document deleted successfully",
	})
}

// SQL handlers (simplified - would need more sophisticated handling for production)
func (h *DynamicAPIHandler) handleSQLGET(c *fiber.Ctx, database models.DatabaseConnection, table, id string) error {
	var db *gorm.DB
	var err error

	if database.Type == "mysql" {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			database.Username, database.Password, database.Host, database.Port, database.Database)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	} else if database.Type == "postgres" {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			database.Host, database.Username, database.Password, database.Database, database.Port)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database connection failed"})
	}

	if id != "" {
		// Get single record
		var result map[string]interface{}
		if err := db.Table(table).Where("id = ?", id).First(&result).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(404).JSON(fiber.Map{"error": "Record not found"})
			}
			return c.Status(500).JSON(fiber.Map{"error": "Database query failed"})
		}
		return c.JSON(result)
	} else {
		// Get all records with pagination
		page, _ := strconv.Atoi(c.Query("page", "1"))
		limit, _ := strconv.Atoi(c.Query("limit", "10"))
		offset := (page - 1) * limit

		var results []map[string]interface{}
		if err := db.Table(table).Offset(offset).Limit(limit).Find(&results).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Database query failed"})
		}

		var total int64
		db.Table(table).Count(&total)

		return c.JSON(fiber.Map{
			"data":  results,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}

func (h *DynamicAPIHandler) handleSQLPOST(c *fiber.Ctx, database models.DatabaseConnection, table string) error {
	var db *gorm.DB
	var err error

	if database.Type == "mysql" {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			database.Username, database.Password, database.Host, database.Port, database.Database)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	} else if database.Type == "postgres" {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			database.Host, database.Username, database.Password, database.Database, database.Port)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database connection failed"})
	}

	var data map[string]interface{}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	if err := db.Table(table).Create(&data).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create record"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Record created successfully",
		"data":    data,
	})
}

func (h *DynamicAPIHandler) handleSQLPUT(c *fiber.Ctx, database models.DatabaseConnection, table, id string) error {
	var db *gorm.DB
	var err error

	if database.Type == "mysql" {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			database.Username, database.Password, database.Host, database.Port, database.Database)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	} else if database.Type == "postgres" {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			database.Host, database.Username, database.Password, database.Database, database.Port)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database connection failed"})
	}

	var data map[string]interface{}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	result := db.Table(table).Where("id = ?", id).Updates(data)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update record"})
	}

	if result.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Record not found"})
	}

	return c.JSON(fiber.Map{
		"message": "Record updated successfully",
	})
}

func (h *DynamicAPIHandler) handleSQLDELETE(c *fiber.Ctx, database models.DatabaseConnection, table, id string) error {
	var db *gorm.DB
	var err error

	if database.Type == "mysql" {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			database.Username, database.Password, database.Host, database.Port, database.Database)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	} else if database.Type == "postgres" {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			database.Host, database.Username, database.Password, database.Database, database.Port)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database connection failed"})
	}

	result := db.Table(table).Where("id = ?", id).Delete(nil)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete record"})
	}

	if result.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Record not found"})
	}

	return c.JSON(fiber.Map{
		"message": "Record deleted successfully",
	})
}
