package dto

type ConversationResponse struct {
	ID            string             `json:"id"`
	Marketplace   string             `json:"marketplace"`
	ExternalID    string             `json:"external_id"`
	CustomerName  string             `json:"customer_name"`
	CustomerID    string             `json:"customer_id"`
	Subject       string             `json:"subject"`
	Status        string             `json:"status"`
	LastMessageAt *string            `json:"last_message_at"`
	LastMessage   *string            `json:"last_message,omitempty"`
	UnreadCount   int                `json:"unread_count"`
	CreatedAt     string             `json:"created_at"`
}

type ConversationDetailResponse struct {
	ConversationResponse
	Messages []MessageResponse `json:"messages"`
}

type MessageResponse struct {
	ID         string `json:"id"`
	ExternalID string `json:"external_id"`
	Sender     string `json:"sender"`
	Body       string `json:"body"`
	IsRead     bool   `json:"is_read"`
	CreatedAt  string `json:"created_at"`
}

type SendMessageRequest struct {
	Body string `json:"body" binding:"required"`
}

type UpdateConversationStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type SyncInboxRequest struct {
	MarketplaceID string `json:"marketplace_id" binding:"required"`
}

type MessageTemplateResponse struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	Shortcut string `json:"shortcut"`
}

type CreateTemplateRequest struct {
	Title    string `json:"title" binding:"required"`
	Body     string `json:"body" binding:"required"`
	Shortcut string `json:"shortcut"`
}
