// @title S3-Like Storage API
// @version 1.0
// @description A complete S3-compatible storage service with authentication, versioning, and bucket management.
// @termsOfService http://swagger.io/terms/

// @contact.name S3-Like API Support
// @contact.email support@s3like.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:9080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"log"
	"s3-like/docs"
	"s3-like/internal/config"
	"s3-like/internal/database"
	"s3-like/internal/handler"
	"s3-like/internal/middleware"
	"s3-like/internal/repository"
	"s3-like/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg := config.Load()

	serverAddress := cfg.Server.IP + ":" + cfg.Server.Port

	// Initialize Swagger
	docs.SwaggerInfo.Host = serverAddress
	docs.SwaggerInfo.BasePath = "/"

	// Initialize database
	db, err := database.NewPostgresConnection(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	bucketRepo := repository.NewBucketRepository(db)
	objectRepo := repository.NewObjectRepository(db)

	// Initialize use cases
	authUseCase := usecase.NewAuthUseCase(userRepo)
	bucketUseCase := usecase.NewBucketUseCase(bucketRepo)
	objectUseCase := usecase.NewObjectUseCase(objectRepo, cfg.Storage.BasePath)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUseCase)
	bucketHandler := handler.NewBucketHandler(bucketUseCase)
	objectHandler := handler.NewObjectHandler(objectUseCase, bucketUseCase)

	// Setup router
	router := gin.Default()

	// Middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.ErrorHandler())

	// Routes
	setupRoutes(router, authHandler, bucketHandler, objectHandler, cfg.JWT.Secret)

	// Start server
	log.Printf("Server starting on port %s", cfg.Server.Port)
	log.Printf("Swagger UI available at: http://%s/swagger/index.html", serverAddress)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func setupRoutes(
	router *gin.Engine,
	authHandler *handler.AuthHandler,
	bucketHandler *handler.BucketHandler,
	objectHandler *handler.ObjectHandler,
	jwtSecret string,
) {
	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public routes
	auth := router.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
	}

	// Protected routes
	api := router.Group("/api/v1")
	api.Use(middleware.JWTAuth(jwtSecret))
	{
		// Bucket routes
		buckets := api.Group("/buckets")
		{
			buckets.GET("", bucketHandler.ListBuckets)
			buckets.POST("", bucketHandler.CreateBucket)
			buckets.GET("/:bucket", bucketHandler.GetBucket)
			buckets.DELETE("/:bucket", bucketHandler.DeleteBucket)
		}

		// Object routes
		objects := api.Group("/buckets/:bucket/objects")
		{
			objects.GET("", objectHandler.ListObjects)
			objects.POST("", objectHandler.UploadObject)
			// objects.GET("/:key", objectHandler.GetObject)
			objects.DELETE("/:key", objectHandler.DeleteObject)
			objects.GET("/:key/versions", objectHandler.ListObjectVersions)
			objects.GET("/:key/versions/:version", objectHandler.GetObjectVersion)
		}
	}

	router.GET("/api/v1/buckets/:bucket/objects/:key", objectHandler.GetObject)

	// Health check
	router.GET("/health", healthCheck)
}

// @Summary Health Check
// @Description Check if the service is running
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Service is healthy"
// @Router /health [get]
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": "s3-like-storage",
		"version": "1.0.0",
	})
}
