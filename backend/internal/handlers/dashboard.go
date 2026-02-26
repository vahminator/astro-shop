package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sellflow/backend/internal/middleware"
	"github.com/sellflow/backend/internal/models"
	"gorm.io/gorm"
)

type DashboardHandler struct {
	db *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

type DashboardResponse struct {
	Products     ProductStats        `json:"products"`
	Inbox        InboxStats          `json:"inbox"`
	Reviews      ReviewStatsOverview `json:"reviews"`
	Marketplaces []MarketplaceStatus `json:"marketplaces"`
	RecentActivity []ActivityItem    `json:"recent_activity"`
}

type ProductStats struct {
	Total    int64 `json:"total"`
	Active   int64 `json:"active"`
	Draft    int64 `json:"draft"`
	Archived int64 `json:"archived"`
	Errors   int64 `json:"errors"`
}

type InboxStats struct {
	TotalConversations int64 `json:"total_conversations"`
	NewConversations   int64 `json:"new_conversations"`
	UnreadMessages     int64 `json:"unread_messages"`
}

type ReviewStatsOverview struct {
	Total         int64   `json:"total"`
	AverageRating float64 `json:"average_rating"`
	NewCount      int64   `json:"new_count"`
}

type MarketplaceStatus struct {
	ID          string  `json:"id"`
	Marketplace string  `json:"marketplace"`
	IsActive    bool    `json:"is_active"`
	LastSyncAt  *string `json:"last_sync_at"`
	ProductCount int64  `json:"product_count"`
}

type ActivityItem struct {
	Type      string `json:"type"` // product_created, message_received, review_new, sync_complete, product_published
	Title     string `json:"title"`
	Details   string `json:"details"`
	CreatedAt string `json:"created_at"`
}

func (h *DashboardHandler) Overview(c *gin.Context) {
	userID := middleware.GetUserID(c)

	// Product stats
	var prodStats ProductStats
	h.db.Model(&models.Product{}).Where("user_id = ?", userID).Count(&prodStats.Total)
	h.db.Model(&models.Product{}).Where("user_id = ? AND status = ?", userID, "active").Count(&prodStats.Active)
	h.db.Model(&models.Product{}).Where("user_id = ? AND status = ?", userID, "draft").Count(&prodStats.Draft)
	h.db.Model(&models.Product{}).Where("user_id = ? AND status = ?", userID, "archived").Count(&prodStats.Archived)

	// Count products with error status in marketplace data
	h.db.Model(&models.ProductMarketplaceData{}).
		Joins("JOIN products ON products.id = product_marketplace_data.product_id").
		Where("products.user_id = ? AND product_marketplace_data.status = ?", userID, "error").
		Count(&prodStats.Errors)

	// Inbox stats
	var inboxStats InboxStats
	h.db.Model(&models.Conversation{}).Where("user_id = ?", userID).Count(&inboxStats.TotalConversations)
	h.db.Model(&models.Conversation{}).Where("user_id = ? AND status = ?", userID, "new").Count(&inboxStats.NewConversations)

	// Count unread messages
	h.db.Model(&models.ConversationMessage{}).
		Joins("JOIN conversations ON conversations.id = conversation_messages.conversation_id").
		Where("conversations.user_id = ? AND conversation_messages.sender = ? AND conversation_messages.is_read = ?",
			userID, "customer", false).
		Count(&inboxStats.UnreadMessages)

	// Review stats
	var reviewStats ReviewStatsOverview
	h.db.Model(&models.Review{}).Where("user_id = ?", userID).Count(&reviewStats.Total)
	h.db.Model(&models.Review{}).Where("user_id = ? AND status = ?", userID, "new").Count(&reviewStats.NewCount)

	var avgRating *float64
	h.db.Model(&models.Review{}).Where("user_id = ? AND rating > 0", userID).
		Select("AVG(rating)").Scan(&avgRating)
	if avgRating != nil {
		reviewStats.AverageRating = *avgRating
	}

	// Marketplace statuses
	var connections []models.MarketplaceConnection
	h.db.Where("user_id = ?", userID).Find(&connections)

	mpStatuses := make([]MarketplaceStatus, 0, len(connections))
	for _, conn := range connections {
		var prodCount int64
		h.db.Model(&models.ProductMarketplaceData{}).
			Joins("JOIN products ON products.id = product_marketplace_data.product_id").
			Where("products.user_id = ? AND product_marketplace_data.marketplace = ?", userID, conn.Marketplace).
			Count(&prodCount)

		ms := MarketplaceStatus{
			ID:           conn.ID.String(),
			Marketplace:  string(conn.Marketplace),
			IsActive:     conn.IsActive,
			ProductCount: prodCount,
		}
		if conn.LastSyncAt != nil {
			t := conn.LastSyncAt.Format("2006-01-02T15:04:05Z")
			ms.LastSyncAt = &t
		}
		mpStatuses = append(mpStatuses, ms)
	}

	// Recent activity from sync logs
	var syncLogs []models.SyncLog
	h.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(10).Find(&syncLogs)

	activities := make([]ActivityItem, 0)

	for _, log := range syncLogs {
		item := ActivityItem{
			Type:      log.EntityType + "_" + log.Action,
			Title:     formatActivityTitle(log),
			Details:   log.ErrorMessage,
			CreatedAt: log.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		activities = append(activities, item)
	}

	// Recent products
	var recentProducts []models.Product
	h.db.Where("user_id = ? AND created_at > ?", userID, time.Now().AddDate(0, 0, -7)).
		Order("created_at DESC").Limit(5).Find(&recentProducts)

	for _, p := range recentProducts {
		activities = append(activities, ActivityItem{
			Type:      "product_created",
			Title:     "Product added: " + p.Title,
			CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	// Sort by time and limit
	if len(activities) > 15 {
		activities = activities[:15]
	}

	c.JSON(http.StatusOK, DashboardResponse{
		Products:       prodStats,
		Inbox:          inboxStats,
		Reviews:        reviewStats,
		Marketplaces:   mpStatuses,
		RecentActivity: activities,
	})
}

func formatActivityTitle(log models.SyncLog) string {
	switch {
	case log.EntityType == "product" && log.Action == "import":
		return "Products imported from " + string(log.Marketplace)
	case log.EntityType == "product" && log.Action == "publish":
		return "Product published to " + string(log.Marketplace)
	case log.EntityType == "product" && log.Action == "sync":
		return "Product synced with " + string(log.Marketplace)
	case log.EntityType == "review" && log.Action == "sync":
		return "Reviews synced from " + string(log.Marketplace)
	case log.EntityType == "conversation" && log.Action == "sync":
		return "Inbox synced from " + string(log.Marketplace)
	default:
		return log.EntityType + " " + log.Action
	}
}
