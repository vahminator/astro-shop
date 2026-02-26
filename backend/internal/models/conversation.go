package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConversationStatus string

const (
	ConversationStatusNew        ConversationStatus = "new"
	ConversationStatusInProgress ConversationStatus = "in_progress"
	ConversationStatusWaiting    ConversationStatus = "waiting"
	ConversationStatusClosed     ConversationStatus = "closed"
)

type Conversation struct {
	ID            uuid.UUID             `gorm:"type:uuid;primaryKey" json:"id"`
	UserID        uuid.UUID             `gorm:"type:uuid;not null;index" json:"user_id"`
	Marketplace   MarketplaceType       `gorm:"type:varchar(50);not null" json:"marketplace"`
	ExternalID    string                `gorm:"index" json:"external_id"`
	CustomerName  string                `json:"customer_name"`
	CustomerID    string                `json:"customer_id"`
	ProductID     *uuid.UUID            `gorm:"type:uuid" json:"product_id"`
	Product       *Product              `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Status        ConversationStatus    `gorm:"type:varchar(20);default:'new'" json:"status"`
	Subject       string                `json:"subject"`
	LastMessageAt *time.Time            `json:"last_message_at"`
	Messages      []ConversationMessage `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
}

func (c *Conversation) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type SenderType string

const (
	SenderCustomer SenderType = "customer"
	SenderSeller   SenderType = "seller"
)

type ConversationMessage struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	ConversationID uuid.UUID  `gorm:"type:uuid;not null;index" json:"conversation_id"`
	ExternalID     string     `json:"external_id"`
	Sender         SenderType `gorm:"type:varchar(20);not null" json:"sender"`
	Body           string     `gorm:"type:text;not null" json:"body"`
	IsRead         bool       `gorm:"default:false" json:"is_read"`
	CreatedAt      time.Time  `json:"created_at"`
}

func (cm *ConversationMessage) BeforeCreate(tx *gorm.DB) error {
	if cm.ID == uuid.Nil {
		cm.ID = uuid.New()
	}
	return nil
}

type MessageTemplate struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID   uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Title    string    `gorm:"not null" json:"title"`
	Body     string    `gorm:"type:text;not null" json:"body"`
	Shortcut string    `json:"shortcut"` // e.g. /delivery, /warranty
}

func (mt *MessageTemplate) BeforeCreate(tx *gorm.DB) error {
	if mt.ID == uuid.Nil {
		mt.ID = uuid.New()
	}
	return nil
}
