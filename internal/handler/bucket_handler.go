package handler

import (
	"net/http"
	"s3-like/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BucketHandler struct {
	bucketUseCase domain.BucketUseCase
}

func NewBucketHandler(bucketUseCase domain.BucketUseCase) *BucketHandler {
	return &BucketHandler{
		bucketUseCase: bucketUseCase,
	}
}

// CreateBucket godoc
// @Summary Create a new bucket
// @Description Create a new storage bucket
// @Tags buckets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CreateBucketRequest true "Bucket creation details"
// @Success 201 {object} domain.Bucket "Bucket created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request or bucket already exists"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /api/v1/buckets [post]
func (h *BucketHandler) CreateBucket(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req domain.CreateBucketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bucket, err := h.bucketUseCase.CreateBucket(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, bucket)
}

// GetBucket godoc
// @Summary Get bucket details
// @Description Get details of a specific bucket
// @Tags buckets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param bucket path string true "Bucket name"
// @Success 200 {object} domain.Bucket "Bucket details"
// @Failure 404 {object} map[string]interface{} "Bucket not found"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /api/v1/buckets/{bucket} [get]
func (h *BucketHandler) GetBucket(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	bucketName := c.Param("bucket")

	bucket, err := h.bucketUseCase.GetBucket(userID, bucketName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bucket)
}

// ListBuckets godoc
// @Summary List user buckets
// @Description Get list of all buckets owned by the authenticated user
// @Tags buckets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "List of buckets"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/buckets [get]
func (h *BucketHandler) ListBuckets(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	buckets, err := h.bucketUseCase.ListBuckets(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"buckets": buckets,
		"count":   len(buckets),
	})
}

// DeleteBucket godoc
// @Summary Delete a bucket
// @Description Delete a bucket and all its contents
// @Tags buckets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param bucket path string true "Bucket name"
// @Success 204 "Bucket deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Bucket not found"
// @Router /api/v1/buckets/{bucket} [delete]
func (h *BucketHandler) DeleteBucket(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	bucketName := c.Param("bucket")

	if err := h.bucketUseCase.DeleteBucket(userID, bucketName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
