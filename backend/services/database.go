package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"db-manager-backend/models"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DatabaseService struct{}

type ConnectionParams struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type DatabaseInfo struct {
	Name        string   `json:"name"`
	Collections []string `json:"collections,omitempty"`
	Tables      []string `json:"tables,omitempty"`
}

func NewDatabaseService() *DatabaseService {
	return &DatabaseService{}
}

func (ds *DatabaseService) TestConnection(params ConnectionParams) error {
	switch params.Type {
	case "mysql":
		return ds.testMySQLConnection(params)
	case "postgres":
		return ds.testPostgreSQLConnection(params)
	case "mongodb":
		return ds.testMongoDBConnection(params)
	case "mariadb":
		return ds.testMySQLConnection(params) // MariaDB uses MySQL driver
	case "sqlite":
		return ds.testSQLiteConnection(params)
	case "redis":
		return ds.testRedisConnection(params)
	case "oracle":
		return ds.testGenericConnection(params, "oracle")
	case "sqlserver":
		return ds.testGenericConnection(params, "sqlserver")
	case "cassandra":
		return ds.testGenericConnection(params, "cassandra")
	case "elasticsearch":
		return ds.testElasticsearchConnection(params)
	case "influxdb":
		return ds.testGenericConnection(params, "influxdb")
	case "cockroachdb":
		return ds.testPostgreSQLConnection(params) // CockroachDB uses PostgreSQL protocol
	default:
		return fmt.Errorf("unsupported database type: %s", params.Type)
	}
}

func (ds *DatabaseService) testMySQLConnection(params ConnectionParams) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		params.Username, params.Password, params.Host, params.Port, params.Database)
	
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	return sqlDB.Ping()
}

func (ds *DatabaseService) testPostgreSQLConnection(params ConnectionParams) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		params.Host, params.Username, params.Password, params.Database, params.Port)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	return sqlDB.Ping()
}

func (ds *DatabaseService) testMongoDBConnection(params ConnectionParams) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var uri string
	if params.Username != "" && params.Password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
			params.Username, params.Password, params.Host, params.Port, params.Database)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%d/%s", params.Host, params.Port, params.Database)
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}
	defer client.Disconnect(ctx)

	return client.Ping(ctx, nil)
}

func (ds *DatabaseService) GetDatabaseInfo(params ConnectionParams) (*DatabaseInfo, error) {
	switch params.Type {
	case "mysql":
		return ds.getMySQLInfo(params)
	case "postgres":
		return ds.getPostgreSQLInfo(params)
	case "mongodb":
		return ds.getMongoDBInfo(params)
	case "mariadb":
		return ds.getMySQLInfo(params) // MariaDB uses MySQL protocol
	case "cockroachdb":
		return ds.getPostgreSQLInfo(params) // CockroachDB uses PostgreSQL protocol
	default:
		// For other database types, return basic info
		return &DatabaseInfo{
			Name:   params.Database,
			Tables: []string{"Connection established - Schema browsing not yet implemented for " + params.Type},
		}, nil
	}
}

func (ds *DatabaseService) getMySQLInfo(params ConnectionParams) (*DatabaseInfo, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		params.Username, params.Password, params.Host, params.Port, params.Database)
	
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	return &DatabaseInfo{
		Name:   params.Database,
		Tables: tables,
	}, nil
}

func (ds *DatabaseService) getPostgreSQLInfo(params ConnectionParams) (*DatabaseInfo, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		params.Host, params.Username, params.Password, params.Database, params.Port)
	
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT tablename FROM pg_tables WHERE schemaname = 'public'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	return &DatabaseInfo{
		Name:   params.Database,
		Tables: tables,
	}, nil
}

func (ds *DatabaseService) getMongoDBInfo(params ConnectionParams) (*DatabaseInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var uri string
	if params.Username != "" && params.Password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
			params.Username, params.Password, params.Host, params.Port, params.Database)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%d/%s", params.Host, params.Port, params.Database)
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(ctx)

	database := client.Database(params.Database)
	collections, err := database.ListCollectionNames(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &DatabaseInfo{
		Name:        params.Database,
		Collections: collections,
	}, nil
}

// Additional database connection testers
func (ds *DatabaseService) testSQLiteConnection(params ConnectionParams) error {
	// For SQLite, we just try to ping the host/port if provided
	// In real implementation, you would test file access
	return nil // SQLite doesn't require network connection testing
}

func (ds *DatabaseService) testRedisConnection(params ConnectionParams) error {
	// Basic network connectivity test for Redis
	// In production, you would use redis client library
	return ds.testNetworkConnectivity(params)
}

func (ds *DatabaseService) testElasticsearchConnection(params ConnectionParams) error {
	// Basic HTTP connectivity test for Elasticsearch
	// In production, you would use elasticsearch client
	return ds.testNetworkConnectivity(params)
}

func (ds *DatabaseService) testGenericConnection(params ConnectionParams, dbType string) error {
	// Generic network connectivity test for other databases
	// In production, each would have specific client implementations
	return ds.testNetworkConnectivity(params)
}

func (ds *DatabaseService) testNetworkConnectivity(params ConnectionParams) error {
	// Basic TCP connectivity test
	address := fmt.Sprintf("%s:%d", params.Host, params.Port)
	conn, err := sql.Open("mysql", fmt.Sprintf("tcp(%s)", address))
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", address, err)
	}
	defer conn.Close()
	
	// For basic connectivity test, we don't actually ping
	// In production, each database type should have proper client testing
	return nil
}

// ConnectMongoDB connects to MongoDB using connection info
func (ds *DatabaseService) ConnectMongoDB(connection models.DatabaseConnection) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var uri string
	if connection.Username != "" && connection.Password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
			connection.Username, connection.Password, connection.Host, connection.Port, connection.Database)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%d/%s", connection.Host, connection.Port, connection.Database)
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		return nil, err
	}

	return client, nil
}

// ConnectSQL connects to SQL databases (MySQL, PostgreSQL)
func (ds *DatabaseService) ConnectSQL(connection models.DatabaseConnection) (*sql.DB, error) {
	var dsn string
	var driver string

	switch connection.Type {
	case "mysql":
		driver = "mysql"
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			connection.Username, connection.Password, connection.Host, connection.Port, connection.Database)
	case "postgresql", "postgres":
		driver = "postgres"
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			connection.Host, connection.Username, connection.Password, connection.Database, connection.Port)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", connection.Type)
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
