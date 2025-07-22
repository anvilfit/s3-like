package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"s3-like/internal/domain"
	"s3-like/internal/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ObjectHandler struct {
	objectUseCase domain.ObjectUseCase
	bucketUseCase domain.BucketUseCase
}

func NewObjectHandler(objectUseCase domain.ObjectUseCase, bucketUseCase domain.BucketUseCase) *ObjectHandler {
	return &ObjectHandler{
		objectUseCase: objectUseCase,
		bucketUseCase: bucketUseCase,
	}
}

// UploadObject godoc
// @Summary Upload a file
// @Description Upload a file to a specific bucket with optional metadata
// @Tags objects
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param bucket path string true "Bucket name"
// @Param file formData file true "File to upload"
// @Param key formData string false "Object key (if not provided, filename will be used)"
// @Param metadata formData string false "JSON string with custom metadata (e.g., {\"author\":\"John\",\"category\":\"documents\"})"
// @Param content-type formData string false "Content type override"
// @Param description formData string false "File description"
// @Param tags formData string false "Comma-separated tags"
// @Success 201 {object} domain.UploadObjectResponse "File uploaded successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Bucket not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/buckets/{bucket}/objects [post]
func (h *ObjectHandler) UploadObject(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	bucketName := c.Param("bucket")

	// Get bucket
	bucket, err := h.bucketUseCase.GetBucket(&userID, bucketName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bucket not found"})
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	// Get key from form or use filename
	key := c.PostForm("key")
	if key == "" {
		key = utils.SanitizeFilename(header.Filename)
	}

	// Parse metadata from form
	metadata := make(map[string]string)

	// Get JSON metadata if provided
	if metadataJSON := c.PostForm("metadata"); metadataJSON != "" {
		var jsonMetadata map[string]any
		if err := json.Unmarshal([]byte(metadataJSON), &jsonMetadata); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid metadata JSON format"})
			return
		}
		// Convert to string map
		for k, v := range jsonMetadata {
			metadata[k] = fmt.Sprintf("%v", v)
		}
	}

	// Get individual metadata fields
	if description := c.PostForm("description"); description != "" {
		metadata["description"] = description
	}

	if tags := c.PostForm("tags"); tags != "" {
		metadata["tags"] = tags
	}

	if contentTypeOverride := c.PostForm("content-type"); contentTypeOverride != "" {
		metadata["content_type_override"] = contentTypeOverride
		// Override the header content type
		header.Header.Set("Content-Type", contentTypeOverride)
	}

	// Add user information to metadata
	if user, exists := c.Get("user"); exists {
		if userObj, ok := user.(*domain.User); ok {
			metadata["uploaded_by"] = userObj.Username
			metadata["uploaded_by_id"] = userObj.ID.String()
		}
	}

	// Add request metadata
	metadata["user_agent"] = c.GetHeader("User-Agent")
	metadata["client_ip"] = c.ClientIP()

	response, err := h.objectUseCase.UploadObject(bucket.ID, key, file, header, metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetObject godoc
// @Summary Download a file
// @Description Download the latest version of a file from a bucket
// @Tags objects
// @Accept json
// @Produce application/octet-stream
// @Security BearerAuth
// @Param bucket path string true "Bucket name"
// @Param key path string true "Object key"
// @Success 200 {file} file "File content"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Object or bucket not found"
// @Router /api/v1/buckets/{bucket}/objects/{key} [get]
func (h *ObjectHandler) GetObject(c *gin.Context) {
	userID, userExists := c.Get("user_id")
	bucketName := c.Param("bucket")
	key := strings.TrimPrefix(c.Param("key"), "/")

	var userIDuuid *uuid.UUID

	if userExists {
		if strID, ok := userID.(string); ok && strID != "" {
			parsedID, _ := uuid.Parse(strID)
			userIDuuid = &parsedID
		}
	}

	// Get bucket
	bucket, err := h.bucketUseCase.GetBucket(userIDuuid, bucketName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bucket not found"})
		return
	}

	object, file, err := h.objectUseCase.GetObject(bucket.ID, key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "object not found"})
		return
	}
	defer file.Close()

	// Set headers
	c.Header("Content-Type", object.ContentType)
	c.Header("Content-Length", strconv.FormatInt(object.Size, 10))
	c.Header("ETag", object.ETag)
	c.Header("Content-Disposition", "attachment; filename=\""+object.Key+"\"")
	c.Header("X-Object-Version-ID", object.VersionID)

	// Add metadata to headers if available
	if object.Metadata != "" {
		var metadata map[string]any
		if err := json.Unmarshal([]byte(object.Metadata), &metadata); err == nil {
			for k, v := range metadata {
				headerKey := "X-Object-Meta-" + strings.ReplaceAll(k, "_", "-")
				c.Header(headerKey, fmt.Sprintf("%v", v))
			}
		}
	}

	// Stream file
	c.DataFromReader(http.StatusOK, object.Size, object.ContentType, file, nil)
}

// GetObjectVersion godoc
// @Summary Download a specific version of a file
// @Description Download a specific version of a file from a bucket
// @Tags objects
// @Accept json
// @Produce application/octet-stream
// @Security BearerAuth
// @Param bucket path string true "Bucket name"
// @Param key path string true "Object key"
// @Param version path string true "Version ID"
// @Success 200 {file} file "File content"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Object version or bucket not found"
// @Router /api/v1/buckets/{bucket}/objects/{key}/versions/{version} [get]
func (h *ObjectHandler) GetObjectVersion(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	bucketName := c.Param("bucket")
	key := strings.TrimPrefix(c.Param("key"), "/")
	versionID := c.Param("version")

	// Get bucket
	bucket, err := h.bucketUseCase.GetBucket(&userID, bucketName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bucket not found"})
		return
	}

	object, file, err := h.objectUseCase.GetObjectVersion(bucket.ID, key, versionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "object version not found"})
		return
	}
	defer file.Close()

	// Set headers
	c.Header("Content-Type", object.ContentType)
	c.Header("Content-Length", strconv.FormatInt(object.Size, 10))
	c.Header("ETag", object.ETag)
	c.Header("Content-Disposition", "attachment; filename=\""+object.Key+"\"")
	c.Header("X-Object-Version-ID", object.VersionID)

	// Add metadata to headers if available
	if object.Metadata != "" {
		var metadata map[string]any
		if err := json.Unmarshal([]byte(object.Metadata), &metadata); err == nil {
			for k, v := range metadata {
				headerKey := "X-Object-Meta-" + strings.ReplaceAll(k, "_", "-")
				c.Header(headerKey, fmt.Sprintf("%v", v))
			}
		}
	}

	// Stream file
	c.DataFromReader(http.StatusOK, object.Size, object.ContentType, file, nil)
}

// ListObjects godoc
// @Summary List objects in a bucket
// @Description Get a paginated list of objects in a bucket
// @Tags objects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param bucket path string true "Bucket name"
// @Param prefix query string false "Object key prefix filter"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 50, max: 1000)"
// @Success 200 {object} domain.ListObjectsResponse "List of objects"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Bucket not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/buckets/{bucket}/objects [get]
func (h *ObjectHandler) ListObjects(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	bucketName := c.Param("bucket")
	prefix := c.Query("prefix")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	// Limit page size
	if pageSize > 1000 {
		pageSize = 1000
	}

	// Get bucket
	bucket, err := h.bucketUseCase.GetBucket(&userID, bucketName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bucket not found"})
		return
	}

	response, err := h.objectUseCase.ListObjects(bucket.ID, prefix, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListObjectVersions godoc
// @Summary List all versions of an object
// @Description Get all versions of a specific object
// @Tags objects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param bucket path string true "Bucket name"
// @Param key path string true "Object key"
// @Success 200 {object} map[string]interface{} "List of object versions"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Bucket not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/buckets/{bucket}/objects/{key}/versions [get]
func (h *ObjectHandler) ListObjectVersions(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	bucketName := c.Param("bucket")
	key := strings.TrimPrefix(c.Param("key"), "/")

	// Get bucket
	bucket, err := h.bucketUseCase.GetBucket(&userID, bucketName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bucket not found"})
		return
	}

	versions, err := h.objectUseCase.ListObjectVersions(bucket.ID, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"versions": versions,
		"count":    len(versions),
		"key":      key,
		"bucket":   bucketName,
	})
}

// DeleteObject godoc
// @Summary Delete an object
// @Description Delete the latest version of an object from a bucket
// @Tags objects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param bucket path string true "Bucket name"
// @Param key path string true "Object key"
// @Success 204 "Object deleted successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Object or bucket not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/buckets/{bucket}/objects/{key} [delete]
func (h *ObjectHandler) DeleteObject(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	bucketName := c.Param("bucket")
	key := strings.TrimPrefix(c.Param("key"), "/")

	// Get bucket
	bucket, err := h.bucketUseCase.GetBucket(&userID, bucketName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bucket not found"})
		return
	}

	if err := h.objectUseCase.DeleteObject(bucket.ID, key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
