package handlers

import (
	"context"
	"net/http"
	"sync"

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

type ImportHandler struct {
	db       *gorm.DB
	cfg      *config.Config
	crypto   *services.CryptoService
	importer *services.ImportService
	mu       sync.RWMutex
	progress map[uuid.UUID]*services.ImportProgress
}

func NewImportHandler(db *gorm.DB, cfg *config.Config, crypto *services.CryptoService, importer *services.ImportService) *ImportHandler {
	return &ImportHandler{
		db:       db,
		cfg:      cfg,
		crypto:   crypto,
		importer: importer,
		progress: make(map[uuid.UUID]*services.ImportProgress),
	}
}

func (h *ImportHandler) Start(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.StartImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if import is already running
	h.mu.RLock()
	if p, ok := h.progress[userID]; ok && p.Status == "running" {
		h.mu.RUnlock()
		c.JSON(http.StatusConflict, gin.H{"error": "Import is already running"})
		return
	}
	h.mu.RUnlock()

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

	apiKey, err := h.crypto.Decrypt(conn.APIKeyEncrypted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt API key"})
		return
	}

	connector := prom.NewClient(h.cfg.PromAPIBaseURL, apiKey)

	progressCh := make(chan services.ImportProgress, 100)

	h.mu.Lock()
	initial := &services.ImportProgress{Status: "running"}
	h.progress[userID] = initial
	h.mu.Unlock()

	go func() {
		for p := range progressCh {
			h.mu.Lock()
			cp := p
			h.progress[userID] = &cp
			h.mu.Unlock()
		}
	}()

	go h.importer.ImportProducts(context.Background(), userID, conn.Marketplace, connector, progressCh)

	c.JSON(http.StatusOK, gin.H{"message": "Import started"})
}

func (h *ImportHandler) Status(c *gin.Context) {
	userID := middleware.GetUserID(c)

	h.mu.RLock()
	p, ok := h.progress[userID]
	h.mu.RUnlock()

	if !ok {
		c.JSON(http.StatusOK, dto.ImportStatusResponse{Status: "idle"})
		return
	}

	c.JSON(http.StatusOK, dto.ImportStatusResponse{
		Total:    p.Total,
		Imported: p.Imported,
		Skipped:  p.Skipped,
		Failed:   p.Failed,
		Status:   p.Status,
		Error:    p.Error,
	})
}
