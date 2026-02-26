package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sellflow/backend/internal/config"
	"github.com/sellflow/backend/internal/connectors"
	"github.com/sellflow/backend/internal/connectors/prom"
	"github.com/sellflow/backend/internal/dto"
	"github.com/sellflow/backend/internal/middleware"
	"github.com/sellflow/backend/internal/models"
	"github.com/sellflow/backend/internal/services"
	"gorm.io/gorm"
)

type PublishHandler struct {
	db        *gorm.DB
	cfg       *config.Config
	crypto    *services.CryptoService
	publisher *services.PublishService
}

func NewPublishHandler(db *gorm.DB, cfg *config.Config, crypto *services.CryptoService, publisher *services.PublishService) *PublishHandler {
	h := &PublishHandler{
		db:        db,
		cfg:       cfg,
		crypto:    crypto,
		publisher: publisher,
	}

	// Start background queue worker
	go h.queueWorker()

	return h
}

func (h *PublishHandler) queueWorker() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		h.publisher.ProcessQueue(context.Background(), h.getConnector)
	}
}

func (h *PublishHandler) getConnector(userID uuid.UUID, marketplace models.MarketplaceType) (connectors.MarketplaceConnector, error) {
	var conn models.MarketplaceConnection
	if err := h.db.Where("user_id = ? AND marketplace = ? AND is_active = true", userID, marketplace).
		First(&conn).Error; err != nil {
		return nil, err
	}

	apiKey, err := h.crypto.Decrypt(conn.APIKeyEncrypted)
	if err != nil {
		return nil, err
	}

	return prom.NewClient(h.cfg.PromAPIBaseURL, apiKey), nil
}

// Publish enqueues a product for publishing to a marketplace
func (h *PublishHandler) Publish(c *gin.Context) {
	userID := middleware.GetUserID(c)

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req dto.PublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify product belongs to user
	var product models.Product
	if err := h.db.Where("id = ? AND user_id = ?", productID, userID).
		Preload("MarketplaceData").First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Get marketplace connection
	connID, err := uuid.Parse(req.MarketplaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid marketplace ID"})
		return
	}

	var conn models.MarketplaceConnection
	if err := h.db.Where("id = ? AND user_id = ?", connID, userID).First(&conn).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Marketplace connection not found"})
		return
	}

	// Check if product has marketplace data
	hasMP := false
	for _, md := range product.MarketplaceData {
		if md.Marketplace == conn.Marketplace {
			hasMP = true
			break
		}
	}

	if !hasMP {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product has no data for this marketplace. Import it first."})
		return
	}

	// Enqueue
	entry, err := h.publisher.EnqueuePublish(productID, conn.Marketplace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue publish"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Product queued for publishing",
		"queue_id": entry.ID.String(),
		"status":   string(entry.Status),
	})
}

// SyncProduct syncs a product with its marketplace data
func (h *PublishHandler) SyncProduct(c *gin.Context) {
	userID := middleware.GetUserID(c)

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req dto.SyncProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	connID, err := uuid.Parse(req.MarketplaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid marketplace ID"})
		return
	}

	var product models.Product
	if err := h.db.Where("id = ? AND user_id = ?", productID, userID).
		Preload("MarketplaceData").First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	var conn models.MarketplaceConnection
	if err := h.db.Where("id = ? AND user_id = ?", connID, userID).First(&conn).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Marketplace connection not found"})
		return
	}

	// Find marketplace data
	var mpData *models.ProductMarketplaceData
	for i := range product.MarketplaceData {
		if product.MarketplaceData[i].Marketplace == conn.Marketplace {
			mpData = &product.MarketplaceData[i]
			break
		}
	}

	if mpData == nil || mpData.ExternalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product has no marketplace data"})
		return
	}

	connector, err := h.getConnector(userID, conn.Marketplace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get connector"})
		return
	}

	if err := h.publisher.SyncProduct(c.Request.Context(), &product, mpData, connector); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product synced successfully"})
}

// GetSyncLogs returns sync logs for the current user
func (h *PublishHandler) GetSyncLogs(c *gin.Context) {
	userID := middleware.GetUserID(c)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 200 {
		limit = 50
	}

	logs := h.publisher.GetSyncLogs(userID, limit)

	result := make([]dto.SyncLogResponse, 0, len(logs))
	for _, l := range logs {
		result = append(result, dto.SyncLogResponse{
			ID:          l.ID.String(),
			Marketplace: string(l.Marketplace),
			EntityType:  l.EntityType,
			EntityID:    l.EntityID,
			Action:      l.Action,
			Status:      l.Status,
			Error:       l.ErrorMessage,
			CreatedAt:   l.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, result)
}

// GetProductQueue returns publish queue entries for a product
func (h *PublishHandler) GetProductQueue(c *gin.Context) {
	userID := middleware.GetUserID(c)

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	// Verify ownership
	var product models.Product
	if err := h.db.Where("id = ? AND user_id = ?", productID, userID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	entries := h.publisher.GetQueueEntries(productID)

	result := make([]dto.PublishQueueItemResponse, 0, len(entries))
	for _, e := range entries {
		item := dto.PublishQueueItemResponse{
			ID:          e.ID.String(),
			ProductID:   e.ProductID.String(),
			Marketplace: string(e.Marketplace),
			Action:      string(e.Action),
			Status:      string(e.Status),
			RetryCount:  e.RetryCount,
			Error:       e.ErrorMessage,
			CreatedAt:   e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if e.ProcessedAt != nil {
			t := e.ProcessedAt.Format("2006-01-02T15:04:05Z")
			item.ProcessedAt = &t
		}
		result = append(result, item)
	}

	c.JSON(http.StatusOK, result)
}
