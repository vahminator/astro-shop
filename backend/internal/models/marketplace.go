package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MarketplaceType string

const (
	MarketplaceProm      MarketplaceType = "prom"
	MarketplaceRozetka   MarketplaceType = "rozetka"
	MarketplaceEpicenter MarketplaceType = "epicenter"
	MarketplaceOLX       MarketplaceType = "olx"
)

type MarketplaceConnection struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	UserID          uuid.UUID       `gorm:"type:uuid;not null;index" json:"user_id"`
	User            User            `gorm:"foreignKey:UserID" json:"-"`
	Marketplace     MarketplaceType `gorm:"type:varchar(50);not null" json:"marketplace"`
	APIKeyEncrypted string          `gorm:"not null" json:"-"`
	ShopURL         string          `json:"shop_url"`
	ShopName        string          `json:"shop_name"`
	IsActive        bool            `gorm:"default:true" json:"is_active"`
	LastSyncAt      *time.Time      `json:"last_sync_at"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

func (m *MarketplaceConnection) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
