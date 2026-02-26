package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sellflow/backend/internal/dto"
	"github.com/sellflow/backend/internal/middleware"
	"github.com/sellflow/backend/internal/models"
	"github.com/sellflow/backend/internal/services"
	"gorm.io/gorm"
)

type ProductHandler struct {
	db      *gorm.DB
	storage *services.StorageService
}

func NewProductHandler(db *gorm.DB, storage *services.StorageService) *ProductHandler {
	return &ProductHandler{db: db, storage: storage}
}

type ProductResponse struct {
	ID              string                    `json:"id"`
	Title           string                    `json:"title"`
	Description     string                    `json:"description"`
	SKU             string                    `json:"sku"`
	EAN             string                    `json:"ean"`
	BasePrice       float64                   `json:"base_price"`
	SalePrice       *float64                  `json:"sale_price"`
	Stock           int                       `json:"stock"`
	Status          string                    `json:"status"`
	CategoryID      *string                   `json:"category_id"`
	Images          []ProductImageResponse    `json:"images"`
	Attributes      []ProductAttrResponse     `json:"attributes"`
	MarketplaceData []ProductMPDataResponse   `json:"marketplace_data"`
	CreatedAt       string                    `json:"created_at"`
	UpdatedAt       string                    `json:"updated_at"`
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

func productToResponse(p models.Product) ProductResponse {
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
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if p.CategoryID != nil {
		catID := p.CategoryID.String()
		pr.CategoryID = &catID
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

	// Ensure slices are not nil in JSON
	if pr.Images == nil {
		pr.Images = []ProductImageResponse{}
	}
	if pr.Attributes == nil {
		pr.Attributes = []ProductAttrResponse{}
	}
	if pr.MarketplaceData == nil {
		pr.MarketplaceData = []ProductMPDataResponse{}
	}

	return pr
}

func (h *ProductHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	status := c.Query("status")

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
	if status != "" {
		query = query.Where("status = ?", status)
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
		resp = append(resp, productToResponse(p))
	}

	c.JSON(http.StatusOK, ProductsListResponse{
		Products: resp,
		Total:    total,
		Page:     page,
		Limit:    limit,
	})
}

func (h *ProductHandler) Get(c *gin.Context) {
	userID := middleware.GetUserID(c)

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product models.Product
	if err := h.db.Where("id = ? AND user_id = ?", productID, userID).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).Preload("Attributes").Preload("MarketplaceData").
		First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, productToResponse(product))
}

func (h *ProductHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product := models.Product{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		SKU:         req.SKU,
		EAN:         req.EAN,
		BasePrice:   req.BasePrice,
		SalePrice:   req.SalePrice,
		Stock:       req.Stock,
		Status:      models.ProductStatusDraft,
	}

	if req.Status != "" {
		product.Status = models.ProductStatus(req.Status)
	}

	if req.CategoryID != nil {
		catID, err := uuid.Parse(*req.CategoryID)
		if err == nil {
			product.CategoryID = &catID
		}
	}

	if err := h.db.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	// Create attributes
	for _, attr := range req.Attributes {
		h.db.Create(&models.ProductAttribute{
			ProductID: product.ID,
			Name:      attr.Name,
			Value:     attr.Value,
		})
	}

	// Reload with associations
	h.db.Where("id = ?", product.ID).
		Preload("Images").Preload("Attributes").Preload("MarketplaceData").
		First(&product)

	c.JSON(http.StatusCreated, productToResponse(product))
}

func (h *ProductHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product models.Product
	if err := h.db.Where("id = ? AND user_id = ?", productID, userID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.SKU != nil {
		updates["sku"] = *req.SKU
	}
	if req.EAN != nil {
		updates["ean"] = *req.EAN
	}
	if req.BasePrice != nil {
		updates["base_price"] = *req.BasePrice
	}
	if req.SalePrice != nil {
		updates["sale_price"] = *req.SalePrice
	}
	if req.Stock != nil {
		updates["stock"] = *req.Stock
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.CategoryID != nil {
		catID, err := uuid.Parse(*req.CategoryID)
		if err == nil {
			updates["category_id"] = catID
		}
	}

	if len(updates) > 0 {
		if err := h.db.Model(&product).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
			return
		}
	}

	// Update attributes if provided
	if req.Attributes != nil {
		// Delete old attributes
		h.db.Where("product_id = ?", productID).Delete(&models.ProductAttribute{})
		// Create new ones
		for _, attr := range req.Attributes {
			h.db.Create(&models.ProductAttribute{
				ProductID: productID,
				Name:      attr.Name,
				Value:     attr.Value,
			})
		}
	}

	// Reload
	h.db.Where("id = ?", productID).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).Preload("Attributes").Preload("MarketplaceData").
		First(&product)

	c.JSON(http.StatusOK, productToResponse(product))
}

func (h *ProductHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product models.Product
	if err := h.db.Where("id = ? AND user_id = ?", productID, userID).
		Preload("Images").First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Delete images from storage
	if h.storage != nil {
		for _, img := range product.Images {
			if img.StorageKey != "" {
				h.storage.Delete(c.Request.Context(), img.StorageKey)
			}
		}
	}

	// Delete all associated records
	h.db.Where("product_id = ?", productID).Delete(&models.ProductImage{})
	h.db.Where("product_id = ?", productID).Delete(&models.ProductAttribute{})
	h.db.Where("product_id = ?", productID).Delete(&models.ProductMarketplaceData{})
	h.db.Where("product_id = ?", productID).Delete(&models.ProductVariant{})
	h.db.Delete(&product)

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted"})
}

func (h *ProductHandler) UploadImage(c *gin.Context) {
	userID := middleware.GetUserID(c)

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product models.Product
	if err := h.db.Where("id = ? AND user_id = ?", productID, userID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	if h.storage == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Storage service not available"})
		return
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	storageKey, objectURL, err := h.storage.Upload(c.Request.Context(), file, header.Size, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	// Get max position
	var maxPos int
	h.db.Model(&models.ProductImage{}).Where("product_id = ?", productID).
		Select("COALESCE(MAX(position), -1)").Scan(&maxPos)

	image := models.ProductImage{
		ProductID:  productID,
		URL:        objectURL,
		StorageKey: storageKey,
		Position:   maxPos + 1,
	}

	if err := h.db.Create(&image).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	c.JSON(http.StatusCreated, ProductImageResponse{
		ID:       image.ID.String(),
		URL:      image.URL,
		Position: image.Position,
	})
}

func (h *ProductHandler) DeleteImage(c *gin.Context) {
	userID := middleware.GetUserID(c)

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	imageID, err := uuid.Parse(c.Param("imageId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID"})
		return
	}

	// Verify product belongs to user
	var product models.Product
	if err := h.db.Where("id = ? AND user_id = ?", productID, userID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	var image models.ProductImage
	if err := h.db.Where("id = ? AND product_id = ?", imageID, productID).First(&image).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	// Delete from storage
	if h.storage != nil && image.StorageKey != "" {
		h.storage.Delete(c.Request.Context(), image.StorageKey)
	}

	h.db.Delete(&image)
	c.JSON(http.StatusOK, gin.H{"message": "Image deleted"})
}
