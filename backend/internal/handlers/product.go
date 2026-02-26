package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sellflow/backend/internal/middleware"
	"github.com/sellflow/backend/internal/models"
	"gorm.io/gorm"
)

type ProductHandler struct {
	db *gorm.DB
}

func NewProductHandler(db *gorm.DB) *ProductHandler {
	return &ProductHandler{db: db}
}

type ProductResponse struct {
	ID          string                    `json:"id"`
	Title       string                    `json:"title"`
	Description string                    `json:"description"`
	SKU         string                    `json:"sku"`
	EAN         string                    `json:"ean"`
	BasePrice   float64                   `json:"base_price"`
	SalePrice   *float64                  `json:"sale_price"`
	Stock       int                       `json:"stock"`
	Status      string                    `json:"status"`
	Images      []ProductImageResponse    `json:"images"`
	Attributes  []ProductAttrResponse     `json:"attributes"`
	MarketplaceData []ProductMPDataResponse `json:"marketplace_data"`
	CreatedAt   string                    `json:"created_at"`
}

type ProductImageResponse struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Position int    `json:"position"`
}

type ProductAttrResponse struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ProductMPDataResponse struct {
	Marketplace    string `json:"marketplace"`
	ExternalID     string `json:"external_id"`
	Status         string `json:"status"`
	MarketplaceURL string `json:"marketplace_url"`
}

type ProductsListResponse struct {
	Products []ProductResponse `json:"products"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
}

func (h *ProductHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	query := h.db.Where("user_id = ?", userID)
	if search != "" {
		query = query.Where("title ILIKE ? OR sku ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Model(&models.Product{}).Count(&total)

	var products []models.Product
	query.Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Order("position ASC")
	}).Preload("Attributes").Preload("MarketplaceData").
		Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&products)

	resp := make([]ProductResponse, 0, len(products))
	for _, p := range products {
		pr := ProductResponse{
			ID:          p.ID.String(),
			Title:       p.Title,
			Description: p.Description,
			SKU:         p.SKU,
			EAN:         p.EAN,
			BasePrice:   p.BasePrice,
			SalePrice:   p.SalePrice,
			Stock:       p.Stock,
			Status:      string(p.Status),
			CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}

		for _, img := range p.Images {
			pr.Images = append(pr.Images, ProductImageResponse{
				ID:       img.ID.String(),
				URL:      img.URL,
				Position: img.Position,
			})
		}

		for _, attr := range p.Attributes {
			pr.Attributes = append(pr.Attributes, ProductAttrResponse{
				Name:  attr.Name,
				Value: attr.Value,
			})
		}

		for _, md := range p.MarketplaceData {
			pr.MarketplaceData = append(pr.MarketplaceData, ProductMPDataResponse{
				Marketplace:    string(md.Marketplace),
				ExternalID:     md.ExternalID,
				Status:         string(md.Status),
				MarketplaceURL: md.MarketplaceURL,
			})
		}

		resp = append(resp, pr)
	}

	c.JSON(http.StatusOK, ProductsListResponse{
		Products: resp,
		Total:    total,
		Page:     page,
		Limit:    limit,
	})
}
