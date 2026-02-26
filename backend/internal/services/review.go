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

type ReviewService struct {
	db *gorm.DB
}

func NewReviewService(db *gorm.DB) *ReviewService {
	return &ReviewService{db: db}
}

// SyncReviews fetches messages from marketplace that may contain reviews/feedback
func (s *ReviewService) SyncReviews(ctx context.Context, userID uuid.UUID, marketplace models.MarketplaceType, connector connectors.MarketplaceConnector) error {
	// Fetch messages from legacy API (may contain reviews/feedback)
	messages, err := connector.GetMessages(ctx, 1, 100)
	if err != nil {
		return fmt.Errorf("failed to fetch messages: %w", err)
	}

	for _, msg := range messages {
		// Check if review with this external_id already exists
		var existing models.Review
		if err := s.db.Where("user_id = ? AND marketplace = ? AND external_id = ?",
			userID, marketplace, msg.ExternalID).First(&existing).Error; err == nil {
			continue // Already exists
		}

		review := models.Review{
			UserID:      userID,
			Marketplace: marketplace,
			ExternalID:  msg.ExternalID,
			AuthorName:  msg.SenderName,
			Rating:      0, // Messages don't have ratings, will need manual assignment
			Body:        msg.Body,
			Status:      models.ReviewStatusNew,
		}

		if msg.Subject != "" {
			review.Body = msg.Subject + "\n\n" + msg.Body
		}

		if err := s.db.Create(&review).Error; err != nil {
			log.Printf("Failed to create review from message %s: %v", msg.ExternalID, err)
			continue
		}
	}

	// Log the sync
	s.db.Create(&models.SyncLog{
		UserID:     userID,
		Marketplace: marketplace,
		EntityType: "review",
		Action:     "sync",
		Status:     "success",
	})

	return nil
}

// ReplyToReview creates a response and optionally sends it via marketplace
func (s *ReviewService) ReplyToReview(ctx context.Context, reviewID uuid.UUID, body string, connector connectors.MarketplaceConnector) error {
	var review models.Review
	if err := s.db.First(&review, reviewID).Error; err != nil {
		return fmt.Errorf("review not found: %w", err)
	}

	// Try to send reply via marketplace
	if connector != nil && review.ExternalID != "" {
		if err := connector.ReplyToMessage(ctx, review.ExternalID, body); err != nil {
			log.Printf("Failed to send reply to marketplace: %v", err)
			// Don't fail — still save locally
		}
	}

	// Save response locally
	response := models.ReviewResponse{
		ReviewID:  reviewID,
		Body:      body,
		CreatedAt: time.Now(),
	}
	if err := s.db.Create(&response).Error; err != nil {
		return fmt.Errorf("failed to save response: %w", err)
	}

	// Update review status
	s.db.Model(&review).Update("status", models.ReviewStatusResponded)

	return nil
}

// GetStats returns review statistics for a user
func (s *ReviewService) GetStats(userID uuid.UUID) (totalReviews int64, avgRating float64, ratingBreakdown map[int]int64, statusCounts map[string]int64) {
	s.db.Model(&models.Review{}).Where("user_id = ?", userID).Count(&totalReviews)

	var avg *float64
	s.db.Model(&models.Review{}).Where("user_id = ? AND rating > 0", userID).
		Select("AVG(rating)").Scan(&avg)
	if avg != nil {
		avgRating = *avg
	}

	ratingBreakdown = make(map[int]int64)
	for i := 1; i <= 5; i++ {
		var count int64
		s.db.Model(&models.Review{}).Where("user_id = ? AND rating = ?", userID, i).Count(&count)
		ratingBreakdown[i] = count
	}

	statusCounts = make(map[string]int64)
	for _, status := range []string{"new", "responded", "resolved"} {
		var count int64
		s.db.Model(&models.Review{}).Where("user_id = ? AND status = ?", userID, status).Count(&count)
		statusCounts[status] = count
	}

	return
}
