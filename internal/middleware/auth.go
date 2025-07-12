package middleware

import (
	"net/http"
	"s3-like/internal/usecase"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		authUseCase := usecase.NewAuthUseCase(nil, jwtSecret)
		userID, err := authUseCase.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", *userID)
		c.Next()
	}
}
