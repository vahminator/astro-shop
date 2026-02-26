package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sellflow/backend/internal/config"
	"github.com/sellflow/backend/internal/connectors/prom"
	"github.com/sellflow/backend/internal/dto"
	"github.com/sellflow/backend/internal/middleware"
	"github.com/sellflow/backend/internal/models"
	"github.com/sellflow/backend/internal/services"
	"gorm.io/gorm"
)

type MarketplaceHandler struct {
	db     *gorm.DB
	cfg    *config.Config
	crypto *services.CryptoService
}

func NewMarketplaceHandler(db *gorm.DB, cfg *config.Config, crypto *services.CryptoService) *MarketplaceHandler {
	return &MarketplaceHandler{db: db, cfg: cfg, crypto: crypto}
}

func (h *MarketplaceHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var connections []models.MarketplaceConnection
	h.db.Where("user_id = ?", userID).Order("created_at desc").Find(&connections)

	result := make([]dto.MarketplaceConnectionResponse, 0, len(connections))
	for _, conn := range connections {
		r := dto.MarketplaceConnectionResponse{
			ID:          conn.ID.String(),
			Marketplace: string(conn.Marketplace),
			ShopURL:     conn.ShopURL,
			ShopName:    conn.ShopName,
			IsActive:    conn.IsActive,
			CreatedAt:   conn.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if conn.LastSyncAt != nil {
			r.LastSyncAt = conn.LastSyncAt.Format("2006-01-02T15:04:05Z")
		}
		result = append(result, r)
	}

	c.JSON(http.StatusOK, result)
}

func (h *MarketplaceHandler) Connect(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.ConnectMarketplaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate marketplace type
	mp := models.MarketplaceType(req.Marketplace)
	if mp != models.MarketplaceProm && mp != models.MarketplaceRozetka {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported marketplace. Supported: prom, rozetka"})
		return
	}

	// Check if connection already exists
	var existing models.MarketplaceConnection
	if err := h.db.Where("user_id = ? AND marketplace = ?", userID, mp).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Connection for this marketplace already exists"})
		return
	}

	// Encrypt API key
	encrypted, err := h.crypto.Encrypt(req.APIKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt API key"})
		return
	}

	conn := models.MarketplaceConnection{
		UserID:          userID,
		Marketplace:     mp,
		APIKeyEncrypted: encrypted,
		ShopURL:         req.ShopURL,
		IsActive:        true,
	}

	if err := h.db.Create(&conn).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create connection"})
		return
	}

	c.JSON(http.StatusCreated, dto.MarketplaceConnectionResponse{
		ID:          conn.ID.String(),
		Marketplace: string(conn.Marketplace),
		ShopURL:     conn.ShopURL,
		IsActive:    conn.IsActive,
		CreatedAt:   conn.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *MarketplaceHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	connID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	var conn models.MarketplaceConnection
	if err := h.db.Where("id = ? AND user_id = ?", connID, userID).First(&conn).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
		return
	}

	var req dto.UpdateMarketplaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.APIKey != "" {
		encrypted, err := h.crypto.Encrypt(req.APIKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt API key"})
			return
		}
		conn.APIKeyEncrypted = encrypted
	}
	if req.ShopURL != "" {
		conn.ShopURL = req.ShopURL
	}
	if req.IsActive != nil {
		conn.IsActive = *req.IsActive
	}

	h.db.Save(&conn)

	c.JSON(http.StatusOK, dto.MarketplaceConnectionResponse{
		ID:          conn.ID.String(),
		Marketplace: string(conn.Marketplace),
		ShopURL:     conn.ShopURL,
		ShopName:    conn.ShopName,
		IsActive:    conn.IsActive,
		CreatedAt:   conn.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *MarketplaceHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	connID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	result := h.db.Where("id = ? AND user_id = ?", connID, userID).Delete(&models.MarketplaceConnection{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Connection deleted"})
}

func (h *MarketplaceHandler) TestConnection(c *gin.Context) {
	userID := middleware.GetUserID(c)
	connID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	var conn models.MarketplaceConnection
	if err := h.db.Where("id = ? AND user_id = ?", connID, userID).First(&conn).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
		return
	}

	apiKey, err := h.crypto.Decrypt(conn.APIKeyEncrypted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt API key"})
		return
	}

	// Create connector based on marketplace type
	switch conn.Marketplace {
	case models.MarketplaceProm:
		client := prom.NewClient(h.cfg.PromAPIBaseURL, apiKey)
		if err := client.TestConnection(c.Request.Context()); err != nil {
			c.JSON(http.StatusOK, dto.TestConnectionResponse{
				Success: false,
				Message: "Connection failed: " + err.Error(),
			})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Connector not implemented for this marketplace"})
		return
	}

	c.JSON(http.StatusOK, dto.TestConnectionResponse{
		Success: true,
		Message: "Connection successful",
	})
}

// GetConnector returns a marketplace connector for the given connection
func (h *MarketplaceHandler) GetConnector(conn *models.MarketplaceConnection) (interface{}, error) {
	apiKey, err := h.crypto.Decrypt(conn.APIKeyEncrypted)
	if err != nil {
		return nil, err
	}

	switch conn.Marketplace {
	case models.MarketplaceProm:
		return prom.NewClient(h.cfg.PromAPIBaseURL, apiKey), nil
	default:
		return nil, fmt.Errorf("connector not implemented for %s", conn.Marketplace)
	}
}
