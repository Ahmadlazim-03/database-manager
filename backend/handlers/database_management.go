package handlers

import (
	"context"
	"database/sql"
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

// Helper function to determine input type based on field name and data type
func determineInputType(fieldName, dataType string) string {
	fieldNameLower := strings.ToLower(fieldName)
	dataTypeLower := strings.ToLower(dataType)
	
	// Check for image/photo fields first by name
	if strings.Contains(fieldNameLower, "photo") || 
	   strings.Contains(fieldNameLower, "image") || 
	   strings.Contains(fieldNameLower, "picture") ||
	   strings.Contains(fieldNameLower, "avatar") ||
	   strings.Contains(fieldNameLower, "thumbnail") ||
	   strings.Contains(fieldNameLower, "logo") ||
	   strings.Contains(fieldNameLower, "icon") {
		return "image"
	}
	
	// Check for email fields
	if strings.Contains(fieldNameLower, "email") {
		return "email"
	}
	
	// Check for URL fields
	if strings.Contains(fieldNameLower, "url") || strings.Contains(fieldNameLower, "link") {
		return "url"
	}
	
	// Check for password fields
	if strings.Contains(fieldNameLower, "password") || strings.Contains(fieldNameLower, "pass") {
		return "password"
	}
	
	// Check for phone fields
	if strings.Contains(fieldNameLower, "phone") || strings.Contains(fieldNameLower, "tel") {
		return "tel"
	}
	
	// Check for date/time fields
	if strings.Contains(fieldNameLower, "date") || strings.Contains(fieldNameLower, "time") ||
	   strings.Contains(dataTypeLower, "date") || strings.Contains(dataTypeLower, "time") ||
	   strings.Contains(dataTypeLower, "timestamp") {
		return "datetime-local"
	}
	
	// Check by data type
	if strings.Contains(dataTypeLower, "int") || strings.Contains(dataTypeLower, "number") ||
	   strings.Contains(dataTypeLower, "decimal") || strings.Contains(dataTypeLower, "float") ||
	   strings.Contains(dataTypeLower, "double") {
		return "number"
	}
	
	if strings.Contains(dataTypeLower, "bool") {
		return "checkbox"
	}
	
	if strings.Contains(dataTypeLower, "text") || strings.Contains(dataTypeLower, "longtext") ||
	   strings.Contains(dataTypeLower, "mediumtext") {
		return "textarea"
	}
	
	// Default to text
	return "text"
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
		log.Printf("Database connection not found: %v", err)
		return c.Status(404).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	log.Printf("Found connection: ID=%s, Type=%s, Host=%s, Database=%s", connection.ID, connection.Type, connection.Host, connection.Database)

	var collections []string

	switch connection.Type {
	case "mongodb":
		log.Printf("Processing MongoDB connection")
		// Connect to MongoDB
		client, err := h.dbService.ConnectMongoDB(connection)
		if err != nil {
			log.Printf("MongoDB connection failed: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to connect to MongoDB: " + err.Error(),
			})
		}
		defer client.Disconnect(context.Background())

		database := client.Database(connection.Database)
		collectionNames, err := database.ListCollectionNames(context.Background(), bson.D{})
		if err != nil {
			log.Printf("Failed to list MongoDB collections: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to list collections: " + err.Error(),
			})
		}
		collections = collectionNames

	case "mysql", "postgresql", "postgres":
		log.Printf("Processing SQL connection type: %s", connection.Type)
		// For SQL databases, get table names
		sqlClient, err := h.dbService.ConnectSQL(connection)
		if err != nil {
			log.Printf("SQL connection failed: %v", err)
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
		log.Printf("Executing query: %s", query)

		rows, err := sqlClient.Query(query)
		if err != nil {
			log.Printf("Query failed: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to list tables: " + err.Error(),
			})
		}
		defer rows.Close()

		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				log.Printf("Error scanning table name: %v", err)
				continue
			}
			collections = append(collections, tableName)
		}
		log.Printf("Found %d collections/tables", len(collections))

	default:
		log.Printf("Unsupported database type: %s", connection.Type)
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

	type FieldInfo struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}

	var fields []FieldInfo

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
		
		// Get a sample document to extract field names and types
		cursor, err := collection.Find(context.Background(), bson.D{}, options.Find().SetLimit(10))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to sample documents: " + err.Error(),
			})
		}
		defer cursor.Close(context.Background())

		fieldSet := make(map[string]string) // field name -> type
		for cursor.Next(context.Background()) {
			var doc bson.M
			if err := cursor.Decode(&doc); err != nil {
				continue
			}
			for key, value := range doc {
				if key != "_id" {
					fieldType := "text"
					// Try to determine field type from value
					switch value.(type) {
					case string:
						fieldType = "text"
						// Check if it looks like an image URL or file path
						if val, ok := value.(string); ok {
							if strings.Contains(strings.ToLower(key), "photo") || 
							   strings.Contains(strings.ToLower(key), "image") || 
							   strings.Contains(strings.ToLower(key), "picture") ||
							   strings.Contains(strings.ToLower(key), "avatar") ||
							   strings.Contains(strings.ToLower(key), "thumbnail") ||
							   strings.HasSuffix(strings.ToLower(val), ".jpg") ||
							   strings.HasSuffix(strings.ToLower(val), ".jpeg") ||
							   strings.HasSuffix(strings.ToLower(val), ".png") ||
							   strings.HasSuffix(strings.ToLower(val), ".gif") ||
							   strings.HasSuffix(strings.ToLower(val), ".webp") {
								fieldType = "image"
							}
						}
					case int, int32, int64:
						fieldType = "number"
					case float32, float64:
						fieldType = "number"
					case bool:
						fieldType = "boolean"
					default:
						fieldType = "text"
					}
					
					if _, exists := fieldSet[key]; !exists {
						fieldSet[key] = fieldType
					}
				}
			}
		}

		for fieldName, fieldType := range fieldSet {
			fields = append(fields, FieldInfo{Name: fieldName, Type: fieldType})
		}

	case "mysql", "postgresql", "postgres":
		// Get column names and types for SQL tables
		sqlClient, err := h.dbService.ConnectSQL(connection)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to connect to database: " + err.Error(),
			})
		}
		defer sqlClient.Close()

		var query string
		if connection.Type == "mysql" {
			// Get column name and data type for MySQL
			query = fmt.Sprintf("DESCRIBE %s", collectionName)
		} else {
			// PostgreSQL is case-sensitive, try both original and lowercase
			query = fmt.Sprintf(`
				SELECT column_name, data_type 
				FROM information_schema.columns 
				WHERE table_name = '%s' AND table_schema = 'public'
				ORDER BY ordinal_position
			`, collectionName)
			
			// Also try lowercase version
			queryLower := fmt.Sprintf(`
				SELECT column_name, data_type 
				FROM information_schema.columns 
				WHERE table_name = '%s' AND table_schema = 'public'
				ORDER BY ordinal_position
			`, strings.ToLower(collectionName))
			
			log.Printf("Will try both queries - original: %s, lowercase: %s", query, queryLower)
		}

		log.Printf("Executing schema query for table '%s': %s", collectionName, query)

		rows, err := sqlClient.Query(query)
		if err != nil {
			log.Printf("Query error for table '%s': %v", collectionName, err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to get table schema: " + err.Error(),
			})
		}
		defer rows.Close()

		// Check if we got any results, if not try lowercase for PostgreSQL
		fieldCount := 0
		if connection.Type == "mysql" {
			for rows.Next() {
				var field, fieldType, null, key, extra string
				var defaultVal sql.NullString // Use sql.NullString for nullable columns
				if err := rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra); err != nil {
					log.Printf("Error scanning MySQL row: %v", err)
					continue
				}
				log.Printf("Found MySQL field: %s, type: %s", field, fieldType)
				
				// Determine input type based on MySQL field type and name
				inputType := determineInputType(field, fieldType)
				fields = append(fields, FieldInfo{Name: field, Type: inputType})
				fieldCount++
			}
		} else {
			for rows.Next() {
				var field, dataType string
				if err := rows.Scan(&field, &dataType); err != nil {
					log.Printf("Error scanning PostgreSQL row: %v", err)
					continue
				}
				log.Printf("Found PostgreSQL field: %s, type: %s", field, dataType)
				
				// Determine input type based on PostgreSQL field type and name
				inputType := determineInputType(field, dataType)
				fields = append(fields, FieldInfo{Name: field, Type: inputType})
				fieldCount++
			}
		}

		// If no fields found and it's PostgreSQL, try lowercase table name
		if fieldCount == 0 && (connection.Type == "postgresql" || connection.Type == "postgres") {
			log.Printf("No fields found with original name, trying lowercase for PostgreSQL")
			rows.Close() // Close previous rows
			
			queryLower := fmt.Sprintf(`
				SELECT column_name, data_type 
				FROM information_schema.columns 
				WHERE table_name = '%s' AND table_schema = 'public'
				ORDER BY ordinal_position
			`, strings.ToLower(collectionName))
			
			log.Printf("Executing lowercase query: %s", queryLower)
			
			rows, err = sqlClient.Query(queryLower)
			if err != nil {
				log.Printf("Lowercase query error: %v", err)
			} else {
				defer rows.Close()
				for rows.Next() {
					var field, dataType string
					if err := rows.Scan(&field, &dataType); err != nil {
						log.Printf("Error scanning PostgreSQL lowercase row: %v", err)
						continue
					}
					log.Printf("Found PostgreSQL field (lowercase): %s, type: %s", field, dataType)
					
					// Determine input type based on PostgreSQL field type and name
					inputType := determineInputType(field, dataType)
					fields = append(fields, FieldInfo{Name: field, Type: inputType})
				}
			}
		}
	}

	// Log the fields before returning
	log.Printf("GetCollectionSchema for table '%s': found fields: %v", collectionName, fields)

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

	case "mysql", "postgresql", "postgres":
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
		
		log.Printf("GetDocuments - Initial query for collection '%s': %s", collectionName, baseQuery)
		
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
						var fieldType, null, key, extra string
						var defaultVal sql.NullString
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
					if connection.Type == "mysql" {
						searchConditions = append(searchConditions, fmt.Sprintf("CAST(%s AS CHAR) LIKE ?", col))
					} else {
						searchConditions = append(searchConditions, fmt.Sprintf("CAST(%s AS TEXT) ILIKE ?", col))
					}
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
		
		log.Printf("GetDocuments - Count query: %s, args: %v", countQuery, countArgs)
		err = sqlClient.QueryRow(countQuery, countArgs...).Scan(&total)
		if err != nil {
			log.Printf("Error counting rows: %v", err)
		}
		log.Printf("GetDocuments - Total count: %d", total)

		// Add sorting and pagination
		if sortField != "" && sortOrder != "" {
			baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortField, strings.ToUpper(sortOrder))
		}
		
		offset := (page - 1) * limit
		baseQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

		log.Printf("GetDocuments - Final query: %s, args: %v", baseQuery, args)
		rows, err := sqlClient.Query(baseQuery, args...)
		if err != nil {
			log.Printf("GetDocuments - Query error: %v", err)
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
		
		log.Printf("GetDocuments - Found %d documents for collection '%s'", len(documents), collectionName)
	}

	log.Printf("GetDocuments - Returning: documents=%d, total=%d, page=%d, limit=%d", len(documents), total, page, limit)
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

	log.Printf("Received request for collection: %s", collectionName)
	
	// Check request body size before parsing
	bodySize := len(c.Body())
	maxBodySize := 5 * 1024 * 1024 // 5MB limit
	if bodySize > maxBodySize {
		log.Printf("Request body too large: %d bytes (max: %d bytes)", bodySize, maxBodySize)
		return c.Status(413).JSON(fiber.Map{
			"error": "Request body too large. Maximum size is 5MB.",
		})
	}
	
	log.Printf("Request body size: %d bytes", bodySize)

	if err := c.BodyParser(&req); err != nil {
		log.Printf("Body parser error: %v", err)
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	// Validate individual field sizes, especially for image data
	for fieldName, fieldValue := range req.Data {
		if valueStr, ok := fieldValue.(string); ok {
			// Check if this might be base64 image data
			if strings.HasPrefix(valueStr, "data:image/") && len(valueStr) > 1024*1024 { // 1MB limit per image
				log.Printf("Image field %s is too large: %d characters", fieldName, len(valueStr))
				return c.Status(413).JSON(fiber.Map{
					"error": fmt.Sprintf("Image in field '%s' is too large. Maximum size per image is 1MB.", fieldName),
				})
			}
		}
	}

	log.Printf("Parsed request - DatabaseID: %s, Data fields: %d", req.DatabaseID, len(req.Data))

	databaseID, err := uuid.Parse(req.DatabaseID)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid database_id",
		})
	}

	log.Printf("Looking for database connection with ID: %s and userID: %s", databaseID, userID)

	// Get database connection info using GORM
	var connection models.DatabaseConnection
	if err := config.DB.Where("id = ? AND user_id = ?", databaseID, userID).First(&connection).Error; err != nil {
		log.Printf("Database connection not found error: %v", err)
		return c.Status(404).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	log.Printf("Found connection: %+v", connection)

	switch connection.Type {
	case "mongodb":
		log.Printf("Processing MongoDB database type: %s", connection.Type)
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

	case "mysql", "postgresql", "postgres":
		log.Printf("Processing SQL database type: %s", connection.Type)
		// For SQL databases, convert map to INSERT statement
		sqlClient, err := h.dbService.ConnectSQL(connection)
		if err != nil {
			log.Printf("SQL connection error: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to connect to database: " + err.Error(),
			})
		}
		defer sqlClient.Close()

		// Build INSERT query
		var columns []string
		var placeholders []string
		var values []interface{}
		
		i := 1
		for key, value := range req.Data {
			columns = append(columns, key)
			
			// Use different placeholder format based on database type
			if connection.Type == "postgresql" || connection.Type == "postgres" {
				placeholders = append(placeholders, fmt.Sprintf("$%d", i))
				i++
			} else {
				placeholders = append(placeholders, "?")
			}
			
			values = append(values, value)
		}

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			collectionName,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "))

		log.Printf("Executing SQL query: %s", query)
		log.Printf("With values: %+v", values)

		result, err := sqlClient.Exec(query, values...)
		if err != nil {
			log.Printf("SQL execution error: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to create record: " + err.Error(),
			})
		}

		id, _ := result.LastInsertId()
		log.Printf("Successfully created record with ID: %d", id)
		return c.JSON(fiber.Map{
			"success": true,
			"id":      id,
		})

	default:
		log.Printf("Unsupported database type: %s", connection.Type)
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

	case "mysql", "postgresql", "postgres":
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
		
		i := 1
		for key, value := range req.Data {
			if connection.Type == "postgresql" || connection.Type == "postgres" {
				setPairs = append(setPairs, fmt.Sprintf("%s = $%d", key, i))
				i++
			} else {
				setPairs = append(setPairs, key+" = ?")
			}
			values = append(values, value)
		}
		
		// Add ID to values
		values = append(values, documentID)

		var query string
		if connection.Type == "postgresql" || connection.Type == "postgres" {
			query = fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d",
				collectionName,
				strings.Join(setPairs, ", "),
				i)
		} else {
			query = fmt.Sprintf("UPDATE %s SET %s WHERE id = ?",
				collectionName,
				strings.Join(setPairs, ", "))
		}

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

	case "mysql", "postgresql", "postgres":
		// For SQL databases
		sqlClient, err := h.dbService.ConnectSQL(connection)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to connect to database: " + err.Error(),
			})
		}
		defer sqlClient.Close()

		var query string
		if connection.Type == "postgresql" || connection.Type == "postgres" {
			query = fmt.Sprintf("DELETE FROM %s WHERE id = $1", collectionName)
		} else {
			query = fmt.Sprintf("DELETE FROM %s WHERE id = ?", collectionName)
		}
		
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
