package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"db-manager-backend/config"
	"db-manager-backend/models"
	"db-manager-backend/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DatabaseManagementHandler struct {
	dbService *services.DatabaseService
}

func NewDatabaseManagementHandler() *DatabaseManagementHandler {
	return &DatabaseManagementHandler{
		dbService: services.NewDatabaseService(),
	}
}

// Helper function to extract user ID from context
func (h *DatabaseManagementHandler) getUserID(c *fiber.Ctx) (uuid.UUID, error) {
	userIDInterface := c.Locals("user_id")
	if userIDInterface == nil {
		return uuid.Nil, fmt.Errorf("user not authenticated")
	}
	
	userIDStr, ok := userIDInterface.(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid user ID format")
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID")
	}
	
	return userID, nil
}

// GetCollections returns all collections for a database connection
func (h *DatabaseManagementHandler) GetCollections(c *fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	// Get database_id from query parameter
	databaseIDStr := c.Query("database_id")
	if databaseIDStr == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "database_id is required",
		})
	}

	databaseID, err := uuid.Parse(databaseIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid database_id",
		})
	}

	// Get database connection info using GORM
	var connection models.DatabaseConnection
	if err := config.DB.Where("id = ? AND user_id = ?", databaseID, userID).First(&connection).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	var collections []string

	switch connection.Type {
	case "mongodb":
		// Connect to MongoDB
		client, err := h.dbService.ConnectMongoDB(connection)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to connect to MongoDB: " + err.Error(),
			})
		}
		defer client.Disconnect(context.Background())

		database := client.Database(connection.Database)
		collectionNames, err := database.ListCollectionNames(context.Background(), bson.D{})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to list collections: " + err.Error(),
			})
		}
		collections = collectionNames

	case "mysql", "postgresql":
		// For SQL databases, get table names
		sqlClient, err := h.dbService.ConnectSQL(connection)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to connect to database: " + err.Error(),
			})
		}
		defer sqlClient.Close()

		var query string
		if connection.Type == "mysql" {
			query = "SHOW TABLES"
		} else {
			query = "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'"
		}

		rows, err := sqlClient.Query(query)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to list tables: " + err.Error(),
			})
		}
		defer rows.Close()

		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				continue
			}
			collections = append(collections, tableName)
		}

	default:
		return c.Status(400).JSON(fiber.Map{
			"error": "Unsupported database type",
		})
	}

	return c.JSON(collections)
}

// GetCollectionSchema returns the schema/structure of a collection
func (h *DatabaseManagementHandler) GetCollectionSchema(c *fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	collectionName := c.Params("collection")
	databaseIDStr := c.Query("database_id")
	
	if databaseIDStr == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "database_id is required",
		})
	}

	databaseID, err := uuid.Parse(databaseIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid database_id",
		})
	}

	// Get database connection info using GORM
	var connection models.DatabaseConnection
	if err := config.DB.Where("id = ? AND user_id = ?", databaseID, userID).First(&connection).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	var fields []string

	switch connection.Type {
	case "mongodb":
		// Connect to MongoDB and sample documents to get field names
		client, err := h.dbService.ConnectMongoDB(connection)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to connect to MongoDB: " + err.Error(),
			})
		}
		defer client.Disconnect(context.Background())

		collection := client.Database(connection.Database).Collection(collectionName)
		
		// Get a sample document to extract field names
		cursor, err := collection.Find(context.Background(), bson.D{}, options.Find().SetLimit(10))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to sample documents: " + err.Error(),
			})
		}
		defer cursor.Close(context.Background())

		fieldSet := make(map[string]bool)
		for cursor.Next(context.Background()) {
			var doc bson.M
			if err := cursor.Decode(&doc); err != nil {
				continue
			}
			for key := range doc {
				if key != "_id" {
					fieldSet[key] = true
				}
			}
		}

		for field := range fieldSet {
			fields = append(fields, field)
		}

	case "mysql", "postgresql":
		// Get column names for SQL tables
		sqlClient, err := h.dbService.ConnectSQL(connection)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to connect to database: " + err.Error(),
			})
		}
		defer sqlClient.Close()

		var query string
		if connection.Type == "mysql" {
			query = fmt.Sprintf("DESCRIBE %s", collectionName)
		} else {
			query = fmt.Sprintf(`
				SELECT column_name 
				FROM information_schema.columns 
				WHERE table_name = '%s' AND table_schema = 'public'
			`, collectionName)
		}

		rows, err := sqlClient.Query(query)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to get table schema: " + err.Error(),
			})
		}
		defer rows.Close()

		if connection.Type == "mysql" {
			for rows.Next() {
				var field, fieldType, null, key, defaultVal, extra string
				if err := rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra); err != nil {
					continue
				}
				fields = append(fields, field)
			}
		} else {
			for rows.Next() {
				var field string
				if err := rows.Scan(&field); err != nil {
					continue
				}
				fields = append(fields, field)
			}
		}
	}

	return c.JSON(fiber.Map{
		"fields": fields,
	})
}

// GetDocuments returns paginated documents from a collection
func (h *DatabaseManagementHandler) GetDocuments(c *fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	collectionName := c.Params("collection")
	databaseIDStr := c.Query("database_id")
	
	if databaseIDStr == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "database_id is required",
		})
	}

	databaseID, err := uuid.Parse(databaseIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid database_id",
		})
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")
	sortField := c.Query("sort", "")
	sortOrder := c.Query("order", "asc")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get database connection info using GORM
	var connection models.DatabaseConnection
	if err := config.DB.Where("id = ? AND user_id = ?", databaseID, userID).First(&connection).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	var documents []interface{}
	var total int64

	switch connection.Type {
	case "mongodb":
		// Connect to MongoDB
		client, err := h.dbService.ConnectMongoDB(connection)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to connect to MongoDB: " + err.Error(),
			})
		}
		defer client.Disconnect(context.Background())

		collection := client.Database(connection.Database).Collection(collectionName)
		
		// Build filter
		filter := bson.D{}
		if search != "" {
			// Create a regex search for multiple fields
			searchRegex := bson.M{"$regex": search, "$options": "i"}
			filter = bson.D{
				{"$or", bson.A{
					bson.M{"name": searchRegex},
					bson.M{"title": searchRegex},
					bson.M{"description": searchRegex},
					bson.M{"content": searchRegex},
				}},
			}
		}

		// Build sort options
		var sortOptions *options.FindOptions
		if sortField != "" {
			sortDirection := 1
			if sortOrder == "desc" {
				sortDirection = -1
			}
			sortOptions = options.Find().SetSort(bson.D{{sortField, sortDirection}})
		} else {
			sortOptions = options.Find()
		}

		// Get total count
		total, err = collection.CountDocuments(context.Background(), filter)
		if err != nil {
			log.Printf("Error counting documents: %v", err)
		}

		// Get documents with pagination
		skip := (page - 1) * limit
		cursor, err := collection.Find(context.Background(), filter, 
			sortOptions.SetSkip(int64(skip)).SetLimit(int64(limit)))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to fetch documents: " + err.Error(),
			})
		}
		defer cursor.Close(context.Background())

		for cursor.Next(context.Background()) {
			var doc bson.M
			if err := cursor.Decode(&doc); err != nil {
				continue
			}
			// Convert ObjectID to string for JSON
			if id, ok := doc["_id"].(primitive.ObjectID); ok {
				doc["id"] = id.Hex()
			}
			documents = append(documents, doc)
		}

	case "mysql", "postgresql":
		// For SQL databases
		sqlClient, err := h.dbService.ConnectSQL(connection)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to connect to database: " + err.Error(),
			})
		}
		defer sqlClient.Close()

		// Build query
		baseQuery := fmt.Sprintf("SELECT * FROM %s", collectionName)
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", collectionName)
		
		var whereClause string
		var args []interface{}
		
		if search != "" {
			// Get column names first for proper search
			var searchColumns []string
			columnsQuery := fmt.Sprintf("SELECT column_name FROM information_schema.columns WHERE table_name = '%s'", collectionName)
			if connection.Type == "mysql" {
				columnsQuery = fmt.Sprintf("DESCRIBE %s", collectionName)
			}
			
			rows, err := sqlClient.Query(columnsQuery)
			if err == nil {
				for rows.Next() {
					var colName string
					if connection.Type == "mysql" {
						var fieldType, null, key, defaultVal, extra string
						rows.Scan(&colName, &fieldType, &null, &key, &defaultVal, &extra)
					} else {
						rows.Scan(&colName)
					}
					searchColumns = append(searchColumns, colName)
				}
				rows.Close()
			}

			if len(searchColumns) > 0 {
				var searchConditions []string
				for _, col := range searchColumns {
					searchConditions = append(searchConditions, fmt.Sprintf("CAST(%s AS TEXT) ILIKE ?", col))
					args = append(args, "%"+search+"%")
				}
				whereClause = " WHERE " + strings.Join(searchConditions, " OR ")
				baseQuery += whereClause
				countQuery += whereClause
			}
		}

		// Get total count
		var countArgs []interface{}
		if len(args) > 0 {
			countArgs = make([]interface{}, len(args))
			copy(countArgs, args)
		}
		err = sqlClient.QueryRow(countQuery, countArgs...).Scan(&total)
		if err != nil {
			log.Printf("Error counting rows: %v", err)
		}

		// Add sorting and pagination
		if sortField != "" && sortOrder != "" {
			baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortField, strings.ToUpper(sortOrder))
		}
		
		offset := (page - 1) * limit
		baseQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

		rows, err := sqlClient.Query(baseQuery, args...)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to fetch rows: " + err.Error(),
			})
		}
		defer rows.Close()

		// Get column names
		columns, err := rows.Columns()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to get columns: " + err.Error(),
			})
		}

		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range columns {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				continue
			}

			doc := make(map[string]interface{})
			for i, col := range columns {
				val := values[i]
				if b, ok := val.([]byte); ok {
					doc[col] = string(b)
				} else {
					doc[col] = val
				}
			}
			documents = append(documents, doc)
		}
	}

	return c.JSON(fiber.Map{
		"documents": documents,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

// CreateDocument creates a new document in a collection
func (h *DatabaseManagementHandler) CreateDocument(c *fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	collectionName := c.Params("collection")
	
	var req struct {
		DatabaseID string                 `json:"database_id"`
		Data       map[string]interface{} `json:"data"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	databaseID, err := uuid.Parse(req.DatabaseID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid database_id",
		})
	}

	// Get database connection info using GORM
	var connection models.DatabaseConnection
	if err := config.DB.Where("id = ? AND user_id = ?", databaseID, userID).First(&connection).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	switch connection.Type {
	case "mongodb":
		// Connect to MongoDB
		client, err := h.dbService.ConnectMongoDB(connection)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to connect to MongoDB: " + err.Error(),
			})
		}
		defer client.Disconnect(context.Background())

		collection := client.Database(connection.Database).Collection(collectionName)
		
		// Add timestamp
		req.Data["created_at"] = time.Now()
		
		result, err := collection.InsertOne(context.Background(), req.Data)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to create document: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"id":      result.InsertedID,
		})

	case "mysql", "postgresql":
		// For SQL databases, convert map to INSERT statement
		sqlClient, err := h.dbService.ConnectSQL(connection)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to connect to database: " + err.Error(),
			})
		}
		defer sqlClient.Close()

		// Build INSERT query
		var columns []string
		var placeholders []string
		var values []interface{}
		
		for key, value := range req.Data {
			columns = append(columns, key)
			placeholders = append(placeholders, "?")
			values = append(values, value)
		}

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			collectionName,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "))

		result, err := sqlClient.Exec(query, values...)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to create record: " + err.Error(),
			})
		}

		id, _ := result.LastInsertId()
		return c.JSON(fiber.Map{
			"success": true,
			"id":      id,
		})

	default:
		return c.Status(400).JSON(fiber.Map{
			"error": "Unsupported database type",
		})
	}
}

// UpdateDocument updates a document in a collection
func (h *DatabaseManagementHandler) UpdateDocument(c *fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	collectionName := c.Params("collection")
	documentID := c.Params("id")
	
	var req struct {
		DatabaseID string                 `json:"database_id"`
		Data       map[string]interface{} `json:"data"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	databaseID, err := uuid.Parse(req.DatabaseID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid database_id",
		})
	}

	// Get database connection info using GORM
	var connection models.DatabaseConnection
	if err := config.DB.Where("id = ? AND user_id = ?", databaseID, userID).First(&connection).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	switch connection.Type {
	case "mongodb":
		// Connect to MongoDB
		client, err := h.dbService.ConnectMongoDB(connection)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to connect to MongoDB: " + err.Error(),
			})
		}
		defer client.Disconnect(context.Background())

		collection := client.Database(connection.Database).Collection(collectionName)
		
		// Convert string ID to ObjectID
		objID, err := primitive.ObjectIDFromHex(documentID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid document ID",
			})
		}

		// Add timestamp
		req.Data["updated_at"] = time.Now()
		
		filter := bson.D{{"_id", objID}}
		update := bson.D{{"$set", req.Data}}
		
		result, err := collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to update document: " + err.Error(),
			})
		}

		if result.MatchedCount == 0 {
			return c.Status(404).JSON(fiber.Map{
				"error": "Document not found",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"modified": result.ModifiedCount,
		})

	case "mysql", "postgresql":
		// For SQL databases
		sqlClient, err := h.dbService.ConnectSQL(connection)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to connect to database: " + err.Error(),
			})
		}
		defer sqlClient.Close()

		// Build UPDATE query
		var setPairs []string
		var values []interface{}
		
		for key, value := range req.Data {
			setPairs = append(setPairs, key+" = ?")
			values = append(values, value)
		}
		
		// Add ID to values
		values = append(values, documentID)

		query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?",
			collectionName,
			strings.Join(setPairs, ", "))

		result, err := sqlClient.Exec(query, values...)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to update record: " + err.Error(),
			})
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return c.Status(404).JSON(fiber.Map{
				"error": "Record not found",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"modified": rowsAffected,
		})

	default:
		return c.Status(400).JSON(fiber.Map{
			"error": "Unsupported database type",
		})
	}
}

// DeleteDocument deletes a document from a collection
func (h *DatabaseManagementHandler) DeleteDocument(c *fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	collectionName := c.Params("collection")
	documentID := c.Params("id")
	
	var req struct {
		DatabaseID string `json:"database_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	databaseID, err := uuid.Parse(req.DatabaseID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid database_id",
		})
	}

	// Get database connection info using GORM
	var connection models.DatabaseConnection
	if err := config.DB.Where("id = ? AND user_id = ?", databaseID, userID).First(&connection).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	switch connection.Type {
	case "mongodb":
		// Connect to MongoDB
		client, err := h.dbService.ConnectMongoDB(connection)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to connect to MongoDB: " + err.Error(),
			})
		}
		defer client.Disconnect(context.Background())

		collection := client.Database(connection.Database).Collection(collectionName)
		
		// Convert string ID to ObjectID
		objID, err := primitive.ObjectIDFromHex(documentID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid document ID",
			})
		}

		filter := bson.D{{"_id", objID}}
		
		result, err := collection.DeleteOne(context.Background(), filter)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to delete document: " + err.Error(),
			})
		}

		if result.DeletedCount == 0 {
			return c.Status(404).JSON(fiber.Map{
				"error": "Document not found",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"deleted": result.DeletedCount,
		})

	case "mysql", "postgresql":
		// For SQL databases
		sqlClient, err := h.dbService.ConnectSQL(connection)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to connect to database: " + err.Error(),
			})
		}
		defer sqlClient.Close()

		query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", collectionName)
		result, err := sqlClient.Exec(query, documentID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to delete record: " + err.Error(),
			})
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return c.Status(404).JSON(fiber.Map{
				"error": "Record not found",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"deleted": rowsAffected,
		})

	default:
		return c.Status(400).JSON(fiber.Map{
			"error": "Unsupported database type",
		})
	}
}
