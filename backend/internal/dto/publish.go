package dto

type PublishRequest struct {
	MarketplaceID string `json:"marketplace_id" binding:"required"`
}

type SyncProductRequest struct {
	MarketplaceID string `json:"marketplace_id" binding:"required"`
}

type PublishQueueItemResponse struct {
	ID          string  `json:"id"`
	ProductID   string  `json:"product_id"`
	Marketplace string  `json:"marketplace"`
	Action      string  `json:"action"`
	Status      string  `json:"status"`
	RetryCount  int     `json:"retry_count"`
	Error       string  `json:"error,omitempty"`
	CreatedAt   string  `json:"created_at"`
	ProcessedAt *string `json:"processed_at,omitempty"`
}

type SyncLogResponse struct {
	ID          string `json:"id"`
	Marketplace string `json:"marketplace"`
	EntityType  string `json:"entity_type"`
	EntityID    string `json:"entity_id"`
	Action      string `json:"action"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
	CreatedAt   string `json:"created_at"`
}
