package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SyncLog struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	UserID       uuid.UUID       `gorm:"type:uuid;not null;index" json:"user_id"`
	Marketplace  MarketplaceType `gorm:"type:varchar(50);not null" json:"marketplace"`
	EntityType   string          `gorm:"not null" json:"entity_type"` // product, price, stock, message, review
	EntityID     string          `json:"entity_id"`
	Action       string          `gorm:"not null" json:"action"` // import, publish, update, sync
	Status       string          `gorm:"not null" json:"status"` // success, error
	ErrorMessage string          `json:"error_message"`
	CreatedAt    time.Time       `json:"created_at"`
}

func (s *SyncLog) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type QueueAction string

const (
	QueueActionPublish   QueueAction = "publish"
	QueueActionUpdate    QueueAction = "update"
	QueueActionUnpublish QueueAction = "unpublish"
)

type QueueStatus string

const (
	QueueStatusPending    QueueStatus = "pending"
	QueueStatusProcessing QueueStatus = "processing"
	QueueStatusDone       QueueStatus = "done"
	QueueStatusFailed     QueueStatus = "failed"
)

type PublishQueue struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	ProductID    uuid.UUID       `gorm:"type:uuid;not null;index" json:"product_id"`
	Product      Product         `gorm:"foreignKey:ProductID" json:"-"`
	Marketplace  MarketplaceType `gorm:"type:varchar(50);not null" json:"marketplace"`
	Action       QueueAction     `gorm:"type:varchar(20);not null" json:"action"`
	Status       QueueStatus     `gorm:"type:varchar(20);default:'pending'" json:"status"`
	RetryCount   int             `gorm:"default:0" json:"retry_count"`
	ErrorMessage string          `json:"error_message"`
	CreatedAt    time.Time       `json:"created_at"`
	ProcessedAt  *time.Time      `json:"processed_at"`
}

func (pq *PublishQueue) BeforeCreate(tx *gorm.DB) error {
	if pq.ID == uuid.Nil {
		pq.ID = uuid.New()
	}
	return nil
}
