package handlers

import (
	"context"
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

type InboxHandler struct {
	db     *gorm.DB
	cfg    *config.Config
	crypto *services.CryptoService
	inbox  *services.InboxService
}

func NewInboxHandler(db *gorm.DB, cfg *config.Config, crypto *services.CryptoService, inbox *services.InboxService) *InboxHandler {
	return &InboxHandler{db: db, cfg: cfg, crypto: crypto, inbox: inbox}
}

// ListConversations returns all conversations for the current user
func (h *InboxHandler) ListConversations(c *gin.Context) {
	userID := middleware.GetUserID(c)
	status := c.Query("status")

	query := h.db.Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var conversations []models.Conversation
	query.Order("last_message_at DESC NULLS LAST").Find(&conversations)

	result := make([]dto.ConversationResponse, 0, len(conversations))
	for _, conv := range conversations {
		cr := dto.ConversationResponse{
			ID:           conv.ID.String(),
			Marketplace:  string(conv.Marketplace),
			ExternalID:   conv.ExternalID,
			CustomerName: conv.CustomerName,
			CustomerID:   conv.CustomerID,
			Subject:      conv.Subject,
			Status:       string(conv.Status),
			CreatedAt:    conv.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if conv.LastMessageAt != nil {
			t := conv.LastMessageAt.Format("2006-01-02T15:04:05Z")
			cr.LastMessageAt = &t
		}

		// Get last message and unread count
		var lastMsg models.ConversationMessage
		if err := h.db.Where("conversation_id = ?", conv.ID).
			Order("created_at DESC").First(&lastMsg).Error; err == nil {
			cr.LastMessage = &lastMsg.Body
		}

		var unread int64
		h.db.Model(&models.ConversationMessage{}).
			Where("conversation_id = ? AND is_read = false AND sender = ?", conv.ID, models.SenderCustomer).
			Count(&unread)
		cr.UnreadCount = int(unread)

		result = append(result, cr)
	}

	c.JSON(http.StatusOK, result)
}

// GetConversation returns a single conversation with messages
func (h *InboxHandler) GetConversation(c *gin.Context) {
	userID := middleware.GetUserID(c)

	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	var conv models.Conversation
	if err := h.db.Where("id = ? AND user_id = ?", convID, userID).First(&conv).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	var messages []models.ConversationMessage
	h.db.Where("conversation_id = ?", convID).Order("created_at ASC").Find(&messages)

	// Mark unread messages as read
	h.db.Model(&models.ConversationMessage{}).
		Where("conversation_id = ? AND is_read = false AND sender = ?", convID, models.SenderCustomer).
		Update("is_read", true)

	cr := dto.ConversationDetailResponse{
		ConversationResponse: dto.ConversationResponse{
			ID:           conv.ID.String(),
			Marketplace:  string(conv.Marketplace),
			ExternalID:   conv.ExternalID,
			CustomerName: conv.CustomerName,
			CustomerID:   conv.CustomerID,
			Subject:      conv.Subject,
			Status:       string(conv.Status),
			CreatedAt:    conv.CreatedAt.Format("2006-01-02T15:04:05Z"),
		},
	}
	if conv.LastMessageAt != nil {
		t := conv.LastMessageAt.Format("2006-01-02T15:04:05Z")
		cr.LastMessageAt = &t
	}

	for _, msg := range messages {
		cr.Messages = append(cr.Messages, dto.MessageResponse{
			ID:         msg.ID.String(),
			ExternalID: msg.ExternalID,
			Sender:     string(msg.Sender),
			Body:       msg.Body,
			IsRead:     msg.IsRead,
			CreatedAt:  msg.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	if cr.Messages == nil {
		cr.Messages = []dto.MessageResponse{}
	}

	c.JSON(http.StatusOK, cr)
}

// SendMessage sends a reply in a conversation
func (h *InboxHandler) SendMessage(c *gin.Context) {
	userID := middleware.GetUserID(c)

	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var conv models.Conversation
	if err := h.db.Where("id = ? AND user_id = ?", convID, userID).First(&conv).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	// Get connector
	connector, err := h.getConnector(userID, conv.Marketplace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get marketplace connector"})
		return
	}

	if err := h.inbox.SendMessage(c.Request.Context(), convID, req.Body, connector); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message sent"})
}

// UpdateStatus updates conversation status
func (h *InboxHandler) UpdateStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)

	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	var req dto.UpdateConversationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var conv models.Conversation
	if err := h.db.Where("id = ? AND user_id = ?", convID, userID).First(&conv).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	h.db.Model(&conv).Update("status", req.Status)
	c.JSON(http.StatusOK, gin.H{"message": "Status updated"})
}

// SyncInbox syncs conversations from marketplace
func (h *InboxHandler) SyncInbox(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.SyncInboxRequest
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

	connector, err := h.getConnector(userID, conn.Marketplace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get connector"})
		return
	}

	go h.inbox.SyncChatRooms(context.Background(), userID, conn.Marketplace, connector)

	c.JSON(http.StatusOK, gin.H{"message": "Inbox sync started"})
}

// ListTemplates returns message templates
func (h *InboxHandler) ListTemplates(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var templates []models.MessageTemplate
	h.db.Where("user_id = ?", userID).Order("title ASC").Find(&templates)

	result := make([]dto.MessageTemplateResponse, 0, len(templates))
	for _, t := range templates {
		result = append(result, dto.MessageTemplateResponse{
			ID:       t.ID.String(),
			Title:    t.Title,
			Body:     t.Body,
			Shortcut: t.Shortcut,
		})
	}

	c.JSON(http.StatusOK, result)
}

// CreateTemplate creates a message template
func (h *InboxHandler) CreateTemplate(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template := models.MessageTemplate{
		UserID:   userID,
		Title:    req.Title,
		Body:     req.Body,
		Shortcut: req.Shortcut,
	}

	if err := h.db.Create(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		return
	}

	c.JSON(http.StatusCreated, dto.MessageTemplateResponse{
		ID:       template.ID.String(),
		Title:    template.Title,
		Body:     template.Body,
		Shortcut: template.Shortcut,
	})
}

// DeleteTemplate deletes a message template
func (h *InboxHandler) DeleteTemplate(c *gin.Context) {
	userID := middleware.GetUserID(c)

	templateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	result := h.db.Where("id = ? AND user_id = ?", templateID, userID).Delete(&models.MessageTemplate{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template deleted"})
}

func (h *InboxHandler) getConnector(userID uuid.UUID, marketplace models.MarketplaceType) (*prom.Client, error) {
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
