package domain

import (
	"io"
	"mime/multipart"

	"github.com/google/uuid"
)

// Repository interfaces
type UserRepository interface {
	Create(user *User) error
	GetByUsername(username string) (*User, error)
	GetByID(id uuid.UUID) (*User, error)
}

type BucketRepository interface {
	Create(bucket *Bucket) error
	GetByName(name string) (*Bucket, error)
	GetByID(id uuid.UUID) (*Bucket, error)
	GetByUserID(userID uuid.UUID) ([]Bucket, error)
	Update(bucket *Bucket) error
	Delete(id uuid.UUID) error
}

type ObjectRepository interface {
	Create(object *Object) error
	GetByKey(bucketID uuid.UUID, key string) (*Object, error)
	GetByKeyAndVersion(bucketID uuid.UUID, key, versionID string) (*Object, error)
	GetVersions(bucketID uuid.UUID, key string) ([]Object, error)
	List(bucketID uuid.UUID, prefix string, page, pageSize int) ([]Object, int64, error)
	Update(object *Object) error
	Delete(id uuid.UUID) error
	MarkAsNotLatest(bucketID uuid.UUID, key string) error
}

// Use case interfaces
type AuthUseCase interface {
	Login(username, password string) (*AuthResponse, error)
	Register(req *RegisterRequest) (*AuthResponse, error)
}

type BucketUseCase interface {
	CreateBucket(userID uuid.UUID, req *CreateBucketRequest) (*Bucket, error)
	GetBucket(userID *uuid.UUID, name string) (*Bucket, error)
	ListBuckets(userID uuid.UUID) ([]Bucket, error)
	DeleteBucket(userID uuid.UUID, name string) error
}

type ObjectUseCase interface {
	UploadObject(bucketID uuid.UUID, key string, file multipart.File, header *multipart.FileHeader, metadata map[string]string) (*UploadObjectResponse, error)
	GetObject(bucketID uuid.UUID, key string) (*Object, io.ReadCloser, error)
	GetObjectVersion(bucketID uuid.UUID, key, versionID string) (*Object, io.ReadCloser, error)
	ListObjects(bucketID uuid.UUID, prefix string, page, pageSize int) (*ListObjectsResponse, error)
	ListObjectVersions(bucketID uuid.UUID, key string) ([]Object, error)
	DeleteObject(bucketID uuid.UUID, key string) error
}
