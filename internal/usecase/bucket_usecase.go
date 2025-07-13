package usecase

import (
	"errors"
	"s3-like/internal/domain"

	"github.com/google/uuid"
)

type bucketUseCase struct {
	bucketRepo domain.BucketRepository
}

func NewBucketUseCase(bucketRepo domain.BucketRepository) domain.BucketUseCase {
	return &bucketUseCase{
		bucketRepo: bucketRepo,
	}
}

func (uc *bucketUseCase) CreateBucket(userID uuid.UUID, req *domain.CreateBucketRequest) (*domain.Bucket, error) {
	// Check if bucket already exists
	if _, err := uc.bucketRepo.GetByName(req.Name); err == nil {
		return nil, errors.New("bucket already exists")
	}

	bucket := &domain.Bucket{
		Name:       req.Name,
		UserID:     userID,
		Public:     req.Public,
		Versioning: req.Versioning,
	}

	if err := uc.bucketRepo.Create(bucket); err != nil {
		return nil, err
	}

	return bucket, nil
}

func (uc *bucketUseCase) GetBucket(userID *uuid.UUID, name string) (*domain.Bucket, error) {
	bucket, err := uc.bucketRepo.GetByName(name)
	if err != nil {
		return nil, err
	}

	// Check if user owns the bucket or if it's public
	if &bucket.UserID != userID && !bucket.Public {
		return nil, errors.New("access denied")
	}

	return bucket, nil
}

func (uc *bucketUseCase) ListBuckets(userID uuid.UUID) ([]domain.Bucket, error) {
	return uc.bucketRepo.GetByUserID(userID)
}

func (uc *bucketUseCase) DeleteBucket(userID uuid.UUID, name string) error {
	bucket, err := uc.bucketRepo.GetByName(name)
	if err != nil {
		return err
	}

	if bucket.UserID != userID {
		return errors.New("access denied")
	}

	return uc.bucketRepo.Delete(bucket.ID)
}
