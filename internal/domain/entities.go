package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username  string    `json:"username" gorm:"unique;not null"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Password  string    `json:"-" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Bucket struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name       string         `json:"name" gorm:"unique;not null"`
	UserID     uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	User       User           `json:"user" gorm:"foreignKey:UserID"`
	Public     bool           `json:"public" gorm:"default:false"`
	Versioning bool           `json:"versioning" gorm:"default:true"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

type Object struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Key         string         `json:"key" gorm:"not null"`
	BucketID    uuid.UUID      `json:"bucket_id" gorm:"type:uuid;not null"`
	Bucket      Bucket         `json:"bucket" gorm:"foreignKey:BucketID"`
	VersionID   string         `json:"version_id" gorm:"not null"`
	Size        int64          `json:"size"`
	ContentType string         `json:"content_type"`
	ETag        string         `json:"etag"`
	StoragePath string         `json:"storage_path"`
	IsLatest    bool           `json:"is_latest" gorm:"default:true"`
	Metadata    string         `json:"metadata" gorm:"type:jsonb"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type ObjectVersion struct {
	Object
	VersionNumber int `json:"version_number"`
}

// Request/Response DTOs
type CreateBucketRequest struct {
	Name       string `json:"name" binding:"required"`
	Public     bool   `json:"public"`
	Versioning bool   `json:"versioning"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type ListObjectsResponse struct {
	Objects    []Object `json:"objects"`
	TotalCount int64    `json:"total_count"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
}

type UploadObjectResponse struct {
	Object    Object `json:"object"`
	VersionID string `json:"version_id"`
}
