package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sellflow/backend/internal/connectors"
	"github.com/sellflow/backend/internal/models"
	"gorm.io/gorm"
)

type PublishService struct {
	db *gorm.DB
}

func NewPublishService(db *gorm.DB) *PublishService {
	return &PublishService{db: db}
}

// EnqueuePublish creates a publish queue entry for a product to a marketplace
func (s *PublishService) EnqueuePublish(productID uuid.UUID, marketplace models.MarketplaceType) (*models.PublishQueue, error) {
	// Check if there's already a pending/processing entry
	var existing models.PublishQueue
	err := s.db.Where("product_id = ? AND marketplace = ? AND status IN ?",
		productID, marketplace, []string{string(models.QueueStatusPending), string(models.QueueStatusProcessing)}).
		First(&existing).Error
	if err == nil {
		return &existing, nil // Already queued
	}

	entry := models.PublishQueue{
		ProductID:   productID,
		Marketplace: marketplace,
		Action:      models.QueueActionUpdate,
		Status:      models.QueueStatusPending,
	}

	if err := s.db.Create(&entry).Error; err != nil {
		return nil, fmt.Errorf("failed to enqueue publish: %w", err)
	}

	return &entry, nil
}

// ProcessQueue processes all pending entries in the publish queue
func (s *PublishService) ProcessQueue(ctx context.Context, getConnector func(userID uuid.UUID, marketplace models.MarketplaceType) (connectors.MarketplaceConnector, error)) {
	var entries []models.PublishQueue
	s.db.Where("status = ?", models.QueueStatusPending).
		Order("created_at ASC").
		Limit(50).
		Find(&entries)

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return
		default:
		}

		s.processEntry(ctx, entry, getConnector)
	}
}

func (s *PublishService) processEntry(ctx context.Context, entry models.PublishQueue, getConnector func(userID uuid.UUID, marketplace models.MarketplaceType) (connectors.MarketplaceConnector, error)) {
	// Mark as processing
	s.db.Model(&entry).Update("status", models.QueueStatusProcessing)

	// Load product with marketplace data
	var product models.Product
	if err := s.db.Preload("MarketplaceData").First(&product, entry.ProductID).Error; err != nil {
		s.failEntry(&entry, fmt.Sprintf("product not found: %v", err))
		return
	}

	// Get connector
	connector, err := getConnector(product.UserID, entry.Marketplace)
	if err != nil {
		s.failEntry(&entry, fmt.Sprintf("failed to get connector: %v", err))
		return
	}

	// Find marketplace data for this product
	var mpData *models.ProductMarketplaceData
	for i := range product.MarketplaceData {
		if product.MarketplaceData[i].Marketplace == entry.Marketplace {
			mpData = &product.MarketplaceData[i]
			break
		}
	}

	if mpData == nil || mpData.ExternalID == "" {
		s.failEntry(&entry, "product has no marketplace data with external ID")
		return
	}

	// Build update payload
	updates := buildUpdatePayload(product, mpData)

	// Execute update
	if err := connector.UpdateProduct(ctx, mpData.ExternalID, updates); err != nil {
		if entry.RetryCount < 3 {
			// Retry later
			s.db.Model(&entry).Updates(map[string]interface{}{
				"status":        models.QueueStatusPending,
				"retry_count":   entry.RetryCount + 1,
				"error_message": err.Error(),
			})
		} else {
			s.failEntry(&entry, err.Error())
		}

		// Update marketplace data status
		s.db.Model(mpData).Updates(map[string]interface{}{
			"status":        models.PublishStatusError,
			"error_message": err.Error(),
		})

		// Log
		s.logSync(product.UserID, entry.Marketplace, "product", product.ID.String(), string(entry.Action), "error", err.Error())
		return
	}

	// Success
	now := time.Now()
	s.db.Model(&entry).Updates(map[string]interface{}{
		"status":       models.QueueStatusDone,
		"processed_at": now,
	})

	// Update marketplace data
	s.db.Model(mpData).Updates(map[string]interface{}{
		"status":         models.PublishStatusPublished,
		"last_synced_at": now,
		"error_message":  "",
	})

	s.logSync(product.UserID, entry.Marketplace, "product", product.ID.String(), string(entry.Action), "success", "")

	log.Printf("Published product %s to %s (external: %s)", product.ID, entry.Marketplace, mpData.ExternalID)
}

func (s *PublishService) failEntry(entry *models.PublishQueue, errMsg string) {
	now := time.Now()
	s.db.Model(entry).Updates(map[string]interface{}{
		"status":        models.QueueStatusFailed,
		"error_message": errMsg,
		"processed_at":  now,
	})
}

func (s *PublishService) logSync(userID uuid.UUID, marketplace models.MarketplaceType, entityType, entityID, action, status, errMsg string) {
	s.db.Create(&models.SyncLog{
		UserID:       userID,
		Marketplace:  marketplace,
		EntityType:   entityType,
		EntityID:     entityID,
		Action:       action,
		Status:       status,
		ErrorMessage: errMsg,
	})
}

func buildUpdatePayload(product models.Product, mpData *models.ProductMarketplaceData) map[string]interface{} {
	updates := map[string]interface{}{}

	// Use custom title/description/price if set, otherwise use base product data
	title := product.Title
	if mpData.CustomTitle != "" {
		title = mpData.CustomTitle
	}
	updates["name"] = title

	desc := product.Description
	if mpData.CustomDescription != "" {
		desc = mpData.CustomDescription
	}
	if desc != "" {
		updates["description"] = desc
	}

	price := product.BasePrice
	if mpData.CustomPrice != nil && *mpData.CustomPrice > 0 {
		price = *mpData.CustomPrice
	} else if mpData.PriceMarkup != nil && *mpData.PriceMarkup != 0 {
		price = product.BasePrice * (1 + *mpData.PriceMarkup/100)
	}
	updates["price"] = fmt.Sprintf("%.2f", price)

	updates["quantity_in_stock"] = fmt.Sprintf("%d", product.Stock)

	if product.SKU != "" {
		updates["sku"] = product.SKU
	}

	return updates
}

// SyncProduct pulls latest data from marketplace and updates local product
func (s *PublishService) SyncProduct(ctx context.Context, product *models.Product, mpData *models.ProductMarketplaceData, connector connectors.MarketplaceConnector) error {
	remote, err := connector.GetProduct(ctx, mpData.ExternalID)
	if err != nil {
		s.logSync(product.UserID, mpData.Marketplace, "product", product.ID.String(), "sync", "error", err.Error())
		return err
	}

	// Update stock from marketplace (bidirectional)
	if remote.Stock >= 0 && remote.Stock != product.Stock {
		s.db.Model(product).Update("stock", remote.Stock)
	}

	// Update marketplace data
	now := time.Now()
	s.db.Model(mpData).Updates(map[string]interface{}{
		"last_synced_at": now,
		"status":         models.PublishStatusPublished,
		"error_message":  "",
	})

	s.logSync(product.UserID, mpData.Marketplace, "product", product.ID.String(), "sync", "success", "")
	return nil
}

// GetSyncLogs returns recent sync logs for a user
func (s *PublishService) GetSyncLogs(userID uuid.UUID, limit int) []models.SyncLog {
	var logs []models.SyncLog
	s.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Find(&logs)
	return logs
}

// GetQueueEntries returns recent queue entries for a product
func (s *PublishService) GetQueueEntries(productID uuid.UUID) []models.PublishQueue {
	var entries []models.PublishQueue
	s.db.Where("product_id = ?", productID).Order("created_at DESC").Limit(20).Find(&entries)
	return entries
}
