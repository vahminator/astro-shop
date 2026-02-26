package dto

type ConnectMarketplaceRequest struct {
	Marketplace string `json:"marketplace" binding:"required"`
	APIKey      string `json:"api_key" binding:"required"`
	ShopURL     string `json:"shop_url"`
}

type UpdateMarketplaceRequest struct {
	APIKey   string `json:"api_key"`
	ShopURL  string `json:"shop_url"`
	IsActive *bool  `json:"is_active"`
}

type MarketplaceConnectionResponse struct {
	ID          string `json:"id"`
	Marketplace string `json:"marketplace"`
	ShopURL     string `json:"shop_url"`
	ShopName    string `json:"shop_name"`
	IsActive    bool   `json:"is_active"`
	LastSyncAt  string `json:"last_sync_at,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type TestConnectionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
