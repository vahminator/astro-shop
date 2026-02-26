package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductStatus string

const (
	ProductStatusDraft    ProductStatus = "draft"
	ProductStatusActive   ProductStatus = "active"
	ProductStatusArchived ProductStatus = "archived"
)

type Product struct {
	ID              uuid.UUID                `gorm:"type:uuid;primaryKey" json:"id"`
	UserID          uuid.UUID                `gorm:"type:uuid;not null;index" json:"user_id"`
	User            User                     `gorm:"foreignKey:UserID" json:"-"`
	Title           string                   `gorm:"not null" json:"title"`
	Description     string                   `gorm:"type:text" json:"description"`
	SKU             string                   `gorm:"index" json:"sku"`
	EAN             string                   `json:"ean"`
	BasePrice       float64                  `gorm:"type:decimal(12,2)" json:"base_price"`
	SalePrice       *float64                 `gorm:"type:decimal(12,2)" json:"sale_price"`
	Stock           int                      `gorm:"default:0" json:"stock"`
	Status          ProductStatus            `gorm:"type:varchar(20);default:'draft'" json:"status"`
	CategoryID      *uuid.UUID               `gorm:"type:uuid" json:"category_id"`
	Category        *Category                `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Images          []ProductImage           `gorm:"foreignKey:ProductID" json:"images"`
	Attributes      []ProductAttribute       `gorm:"foreignKey:ProductID" json:"attributes"`
	Variants        []ProductVariant         `gorm:"foreignKey:ProductID" json:"variants"`
	MarketplaceData []ProductMarketplaceData `gorm:"foreignKey:ProductID" json:"marketplace_data"`
	CreatedAt       time.Time                `json:"created_at"`
	UpdatedAt       time.Time                `json:"updated_at"`
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type ProductImage struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProductID  uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	URL        string    `gorm:"not null" json:"url"`
	StorageKey string    `json:"storage_key"`
	Position   int       `gorm:"default:0" json:"position"`
	CreatedAt  time.Time `json:"created_at"`
}

func (pi *ProductImage) BeforeCreate(tx *gorm.DB) error {
	if pi.ID == uuid.Nil {
		pi.ID = uuid.New()
	}
	return nil
}

type ProductAttribute struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	Name      string    `gorm:"not null" json:"name"`
	Value     string    `gorm:"not null" json:"value"`
}

func (pa *ProductAttribute) BeforeCreate(tx *gorm.DB) error {
	if pa.ID == uuid.Nil {
		pa.ID = uuid.New()
	}
	return nil
}

type ProductVariant struct {
	ID         uuid.UUID              `gorm:"type:uuid;primaryKey" json:"id"`
	ProductID  uuid.UUID              `gorm:"type:uuid;not null;index" json:"product_id"`
	Title      string                 `gorm:"not null" json:"title"`
	SKU        string                 `json:"sku"`
	Price      float64                `gorm:"type:decimal(12,2)" json:"price"`
	Stock      int                    `gorm:"default:0" json:"stock"`
	Attributes map[string]interface{} `gorm:"type:jsonb;serializer:json" json:"attributes"`
}

func (pv *ProductVariant) BeforeCreate(tx *gorm.DB) error {
	if pv.ID == uuid.Nil {
		pv.ID = uuid.New()
	}
	return nil
}

type PublishStatus string

const (
	PublishStatusPending    PublishStatus = "pending"
	PublishStatusPublished  PublishStatus = "published"
	PublishStatusError      PublishStatus = "error"
	PublishStatusModeration PublishStatus = "moderation"
	PublishStatusRejected   PublishStatus = "rejected"
)

type ProductMarketplaceData struct {
	ID                uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	ProductID         uuid.UUID       `gorm:"type:uuid;not null;index" json:"product_id"`
	Marketplace       MarketplaceType `gorm:"type:varchar(50);not null" json:"marketplace"`
	ExternalID        string          `gorm:"index" json:"external_id"`
	MarketplaceURL    string          `json:"marketplace_url"`
	CustomTitle       string          `json:"custom_title"`
	CustomDescription string          `gorm:"type:text" json:"custom_description"`
	CustomPrice       *float64        `gorm:"type:decimal(12,2)" json:"custom_price"`
	PriceMarkup       *float64        `gorm:"type:decimal(5,2)" json:"price_markup"` // percentage
	Status            PublishStatus   `gorm:"type:varchar(20);default:'pending'" json:"status"`
	LastSyncedAt      *time.Time      `json:"last_synced_at"`
	ErrorMessage      string          `json:"error_message"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

func (pmd *ProductMarketplaceData) BeforeCreate(tx *gorm.DB) error {
	if pmd.ID == uuid.Nil {
		pmd.ID = uuid.New()
	}
	return nil
}
