package services

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sellflow/backend/internal/connectors"
	"github.com/sellflow/backend/internal/models"
	"gorm.io/gorm"
)

type InboxService struct {
	db *gorm.DB
}

func NewInboxService(db *gorm.DB) *InboxService {
	return &InboxService{db: db}
}

// SyncChatRooms fetches chat rooms from marketplace and syncs with local conversations
func (s *InboxService) SyncChatRooms(ctx context.Context, userID uuid.UUID, marketplace models.MarketplaceType, connector connectors.MarketplaceConnector) error {
	rooms, err := connector.GetChatRooms(ctx)
	if err != nil {
		return err
	}

	for _, room := range rooms {
		var conv models.Conversation
		err := s.db.Where("user_id = ? AND marketplace = ? AND external_id = ?",
			userID, marketplace, room.ExternalID).First(&conv).Error

		if err == gorm.ErrRecordNotFound {
			// Create new conversation
			conv = models.Conversation{
				UserID:       userID,
				Marketplace:  marketplace,
				ExternalID:   room.ExternalID,
				CustomerName: room.CustomerName,
				CustomerID:   room.CustomerID,
				Subject:      room.Subject,
				Status:       models.ConversationStatusNew,
			}
			s.db.Create(&conv)
		} else if err == nil {
			// Update customer name if changed
			if room.CustomerName != "" && room.CustomerName != conv.CustomerName {
				s.db.Model(&conv).Update("customer_name", room.CustomerName)
			}
		}

		// Sync messages for this room
		s.syncMessages(ctx, conv.ID, room.ExternalID, connector)
	}

	log.Printf("Synced %d chat rooms for user %s from %s", len(rooms), userID, marketplace)
	return nil
}

func (s *InboxService) syncMessages(ctx context.Context, convID uuid.UUID, chatRoomID string, connector connectors.MarketplaceConnector) {
	messages, err := connector.GetChatMessages(ctx, chatRoomID)
	if err != nil {
		log.Printf("Failed to sync messages for room %s: %v", chatRoomID, err)
		return
	}

	var latestTime *time.Time

	for _, msg := range messages {
		// Check if message already exists
		var count int64
		s.db.Model(&models.ConversationMessage{}).
			Where("conversation_id = ? AND external_id = ?", convID, msg.ExternalID).
			Count(&count)

		if count > 0 {
			continue
		}

		sender := models.SenderCustomer
		if msg.SenderType == "seller" {
			sender = models.SenderSeller
		}

		createdAt := parseTime(msg.CreatedAt)

		cm := models.ConversationMessage{
			ConversationID: convID,
			ExternalID:     msg.ExternalID,
			Sender:         sender,
			Body:           msg.Body,
			IsRead:         msg.IsRead,
			CreatedAt:      createdAt,
		}
		s.db.Create(&cm)

		if latestTime == nil || createdAt.After(*latestTime) {
			latestTime = &createdAt
		}
	}

	// Update last_message_at
	if latestTime != nil {
		s.db.Model(&models.Conversation{}).Where("id = ?", convID).
			Update("last_message_at", latestTime)
	}
}

func parseTime(s string) time.Time {
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t
		}
	}
	return time.Now()
}

// SendMessage sends a message via marketplace connector and stores locally
func (s *InboxService) SendMessage(ctx context.Context, convID uuid.UUID, body string, connector connectors.MarketplaceConnector) error {
	var conv models.Conversation
	if err := s.db.First(&conv, convID).Error; err != nil {
		return err
	}

	// Send via marketplace
	err := connector.SendMessage(ctx, connectors.SendMessageRequest{
		ChatRoomID: conv.ExternalID,
		Body:       body,
	})
	if err != nil {
		return err
	}

	// Store locally
	now := time.Now()
	s.db.Create(&models.ConversationMessage{
		ConversationID: convID,
		Sender:         models.SenderSeller,
		Body:           body,
		IsRead:         true,
		CreatedAt:      now,
	})

	// Update conversation
	s.db.Model(&conv).Updates(map[string]interface{}{
		"last_message_at": now,
		"status":          models.ConversationStatusInProgress,
	})

	return nil
}
