package repository

import (
	"s3-like/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type bucketRepository struct {
	db *gorm.DB
}

func NewBucketRepository(db *gorm.DB) domain.BucketRepository {
	return &bucketRepository{db: db}
}

func (r *bucketRepository) Create(bucket *domain.Bucket) error {
	return r.db.Create(bucket).Error
}

func (r *bucketRepository) GetByName(name string) (*domain.Bucket, error) {
	var bucket domain.Bucket
	err := r.db.Preload("User").Where("name = ?", name).First(&bucket).Error
	if err != nil {
		return nil, err
	}
	return &bucket, nil
}

func (r *bucketRepository) GetByID(id uuid.UUID) (*domain.Bucket, error) {
	var bucket domain.Bucket
	err := r.db.Preload("User").Where("id = ?", id).First(&bucket).Error
	if err != nil {
		return nil, err
	}
	return &bucket, nil
}

func (r *bucketRepository) GetByUserID(userID uuid.UUID) ([]domain.Bucket, error) {
	var buckets []domain.Bucket
	err := r.db.Where("user_id = ?", userID).Find(&buckets).Error
	return buckets, err
}

func (r *bucketRepository) Update(bucket *domain.Bucket) error {
	return r.db.Save(bucket).Error
}

func (r *bucketRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Bucket{}, id).Error
}
