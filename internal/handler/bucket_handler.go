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

func (h *BucketHandler) ListBuckets(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	buckets, err := h.bucketUseCase.ListBuckets(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"buckets": buckets})
}

func (h *BucketHandler) DeleteBucket(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	bucketName := c.Param("bucket")

	if err := h.bucketUseCase.DeleteBucket(userID, bucketName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
