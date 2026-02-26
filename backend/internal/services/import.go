package services

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/sellflow/backend/internal/connectors"
	"github.com/sellflow/backend/internal/models"
	"gorm.io/gorm"
)

type ImportProgress struct {
	Total    int    `json:"total"`
	Imported int    `json:"imported"`
	Skipped  int    `json:"skipped"`
	Failed   int    `json:"failed"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}

type ImportService struct {
	db      *gorm.DB
	storage *StorageService
}

func NewImportService(db *gorm.DB, storage *StorageService) *ImportService {
	return &ImportService{db: db, storage: storage}
}

func (s *ImportService) ImportProducts(ctx context.Context, userID uuid.UUID, marketplace models.MarketplaceType, connector connectors.MarketplaceConnector, progressCh chan<- ImportProgress) {
	defer close(progressCh)

	progress := ImportProgress{Status: "running"}

	// Get total count first
	_, total, err := connector.GetProducts(ctx, 1, 1)
	if err != nil {
		progress.Status = "failed"
		progress.Error = fmt.Sprintf("failed to get products count: %v", err)
		progressCh <- progress
		return
	}
	progress.Total = total
	progressCh <- progress

	// Import categories first
	categories, err := connector.GetCategories(ctx)
	if err != nil {
		log.Printf("Warning: failed to import categories: %v", err)
	} else {
		s.importCategories(userID, marketplace, categories)
	}

	// Import products page by page
	page := 1
	limit := 100
	for {
		products, _, err := connector.GetProducts(ctx, page, limit)
		if err != nil {
			progress.Failed++
			log.Printf("Error fetching page %d: %v", page, err)
			page++
			if progress.Imported+progress.Skipped+progress.Failed >= progress.Total {
				break
			}
			continue
		}

		if len(products) == 0 {
			break
		}

		for _, mp := range products {
			if err := s.importProduct(ctx, userID, marketplace, mp); err != nil {
				progress.Failed++
				log.Printf("Error importing product %s: %v", mp.ExternalID, err)
			} else {
				progress.Imported++
			}
			progressCh <- progress
		}

		if len(products) < limit {
			break
		}
		page++
	}

	progress.Status = "completed"
	progressCh <- progress
}

func (s *ImportService) importCategories(userID uuid.UUID, marketplace models.MarketplaceType, categories []connectors.MarketplaceCategory) {
	for _, mc := range categories {
		var mapping models.CategoryMarketplaceMapping
		result := s.db.Where("external_category_id = ? AND marketplace = ?", mc.ExternalID, marketplace).First(&mapping)
		if result.Error == nil {
			continue
		}

		cat := models.Category{
			UserID: userID,
			Name:   mc.Name,
		}
		if err := s.db.Create(&cat).Error; err != nil {
			log.Printf("Error creating category %s: %v", mc.Name, err)
			continue
		}

		catMapping := models.CategoryMarketplaceMapping{
			CategoryID:           cat.ID,
			Marketplace:          marketplace,
			ExternalCategoryID:   mc.ExternalID,
			ExternalCategoryName: mc.Name,
		}
		s.db.Create(&catMapping)
	}
}

func (s *ImportService) importProduct(ctx context.Context, userID uuid.UUID, marketplace models.MarketplaceType, mp connectors.MarketplaceProduct) error {
	// Check if product already imported
	var existing models.ProductMarketplaceData
	if err := s.db.Where("external_id = ? AND marketplace = ?", mp.ExternalID, marketplace).First(&existing).Error; err == nil {
		return nil
	}

	// Find category
	var categoryID *uuid.UUID
	if mp.CategoryID != "" {
		var mapping models.CategoryMarketplaceMapping
		if err := s.db.Where("external_category_id = ? AND marketplace = ?", mp.CategoryID, marketplace).First(&mapping).Error; err == nil {
			categoryID = &mapping.CategoryID
		}
	}

	product := models.Product{
		UserID:      userID,
		Title:       mp.Name,
		Description: mp.Description,
		SKU:         mp.SKU,
		BasePrice:   mp.Price,
		Stock:       mp.Stock,
		Status:      models.ProductStatusActive,
		CategoryID:  categoryID,
	}

	if err := s.db.Create(&product).Error; err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	// Import images
	for i, imageURL := range mp.Images {
		storageKey, objectURL, err := s.storage.UploadFromURL(ctx, imageURL)
		if err != nil {
			log.Printf("Warning: failed to upload image for product %s: %v", mp.ExternalID, err)
			img := models.ProductImage{
				ProductID: product.ID,
				URL:       imageURL,
				Position:  i,
			}
			s.db.Create(&img)
			continue
		}

		img := models.ProductImage{
			ProductID:  product.ID,
			URL:        objectURL,
			StorageKey: storageKey,
			Position:   i,
		}
		s.db.Create(&img)
	}

	// Import attributes
	for name, value := range mp.Attributes {
		attr := models.ProductAttribute{
			ProductID: product.ID,
			Name:      name,
			Value:     value,
		}
		s.db.Create(&attr)
	}

	// Create marketplace data link
	marketplaceData := models.ProductMarketplaceData{
		ProductID:      product.ID,
		Marketplace:    marketplace,
		ExternalID:     mp.ExternalID,
		MarketplaceURL: mp.URL,
		Status:         models.PublishStatusPublished,
	}
	s.db.Create(&marketplaceData)

	return nil
}
