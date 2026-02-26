package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	ID       uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Name     string     `gorm:"not null" json:"name"`
	ParentID *uuid.UUID `gorm:"type:uuid" json:"parent_id"`
	Parent   *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type CategoryMarketplaceMapping struct {
	ID                   uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	CategoryID           uuid.UUID       `gorm:"type:uuid;not null;index" json:"category_id"`
	Category             Category        `gorm:"foreignKey:CategoryID" json:"-"`
	Marketplace          MarketplaceType `gorm:"type:varchar(50);not null" json:"marketplace"`
	ExternalCategoryID   string          `gorm:"not null" json:"external_category_id"`
	ExternalCategoryName string          `json:"external_category_name"`
}

func (cm *CategoryMarketplaceMapping) BeforeCreate(tx *gorm.DB) error {
	if cm.ID == uuid.Nil {
		cm.ID = uuid.New()
	}
	return nil
}
