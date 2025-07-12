package usecase

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"s3-like/internal/domain"
	"time"

	"github.com/google/uuid"
)

type objectUseCase struct {
	objectRepo domain.ObjectRepository
	basePath   string
}

func NewObjectUseCase(objectRepo domain.ObjectRepository, basePath string) domain.ObjectUseCase {
	return &objectUseCase{
		objectRepo: objectRepo,
		basePath:   basePath,
	}
}

func (uc *objectUseCase) UploadObject(bucketID uuid.UUID, key string, file multipart.File, header *multipart.FileHeader, metadata map[string]string) (*domain.UploadObjectResponse, error) {
	// Generate version ID
	versionID := uuid.New().String()

	// Create storage path
	storagePath := filepath.Join(uc.basePath, bucketID.String(), key, versionID)
	if err := os.MkdirAll(filepath.Dir(storagePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Create file on disk
	dst, err := os.Create(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy file content and calculate hash
	hasher := md5.New()
	multiWriter := io.MultiWriter(dst, hasher)

	size, err := io.Copy(multiWriter, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	etag := fmt.Sprintf("%x", hasher.Sum(nil))

	// Prepare metadata JSON
	var metadataJSON string
	if metadata != nil && len(metadata) > 0 {
		// Add system metadata
		systemMetadata := make(map[string]interface{})
		for k, v := range metadata {
			systemMetadata[k] = v
		}

		// Add automatic metadata
		systemMetadata["upload_time"] = time.Now().UTC().Format(time.RFC3339)
		systemMetadata["original_filename"] = header.Filename
		systemMetadata["file_size"] = size
		systemMetadata["content_type"] = header.Header.Get("Content-Type")

		metadataBytes, err := json.Marshal(systemMetadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = string(metadataBytes)
	} else {
		// Default metadata if none provided
		defaultMetadata := map[string]interface{}{
			"upload_time":       time.Now().UTC().Format(time.RFC3339),
			"original_filename": header.Filename,
			"file_size":         size,
			"content_type":      header.Header.Get("Content-Type"),
		}
		metadataBytes, err := json.Marshal(defaultMetadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal default metadata: %w", err)
		}
		metadataJSON = string(metadataBytes)
	}

	// Determine content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Mark previous versions as not latest
	uc.objectRepo.MarkAsNotLatest(bucketID, key)

	// Create object record
	object := &domain.Object{
		Key:         key,
		BucketID:    bucketID,
		VersionID:   versionID,
		Size:        size,
		ContentType: contentType,
		ETag:        etag,
		StoragePath: storagePath,
		IsLatest:    true,
		Metadata:    metadataJSON,
	}

	if err := uc.objectRepo.Create(object); err != nil {
		// Clean up file if database operation fails
		os.Remove(storagePath)
		return nil, err
	}

	return &domain.UploadObjectResponse{
		Object:    *object,
		VersionID: versionID,
	}, nil
}

func (uc *objectUseCase) GetObject(bucketID uuid.UUID, key string) (*domain.Object, io.ReadCloser, error) {
	object, err := uc.objectRepo.GetByKey(bucketID, key)
	if err != nil {
		return nil, nil, err
	}

	file, err := os.Open(object.StoragePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	return object, file, nil
}

func (uc *objectUseCase) GetObjectVersion(bucketID uuid.UUID, key, versionID string) (*domain.Object, io.ReadCloser, error) {
	object, err := uc.objectRepo.GetByKeyAndVersion(bucketID, key, versionID)
	if err != nil {
		return nil, nil, err
	}

	file, err := os.Open(object.StoragePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	return object, file, nil
}

func (uc *objectUseCase) ListObjects(bucketID uuid.UUID, prefix string, page, pageSize int) (*domain.ListObjectsResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}

	objects, total, err := uc.objectRepo.List(bucketID, prefix, page, pageSize)
	if err != nil {
		return nil, err
	}

	return &domain.ListObjectsResponse{
		Objects:    objects,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (uc *objectUseCase) ListObjectVersions(bucketID uuid.UUID, key string) ([]domain.Object, error) {
	return uc.objectRepo.GetVersions(bucketID, key)
}

func (uc *objectUseCase) DeleteObject(bucketID uuid.UUID, key string) error {
	object, err := uc.objectRepo.GetByKey(bucketID, key)
	if err != nil {
		return err
	}

	// Delete file from storage
	if err := os.Remove(object.StoragePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Delete from database
	return uc.objectRepo.Delete(object.ID)
}
