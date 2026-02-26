package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReviewStatus string

const (
	ReviewStatusNew       ReviewStatus = "new"
	ReviewStatusResponded ReviewStatus = "responded"
	ReviewStatusResolved  ReviewStatus = "resolved"
)

type Review struct {
	ID          uuid.UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID        `gorm:"type:uuid;not null;index" json:"user_id"`
	Marketplace MarketplaceType  `gorm:"type:varchar(50);not null" json:"marketplace"`
	ExternalID  string           `gorm:"index" json:"external_id"`
	ProductID   *uuid.UUID       `gorm:"type:uuid" json:"product_id"`
	Product     *Product         `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	AuthorName  string           `json:"author_name"`
	Rating      int              `gorm:"not null" json:"rating"` // 1-5
	Body        string           `gorm:"type:text" json:"body"`
	Status      ReviewStatus     `gorm:"type:varchar(20);default:'new'" json:"status"`
	Responses   []ReviewResponse `gorm:"foreignKey:ReviewID" json:"responses"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

func (r *Review) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

type ReviewResponse struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ReviewID  uuid.UUID `gorm:"type:uuid;not null;index" json:"review_id"`
	Body      string    `gorm:"type:text;not null" json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

func (rr *ReviewResponse) BeforeCreate(tx *gorm.DB) error {
	if rr.ID == uuid.Nil {
		rr.ID = uuid.New()
	}
	return nil
}
