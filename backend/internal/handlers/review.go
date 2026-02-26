package handlers

import (
	"net/http"
	"strconv"

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

type ReviewHandler struct {
	db        *gorm.DB
	cfg       *config.Config
	crypto    *services.CryptoService
	reviewSvc *services.ReviewService
}

func NewReviewHandler(db *gorm.DB, cfg *config.Config, crypto *services.CryptoService, reviewSvc *services.ReviewService) *ReviewHandler {
	return &ReviewHandler{db: db, cfg: cfg, crypto: crypto, reviewSvc: reviewSvc}
}

func (h *ReviewHandler) getConnector(userID uuid.UUID, marketplaceType models.MarketplaceType) *prom.Client {
	var conn models.MarketplaceConnection
	if err := h.db.Where("user_id = ? AND marketplace = ? AND is_active = ?",
		userID, marketplaceType, true).First(&conn).Error; err != nil {
		return nil
	}
	apiKey, err := h.crypto.Decrypt(conn.APIKeyEncrypted)
	if err != nil {
		return nil
	}
	return prom.NewClient(conn.ShopURL, apiKey)
}

func reviewToResponse(r models.Review) dto.ReviewResponse {
	resp := dto.ReviewResponse{
		ID:          r.ID.String(),
		Marketplace: string(r.Marketplace),
		ExternalID:  r.ExternalID,
		AuthorName:  r.AuthorName,
		Rating:      r.Rating,
		Body:        r.Body,
		Status:      string(r.Status),
		Responses:   []dto.ReviewResponseItem{},
		CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if r.ProductID != nil {
		pid := r.ProductID.String()
		resp.ProductID = &pid
	}

	if r.Product != nil {
		resp.ProductTitle = r.Product.Title
	}

	for _, rr := range r.Responses {
		resp.Responses = append(resp.Responses, dto.ReviewResponseItem{
			ID:        rr.ID.String(),
			Body:      rr.Body,
			CreatedAt: rr.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return resp
}

func (h *ReviewHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	rating := c.Query("rating")
	marketplace := c.Query("marketplace")
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	query := h.db.Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if rating != "" {
		r, err := strconv.Atoi(rating)
		if err == nil {
			query = query.Where("rating = ?", r)
		}
	}
	if marketplace != "" {
		query = query.Where("marketplace = ?", marketplace)
	}
	if search != "" {
		query = query.Where("body ILIKE ? OR author_name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Model(&models.Review{}).Count(&total)

	var reviews []models.Review
	query.Preload("Product").Preload("Responses").
		Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&reviews)

	resp := make([]dto.ReviewResponse, 0, len(reviews))
	for _, r := range reviews {
		resp = append(resp, reviewToResponse(r))
	}

	c.JSON(http.StatusOK, dto.ReviewsListResponse{
		Reviews: resp,
		Total:   total,
		Page:    page,
		Limit:   limit,
	})
}

func (h *ReviewHandler) Get(c *gin.Context) {
	userID := middleware.GetUserID(c)
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	var review models.Review
	if err := h.db.Where("id = ? AND user_id = ?", reviewID, userID).
		Preload("Product").Preload("Responses").
		First(&review).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	c.JSON(http.StatusOK, reviewToResponse(review))
}

func (h *ReviewHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	review := models.Review{
		UserID:      userID,
		Marketplace: models.MarketplaceType(req.Marketplace),
		ExternalID:  req.ExternalID,
		AuthorName:  req.AuthorName,
		Rating:      req.Rating,
		Body:        req.Body,
		Status:      models.ReviewStatusNew,
	}

	if req.ProductID != nil {
		pid, err := uuid.Parse(*req.ProductID)
		if err == nil {
			review.ProductID = &pid
		}
	}

	if err := h.db.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review"})
		return
	}

	// Reload with associations
	h.db.Preload("Product").Preload("Responses").First(&review, review.ID)

	c.JSON(http.StatusCreated, reviewToResponse(review))
}

func (h *ReviewHandler) Reply(c *gin.Context) {
	userID := middleware.GetUserID(c)
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	var review models.Review
	if err := h.db.Where("id = ? AND user_id = ?", reviewID, userID).First(&review).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	var req dto.ReplyToReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get connector for sending reply
	connector := h.getConnector(userID, review.Marketplace)

	if err := h.reviewSvc.ReplyToReview(c.Request.Context(), reviewID, req.Body, connector); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Reload
	h.db.Preload("Product").Preload("Responses").First(&review, reviewID)
	c.JSON(http.StatusOK, reviewToResponse(review))
}

func (h *ReviewHandler) UpdateStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	var req dto.UpdateReviewStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var review models.Review
	if err := h.db.Where("id = ? AND user_id = ?", reviewID, userID).First(&review).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	h.db.Model(&review).Update("status", req.Status)
	c.JSON(http.StatusOK, gin.H{"message": "Status updated"})
}

func (h *ReviewHandler) SyncReviews(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.SyncReviewsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	connector := prom.NewClient(conn.ShopURL, apiKey)

	// Run sync in background
	go func() {
		if err := h.reviewSvc.SyncReviews(c.Request.Context(), userID, conn.Marketplace, connector); err != nil {
			// Logged inside SyncReviews
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Review sync started"})
}

func (h *ReviewHandler) Stats(c *gin.Context) {
	userID := middleware.GetUserID(c)

	totalReviews, avgRating, ratingBreakdown, statusCounts := h.reviewSvc.GetStats(userID)

	c.JSON(http.StatusOK, dto.ReviewStatsResponse{
		TotalReviews:    totalReviews,
		AverageRating:   avgRating,
		RatingBreakdown: ratingBreakdown,
		NewCount:        statusCounts["new"],
		RespondedCount:  statusCounts["responded"],
		ResolvedCount:   statusCounts["resolved"],
	})
}

func (h *ReviewHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	var review models.Review
	if err := h.db.Where("id = ? AND user_id = ?", reviewID, userID).First(&review).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	h.db.Where("review_id = ?", reviewID).Delete(&models.ReviewResponse{})
	h.db.Delete(&review)

	c.JSON(http.StatusOK, gin.H{"message": "Review deleted"})
}
