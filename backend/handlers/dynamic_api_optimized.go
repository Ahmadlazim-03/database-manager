package handlers

import (
	"context"
	"fmt"
	"runtime"
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

// DynamicAPIHandlerOptimized - Memory optimized version using pointers
type DynamicAPIHandlerOptimized struct {
	dbConnPool map[string]*gorm.DB // Connection pool untuk reuse
}

func NewDynamicAPIHandlerOptimized() *DynamicAPIHandlerOptimized {
	return &DynamicAPIHandlerOptimized{
		dbConnPool: make(map[string]*gorm.DB),
	}
}

// Memory monitoring middleware
func (h *DynamicAPIHandlerOptimized) MemoryMonitor(c *fiber.Ctx) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Log memory usage
	c.Set("X-Memory-Alloc", fmt.Sprintf("%d KB", m.Alloc/1024))
	c.Set("X-Memory-Total", fmt.Sprintf("%d KB", m.TotalAlloc/1024))
	c.Set("X-Memory-Sys", fmt.Sprintf("%d KB", m.Sys/1024))
	c.Set("X-Goroutines", fmt.Sprintf("%d", runtime.NumGoroutine()))
	
	return c.Next()
}

// ValidateEndpoint - Check if endpoint is active
func (h *DynamicAPIHandlerOptimized) ValidateEndpoint(c *fiber.Ctx) error {
	collection := c.Params("collection")
	method := c.Method()
	
	// Get database from locals
	databasePtr, ok := c.Locals("database").(*models.DatabaseConnection)
	if !ok || databasePtr == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database connection not found"})
	}

	// Check if endpoint exists and is active
	var endpoint models.APIEndpoint
	if err := config.DB.Where("database_id = ? AND collection = ? AND method = ? AND is_active = ?", 
		databasePtr.ID, collection, method, true).First(&endpoint).Error; err != nil {
		return c.Status(403).JSON(fiber.Map{
			"error": "Endpoint not found or inactive",
		})
	}

	c.Locals("endpoint", &endpoint)
	return c.Next()
}

// Optimized ValidateAPIKey using pointer
func (h *DynamicAPIHandlerOptimized) ValidateAPIKey(c *fiber.Ctx) error {
	apiKey := c.Get("X-API-Key")
	if apiKey == "" {
		return c.Status(401).JSON(fiber.Map{
			"error": "Missing API key",
		})
	}

	// Use pointer to avoid copying struct
	var key *models.APIKey = &models.APIKey{}
	if err := config.DB.Preload("Database").Where("key = ? AND is_active = ?", apiKey, true).First(key).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid API key",
		})
	}

	// Store pointer in locals
	c.Locals("apiKey", key)
	c.Locals("database", &key.Database)
	return c.Next()
}

// Optimized LogRequest with memory-efficient approach
func (h *DynamicAPIHandlerOptimized) LogRequest(c *fiber.Ctx) error {
	start := time.Now()
	err := c.Next()

	// Collect data before goroutine to avoid race conditions
	logData := struct {
		statusCode int
		method     string
		path       string
		ipAddress  string
		userAgent  string
		duration   int64
	}{
		statusCode: c.Response().StatusCode(),
		method:     c.Method(),
		path:       c.Path(),
		ipAddress:  c.IP(),
		userAgent:  c.Get("User-Agent"),
		duration:   time.Since(start).Milliseconds(),
	}

	// Use buffered channel to prevent goroutine leak
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Handle panic gracefully
				return
			}
		}()

		apiKeyPtr, ok := c.Locals("apiKey").(*models.APIKey)
		if !ok || apiKeyPtr == nil {
			return
		}

		// Use pointer to avoid struct copying
		endpoint := &models.APIEndpoint{}
		config.DB.Where("path = ? AND method = ? AND database_id = ?", 
			logData.path, logData.method, apiKeyPtr.DatabaseID).First(endpoint)

		// Create log entry using pointers
		logEntry := &models.APILog{
			APIKeyID:     apiKeyPtr.ID,
			EndpointID:   endpoint.ID,
			Method:       logData.method,
			Path:         logData.path,
			StatusCode:   logData.statusCode,
			ResponseTime: logData.duration,
			IPAddress:    logData.ipAddress,
			UserAgent:    logData.userAgent,
		}

		config.DB.Create(logEntry)
	}()

	return err
}

// Connection pool for database reuse
func (h *DynamicAPIHandlerOptimized) getDBConnection(database *models.DatabaseConnection) (*gorm.DB, error) {
	// Create connection key
	connKey := fmt.Sprintf("%s_%s_%s_%d", database.Type, database.Host, database.Database, database.Port)
	
	// Check if connection exists in pool
	if db, exists := h.dbConnPool[connKey]; exists {
		// Test connection
		sqlDB, _ := db.DB()
		if err := sqlDB.Ping(); err == nil {
			return db, nil
		}
		// Remove dead connection
		delete(h.dbConnPool, connKey)
	}

	// Create new connection
	var db *gorm.DB
	var err error

	switch database.Type {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			database.Username, database.Password, database.Host, database.Port, database.Database)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			database.Host, database.Username, database.Password, database.Database, database.Port)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported database type")
	}

	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Store in pool
	h.dbConnPool[connKey] = db
	return db, nil
}

// Optimized HandlePOST using pointers and address
func (h *DynamicAPIHandlerOptimized) HandlePOST(c *fiber.Ctx) error {
	// Use pointer from locals
	databasePtr, ok := c.Locals("database").(*models.DatabaseConnection)
	if !ok || databasePtr == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database connection not found"})
	}
	
	collection := c.Params("collection")

	switch databasePtr.Type {
	case "mongodb":
		return h.handleMongoPOSTOptimized(c, databasePtr, collection)
	case "mysql", "postgres":
		return h.handleSQLPOSTOptimized(c, databasePtr, collection)
	default:
		return c.Status(400).JSON(fiber.Map{
			"error": "Unsupported database type",
		})
	}
}

// Optimized SQL POST handler
func (h *DynamicAPIHandlerOptimized) handleSQLPOSTOptimized(c *fiber.Ctx, database *models.DatabaseConnection, table string) error {
	// Use connection pool
	db, err := h.getDBConnection(database)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database connection failed"})
	}

	// Use pointer for data to avoid copying
	data := make(map[string]interface{})
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	// Use address of data
	if err := db.Table(table).Create(&data).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create record",
			"details": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Record created successfully",
		"data":    &data, // Return pointer to avoid copying
	})
}

// MongoDB optimized handler
func (h *DynamicAPIHandlerOptimized) handleMongoPOSTOptimized(c *fiber.Ctx, database *models.DatabaseConnection, collection string) error {
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

	// Use pointer for document
	document := make(bson.M)
	if err := c.BodyParser(&document); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	db := client.Database(database.Database)
	coll := db.Collection(collection)

	result, err := coll.InsertOne(ctx, &document) // Use address
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to insert document"})
	}

	return c.Status(201).JSON(fiber.Map{
		"id":      result.InsertedID,
		"message": "Document created successfully",
		"data":    &document, // Return pointer
	})
}

// Memory usage endpoint
func (h *DynamicAPIHandlerOptimized) GetMemoryStats(c *fiber.Ctx) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Force garbage collection for accurate reading
	runtime.GC()
	runtime.ReadMemStats(&m)

	stats := map[string]interface{}{
		"memory": map[string]interface{}{
			"alloc_kb":       m.Alloc / 1024,
			"total_alloc_kb": m.TotalAlloc / 1024,
			"sys_kb":         m.Sys / 1024,
			"heap_kb":        m.HeapAlloc / 1024,
			"heap_sys_kb":    m.HeapSys / 1024,
			"heap_idle_kb":   m.HeapIdle / 1024,
			"heap_inuse_kb":  m.HeapInuse / 1024,
			"stack_kb":       m.StackSys / 1024,
		},
		"gc": map[string]interface{}{
			"num_gc":        m.NumGC,
			"gc_cpu_fraction": m.GCCPUFraction,
		},
		"goroutines": runtime.NumGoroutine(),
		"connections": map[string]interface{}{
			"pool_size": len(h.dbConnPool),
			"active_connections": func() int {
				active := 0
				for _, db := range h.dbConnPool {
					if sqlDB, err := db.DB(); err == nil {
						if stats := sqlDB.Stats(); stats.OpenConnections > 0 {
							active++
						}
					}
				}
				return active
			}(),
		},
	}

	return c.JSON(stats)
}

// Handle GET requests
func (h *DynamicAPIHandlerOptimized) HandleGET(c *fiber.Ctx) error {
	databasePtr, ok := c.Locals("database").(*models.DatabaseConnection)
	if !ok || databasePtr == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database connection not found"})
	}
	
	collection := c.Params("collection")
	id := c.Params("id", "")

	switch databasePtr.Type {
	case "mongodb":
		return h.handleMongoGETOptimized(c, databasePtr, collection, id)
	case "mysql", "postgres":
		return h.handleSQLGETOptimized(c, databasePtr, collection, id)
	default:
		return c.Status(400).JSON(fiber.Map{"error": "Unsupported database type"})
	}
}

// Handle PUT requests
func (h *DynamicAPIHandlerOptimized) HandlePUT(c *fiber.Ctx) error {
	databasePtr, ok := c.Locals("database").(*models.DatabaseConnection)
	if !ok || databasePtr == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database connection not found"})
	}
	
	collection := c.Params("collection")
	id := c.Params("id")

	switch databasePtr.Type {
	case "mongodb":
		return h.handleMongoPUTOptimized(c, databasePtr, collection, id)
	case "mysql", "postgres":
		return h.handleSQLPUTOptimized(c, databasePtr, collection, id)
	default:
		return c.Status(400).JSON(fiber.Map{"error": "Unsupported database type"})
	}
}

// Handle DELETE requests
func (h *DynamicAPIHandlerOptimized) HandleDELETE(c *fiber.Ctx) error {
	databasePtr, ok := c.Locals("database").(*models.DatabaseConnection)
	if !ok || databasePtr == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database connection not found"})
	}
	
	collection := c.Params("collection")
	id := c.Params("id")

	switch databasePtr.Type {
	case "mongodb":
		return h.handleMongoDELETEOptimized(c, databasePtr, collection, id)
	case "mysql", "postgres":
		return h.handleSQLDELETEOptimized(c, databasePtr, collection, id)
	default:
		return c.Status(400).JSON(fiber.Map{"error": "Unsupported database type"})
	}
}

// Optimized SQL GET handler
func (h *DynamicAPIHandlerOptimized) handleSQLGETOptimized(c *fiber.Ctx, database *models.DatabaseConnection, table, id string) error {
	db, err := h.getDBConnection(database)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database connection failed"})
	}

	if id != "" {
		result := make(map[string]interface{})
		if err := db.Table(table).Where("id = ?", id).First(&result).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(404).JSON(fiber.Map{"error": "Record not found"})
			}
			return c.Status(500).JSON(fiber.Map{"error": "Database query failed"})
		}
		return c.JSON(&result) // Return pointer
	} else {
		page, _ := strconv.Atoi(c.Query("page", "1"))
		limit, _ := strconv.Atoi(c.Query("limit", "10"))
		offset := (page - 1) * limit

		results := make([]map[string]interface{}, 0)
		if err := db.Table(table).Offset(offset).Limit(limit).Find(&results).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Database query failed"})
		}

		var total int64
		db.Table(table).Count(&total)

		return c.JSON(fiber.Map{
			"data":  &results, // Return pointer
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}

// Optimized SQL PUT handler
func (h *DynamicAPIHandlerOptimized) handleSQLPUTOptimized(c *fiber.Ctx, database *models.DatabaseConnection, table, id string) error {
	db, err := h.getDBConnection(database)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database connection failed"})
	}

	data := make(map[string]interface{})
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	result := db.Table(table).Where("id = ?", id).Updates(&data)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update record"})
	}

	if result.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Record not found"})
	}

	return c.JSON(fiber.Map{
		"message": "Record updated successfully",
		"data":    &data,
	})
}

// Optimized SQL DELETE handler
func (h *DynamicAPIHandlerOptimized) handleSQLDELETEOptimized(c *fiber.Ctx, database *models.DatabaseConnection, table, id string) error {
	db, err := h.getDBConnection(database)
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

// Optimized MongoDB handlers
func (h *DynamicAPIHandlerOptimized) handleMongoGETOptimized(c *fiber.Ctx, database *models.DatabaseConnection, collection, id string) error {
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
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid ID format"})
		}

		result := make(bson.M)
		err = coll.FindOne(ctx, bson.M{"_id": objectID}).Decode(&result)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return c.Status(404).JSON(fiber.Map{"error": "Document not found"})
			}
			return c.Status(500).JSON(fiber.Map{"error": "Database query failed"})
		}

		return c.JSON(&result)
	} else {
		page, _ := strconv.Atoi(c.Query("page", "1"))
		limit, _ := strconv.Atoi(c.Query("limit", "10"))
		skip := (page - 1) * limit

		cursor, err := coll.Find(ctx, bson.M{}, options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Database query failed"})
		}
		defer cursor.Close(ctx)

		results := make([]bson.M, 0)
		if err = cursor.All(ctx, &results); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to decode results"})
		}

		total, _ := coll.CountDocuments(ctx, bson.M{})

		return c.JSON(fiber.Map{
			"data":  &results,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}

func (h *DynamicAPIHandlerOptimized) handleMongoPUTOptimized(c *fiber.Ctx, database *models.DatabaseConnection, collection, id string) error {
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

	update := make(bson.M)
	if err := c.BodyParser(&update); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	db := client.Database(database.Database)
	coll := db.Collection(collection)

	result, err := coll.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": &update})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update document"})
	}

	if result.MatchedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Document not found"})
	}

	return c.JSON(fiber.Map{
		"message": "Document updated successfully",
		"data":    &update,
	})
}

func (h *DynamicAPIHandlerOptimized) handleMongoDELETEOptimized(c *fiber.Ctx, database *models.DatabaseConnection, collection, id string) error {
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

// Cleanup function untuk membersihkan connection pool
func (h *DynamicAPIHandlerOptimized) Cleanup() {
	for key, db := range h.dbConnPool {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		delete(h.dbConnPool, key)
	}
}
