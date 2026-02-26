package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sellflow/backend/internal/config"
	"github.com/sellflow/backend/internal/handlers"
	"github.com/sellflow/backend/internal/middleware"
	"github.com/sellflow/backend/internal/models"
	"github.com/sellflow/backend/internal/services"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(
		&models.User{},
		&models.MarketplaceConnection{},
		&models.Product{},
		&models.ProductImage{},
		&models.ProductAttribute{},
		&models.ProductVariant{},
		&models.ProductMarketplaceData{},
		&models.Category{},
		&models.CategoryMarketplaceMapping{},
		&models.Conversation{},
		&models.ConversationMessage{},
		&models.MessageTemplate{},
		&models.Review{},
		&models.ReviewResponse{},
		&models.SyncLog{},
		&models.PublishQueue{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migrated successfully")

	// Initialize services
	crypto, err := services.NewCryptoService(cfg.EncryptionKey)
	if err != nil {
		log.Fatal("Failed to initialize crypto service:", err)
	}

	storage, err := services.NewStorageService(cfg)
	if err != nil {
		log.Printf("Warning: MinIO not available: %v", err)
		storage = nil
	}

	importer := services.NewImportService(db, storage)

	// Setup router
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes
	authHandler := handlers.NewAuthHandler(db, cfg)
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Protected routes
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		protected.GET("/auth/me", authHandler.Me)

		// Marketplace routes
		marketplaceHandler := handlers.NewMarketplaceHandler(db, cfg, crypto)
		marketplaces := protected.Group("/marketplaces")
		{
			marketplaces.GET("", marketplaceHandler.List)
			marketplaces.POST("", marketplaceHandler.Connect)
			marketplaces.PUT("/:id", marketplaceHandler.Update)
			marketplaces.DELETE("/:id", marketplaceHandler.Delete)
			marketplaces.POST("/:id/test", marketplaceHandler.TestConnection)
		}

		// Product routes
		productHandler := handlers.NewProductHandler(db, storage)
		products := protected.Group("/products")
		{
			products.GET("", productHandler.List)
			products.GET("/:id", productHandler.Get)
			products.POST("", productHandler.Create)
			products.PUT("/:id", productHandler.Update)
			products.DELETE("/:id", productHandler.Delete)
			products.POST("/:id/images", productHandler.UploadImage)
			products.DELETE("/:id/images/:imageId", productHandler.DeleteImage)
		}

		// Import routes
		importHandler := handlers.NewImportHandler(db, cfg, crypto, importer)
		imports := protected.Group("/import")
		{
			imports.POST("/start", importHandler.Start)
			imports.GET("/status", importHandler.Status)
		}
	}

	// Start server
	addr := cfg.ServerHost + ":" + cfg.ServerPort
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
