package repository

import (
	"s3-like/internal/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) domain.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(token *domain.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *refreshTokenRepository) GetByToken(token string) (*domain.RefreshToken, error) {
	var refreshToken domain.RefreshToken
	err := r.db.Preload("User").Where("token = ? AND is_revoked = false AND expires_at > ?", token, time.Now()).First(&refreshToken).Error
	if err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (r *refreshTokenRepository) RevokeToken(token string) error {
	return r.db.Model(&domain.RefreshToken{}).
		Where("token = ?", token).
		Update("is_revoked", true).Error
}

func (r *refreshTokenRepository) RevokeAllUserTokens(userID uuid.UUID) error {
	return r.db.Model(&domain.RefreshToken{}).
		Where("user_id = ? AND is_revoked = false", userID).
		Update("is_revoked", true).Error
}

func (r *refreshTokenRepository) CleanupExpiredTokens() error {
	return r.db.Where("expires_at < ? OR is_revoked = true", time.Now()).
		Delete(&domain.RefreshToken{}).Error
}
