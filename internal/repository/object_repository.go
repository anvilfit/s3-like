package repository

import (
	"s3-like/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type objectRepository struct {
	db *gorm.DB
}

func NewObjectRepository(db *gorm.DB) domain.ObjectRepository {
	return &objectRepository{db: db}
}

func (r *objectRepository) Create(object *domain.Object) error {
	return r.db.Create(object).Error
}

func (r *objectRepository) GetByKey(bucketID uuid.UUID, key string) (*domain.Object, error) {
	var object domain.Object
	err := r.db.Where("bucket_id = ? AND key = ? AND is_latest = true", bucketID, key).First(&object).Error
	if err != nil {
		return nil, err
	}
	return &object, nil
}

func (r *objectRepository) GetByKeyAndVersion(bucketID uuid.UUID, key, versionID string) (*domain.Object, error) {
	var object domain.Object
	err := r.db.Where("bucket_id = ? AND key = ? AND version_id = ?", bucketID, key, versionID).First(&object).Error
	if err != nil {
		return nil, err
	}
	return &object, nil
}

func (r *objectRepository) GetVersions(bucketID uuid.UUID, key string) ([]domain.Object, error) {
	var objects []domain.Object
	err := r.db.Where("bucket_id = ? AND key = ?", bucketID, key).
		Order("created_at DESC").Find(&objects).Error
	return objects, err
}

func (r *objectRepository) List(bucketID uuid.UUID, prefix string, page, pageSize int) ([]domain.Object, int64, error) {
	var objects []domain.Object
	var total int64

	query := r.db.Where("bucket_id = ? AND is_latest = true", bucketID)
	if prefix != "" {
		query = query.Where("key LIKE ?", prefix+"%")
	}

	// Count total
	query.Model(&domain.Object{}).Count(&total)

	// Get paginated results
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("key").Find(&objects).Error

	return objects, total, err
}

func (r *objectRepository) Update(object *domain.Object) error {
	return r.db.Save(object).Error
}

func (r *objectRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Object{}, id).Error
}

func (r *objectRepository) MarkAsNotLatest(bucketID uuid.UUID, key string) error {
	return r.db.Model(&domain.Object{}).
		Where("bucket_id = ? AND key = ?", bucketID, key).
		Update("is_latest", false).Error
}
