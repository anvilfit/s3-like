package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"s3-like/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type authUseCase struct {
	userRepo         domain.UserRepository
	refreshTokenRepo domain.RefreshTokenRepository
	jwtSecret        string
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
}

func NewAuthUseCase(userRepo domain.UserRepository, refreshTokenRepo domain.RefreshTokenRepository, jwtSecret string) domain.AuthUseCase {
	return &authUseCase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtSecret:        jwtSecret,
		accessTokenTTL:   time.Hour * 1,      // 1 hour
		refreshTokenTTL:  time.Hour * 24 * 7, // 7 days
	}
}

func (uc *authUseCase) Login(username, password string) (*domain.AuthResponse, error) {
	user, err := uc.userRepo.GetByUsername(username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Revoke all existing refresh tokens for this user (optional - for single session)
	// uc.refreshTokenRepo.RevokeAllUserTokens(user.ID)

	return uc.generateTokenPair(user)
}

func (uc *authUseCase) Register(req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := uc.userRepo.Create(user); err != nil {
		return nil, err
	}

	return uc.generateTokenPair(user)
}

func (uc *authUseCase) RefreshToken(refreshToken string) (*domain.AuthResponse, error) {
	// Get refresh token from database
	tokenRecord, err := uc.refreshTokenRepo.GetByToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Check if token is expired
	if tokenRecord.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("refresh token expired")
	}

	// Check if token is revoked
	if tokenRecord.IsRevoked {
		return nil, errors.New("refresh token revoked")
	}

	// Revoke the used refresh token (token rotation)
	if err := uc.refreshTokenRepo.RevokeToken(refreshToken); err != nil {
		return nil, errors.New("failed to revoke token")
	}

	// Generate new token pair
	return uc.generateTokenPair(&tokenRecord.User)
}

func (uc *authUseCase) ValidateToken(tokenString string) (*domain.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(uc.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("invalid user ID in token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	return uc.userRepo.GetByID(userID)
}

func (uc *authUseCase) RevokeRefreshToken(refreshToken string) error {
	return uc.refreshTokenRepo.RevokeToken(refreshToken)
}

func (uc *authUseCase) RevokeAllUserTokens(userID uuid.UUID) error {
	return uc.refreshTokenRepo.RevokeAllUserTokens(userID)
}

func (uc *authUseCase) generateTokenPair(user *domain.User) (*domain.AuthResponse, error) {
	// Generate access token
	accessToken, err := uc.generateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := uc.generateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(uc.accessTokenTTL.Seconds()),
		User:         *user,
	}, nil
}

func (uc *authUseCase) generateAccessToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(uc.accessTokenTTL).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.jwtSecret))
}

func (uc *authUseCase) generateRefreshToken(userID uuid.UUID) (string, error) {
	// Generate random token
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	tokenString := hex.EncodeToString(bytes)

	// Create refresh token record
	refreshToken := &domain.RefreshToken{
		Token:     tokenString,
		UserID:    userID,
		ExpiresAt: time.Now().Add(uc.refreshTokenTTL),
		IsRevoked: false,
	}

	if err := uc.refreshTokenRepo.Create(refreshToken); err != nil {
		return "", err
	}

	return tokenString, nil
}

// Cleanup expired tokens (should be called periodically)
func (uc *authUseCase) CleanupExpiredTokens() error {
	return uc.refreshTokenRepo.CleanupExpiredTokens()
}
