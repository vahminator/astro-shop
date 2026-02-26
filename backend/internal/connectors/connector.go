package connectors

import "context"

// MarketplaceProduct represents a product from a marketplace
type MarketplaceProduct struct {
	ExternalID   string
	Name         string
	Description  string
	Price        float64
	Stock        int
	SKU          string
	URL          string
	Images       []string
	CategoryID   string
	CategoryName string
	Attributes   map[string]string
	Status       string
}

type MarketplaceCategory struct {
	ExternalID string
	Name       string
	ParentID   string
}

type MarketplaceMessage struct {
	ExternalID string
	Subject    string
	Body       string
	SenderName string
	SenderType string // "customer" or "seller"
	ChatRoomID string
	ProductID  string
	CreatedAt  string
	IsRead     bool
}

type MarketplaceChatRoom struct {
	ExternalID   string
	CustomerName string
	CustomerID   string
	Subject      string
	LastMessage  string
	UpdatedAt    string
	UnreadCount  int
}

type SendMessageRequest struct {
	ChatRoomID string
	Body       string
}

// MarketplaceConnector defines the interface for marketplace integrations
type MarketplaceConnector interface {
	// TestConnection verifies the API key is valid
	TestConnection(ctx context.Context) error

	// Products
	GetProducts(ctx context.Context, page, limit int) ([]MarketplaceProduct, int, error)
	GetProduct(ctx context.Context, externalID string) (*MarketplaceProduct, error)
	UpdateProduct(ctx context.Context, externalID string, data map[string]interface{}) error

	// Categories
	GetCategories(ctx context.Context) ([]MarketplaceCategory, error)

	// Messages / Chat
	GetChatRooms(ctx context.Context) ([]MarketplaceChatRoom, error)
	GetChatMessages(ctx context.Context, chatRoomID string) ([]MarketplaceMessage, error)
	SendMessage(ctx context.Context, req SendMessageRequest) error

	// Messages (legacy API)
	GetMessages(ctx context.Context, page, limit int) ([]MarketplaceMessage, error)
	ReplyToMessage(ctx context.Context, messageID string, body string) error
}
