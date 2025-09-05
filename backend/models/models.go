package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID      `json:"id" gorm:"type:char(36);primaryKey"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type DatabaseConnection struct {
	ID           uuid.UUID      `json:"id" gorm:"type:char(36);primaryKey"`
	UserID       uuid.UUID      `json:"user_id" gorm:"type:char(36);not null"`
	Name         string         `json:"name" gorm:"not null"`
	Type         string         `json:"type" gorm:"not null"` // mysql, mongodb, postgres
	Host         string         `json:"host" gorm:"not null"`
	Port         int            `json:"port" gorm:"not null"`
	Database     string         `json:"database" gorm:"not null"`
	Username     string         `json:"username"`
	Password     string         `json:"-"`
	Status       string         `json:"status" gorm:"default:'active'"` // active, inactive
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	User         User           `json:"user" gorm:"foreignKey:UserID"`
}

type APIKey struct {
	ID           uuid.UUID         `json:"id" gorm:"type:char(36);primaryKey"`
	UserID       uuid.UUID         `json:"user_id" gorm:"type:char(36);not null"`
	DatabaseID   uuid.UUID         `json:"database_id" gorm:"type:char(36);not null"`
	Name         string            `json:"name" gorm:"not null"`
	Key          string            `json:"key" gorm:"uniqueIndex;not null"`
	IsActive     bool              `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	DeletedAt    gorm.DeletedAt    `json:"-" gorm:"index"`
	User         User              `json:"user" gorm:"foreignKey:UserID"`
	Database     DatabaseConnection `json:"database" gorm:"foreignKey:DatabaseID"`
}

type APIEndpoint struct {
	ID           uuid.UUID         `json:"id" gorm:"type:char(36);primaryKey"`
	DatabaseID   uuid.UUID         `json:"database_id" gorm:"type:char(36);not null"`
	Collection   string            `json:"collection" gorm:"not null"`
	Path         string            `json:"path" gorm:"not null"`
	Method       string            `json:"method" gorm:"not null"` // GET, POST, PUT, DELETE
	IsActive     bool              `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	DeletedAt    gorm.DeletedAt    `json:"-" gorm:"index"`
	Database     DatabaseConnection `json:"database" gorm:"foreignKey:DatabaseID"`
}

type APILog struct {
	ID         uuid.UUID `json:"id" gorm:"type:char(36);primaryKey"`
	APIKeyID   uuid.UUID `json:"api_key_id" gorm:"type:char(36);not null"`
	EndpointID uuid.UUID `json:"endpoint_id" gorm:"type:char(36);not null"`
	Method     string    `json:"method" gorm:"not null"`
	Path       string    `json:"path" gorm:"not null"`
	StatusCode int       `json:"status_code" gorm:"not null"`
	ResponseTime int64   `json:"response_time"` // in milliseconds
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
	APIKey     APIKey    `json:"api_key" gorm:"foreignKey:APIKeyID"`
	Endpoint   APIEndpoint `json:"endpoint" gorm:"foreignKey:EndpointID"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}

func (dc *DatabaseConnection) BeforeCreate(tx *gorm.DB) error {
	dc.ID = uuid.New()
	return nil
}

func (ak *APIKey) BeforeCreate(tx *gorm.DB) error {
	ak.ID = uuid.New()
	return nil
}

func (ae *APIEndpoint) BeforeCreate(tx *gorm.DB) error {
	ae.ID = uuid.New()
	return nil
}

func (al *APILog) BeforeCreate(tx *gorm.DB) error {
	al.ID = uuid.New()
	return nil
}
