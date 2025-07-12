package handler

import (
	"net/http"
	"s3-like/internal/domain"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ObjectHandler struct {
	objectUseCase domain.ObjectUseCase
	bucketUseCase domain.BucketUseCase
}

func NewObjectHandler(objectUseCase domain.ObjectUseCase) *ObjectHandler {
	return &ObjectHandler{
		objectUseCase: objectUseCase,
	}
}

func (h *ObjectHandler) UploadObject(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	bucketName := c.Param("bucket")

	// Get bucket
	bucket, err := h.bucketUseCase.GetBucket(userID, bucketName)
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
		key = header.Filename
	}

	response, err := h.objectUseCase.UploadObject(bucket.ID, key, file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *ObjectHandler) GetObject(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	bucketName := c.Param("bucket")
	key := strings.TrimPrefix(c.Param("key"), "/")

	// Get bucket
	bucket, err := h.bucketUseCase.GetBucket(userID, bucketName)
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

	// Stream file
	c.DataFromReader(http.StatusOK, object.Size, object.ContentType, file, nil)
}

func (h *ObjectHandler) GetObjectVersion(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	bucketName := c.Param("bucket")
	key := strings.TrimPrefix(c.Param("key"), "/")
	versionID := c.Param("version")

	// Get bucket
	bucket, err := h.bucketUseCase.GetBucket(userID, bucketName)
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

	// Stream file
	c.DataFromReader(http.StatusOK, object.Size, object.ContentType, file, nil)
}

func (h *ObjectHandler) ListObjects(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	bucketName := c.Param("bucket")
	prefix := c.Query("prefix")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	// Get bucket
	bucket, err := h.bucketUseCase.GetBucket(userID, bucketName)
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

func (h *ObjectHandler) ListObjectVersions(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	bucketName := c.Param("bucket")
	key := strings.TrimPrefix(c.Param("key"), "/")

	// Get bucket
	bucket, err := h.bucketUseCase.GetBucket(userID, bucketName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bucket not found"})
		return
	}

	versions, err := h.objectUseCase.ListObjectVersions(bucket.ID, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"versions": versions})
}

func (h *ObjectHandler) DeleteObject(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	bucketName := c.Param("bucket")
	key := strings.TrimPrefix(c.Param("key"), "/")

	// Get bucket
	bucket, err := h.bucketUseCase.GetBucket(userID, bucketName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bucket not found"})
		return
	}

	if err := h.objectUseCase.DeleteObject(bucket.ID, key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
