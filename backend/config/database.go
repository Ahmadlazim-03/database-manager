package config

import (
	"fmt"
	"log"
	"os"
	"sync"

	"db-manager-backend/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// In-memory storage for development (when PostgreSQL is not available)
type InMemoryDB struct {
	Users       map[string]*models.User
	Connections map[string]*models.DatabaseConnection
	APIKeys     map[string]*models.APIKey
	Endpoints   map[string]*models.APIEndpoint
	Logs        map[string]*models.APILog
	mutex       sync.RWMutex
}

var memDB *InMemoryDB

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found, using default values")
	}
}

func ConnectDB() {
	var err error
	
	// Try to connect to PostgreSQL first
	host := GetEnv("DB_HOST", "localhost")
	port := GetEnv("DB_PORT", "5432")
	dbname := GetEnv("DB_NAME", "postgres")
	username := GetEnv("DB_USER", "postgres")
	password := GetEnv("DB_PASSWORD", "")
	
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		host, username, password, dbname, port)
	
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("PostgreSQL connection failed: %v", err)
		log.Println("Falling back to in-memory database for development...")
		log.Println("For production, please set up PostgreSQL and configure environment variables:")
		log.Println("DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD")
		
		// Initialize in-memory database
		initInMemoryDB()
		fmt.Println("In-memory database initialized for development")
		return
	}

	// Auto migrate the schema for PostgreSQL
	err = DB.AutoMigrate(
		&models.User{},
		&models.DatabaseConnection{},
		&models.APIKey{},
		&models.APIEndpoint{},
		&models.APILog{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	fmt.Println("PostgreSQL database connected and migrated successfully")
}

func initInMemoryDB() {
	memDB = &InMemoryDB{
		Users:       make(map[string]*models.User),
		Connections: make(map[string]*models.DatabaseConnection),
		APIKeys:     make(map[string]*models.APIKey),
		Endpoints:   make(map[string]*models.APIEndpoint),
		Logs:        make(map[string]*models.APILog),
	}
}

func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Helper function to check if using in-memory DB
func IsInMemoryDB() bool {
	return DB == nil && memDB != nil
}
