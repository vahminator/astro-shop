package prom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sellflow/backend/internal/connectors"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) ([]byte, error) {
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("prom API error %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}

// --- TestConnection ---

func (c *Client) TestConnection(ctx context.Context) error {
	_, err := c.doRequest(ctx, "GET", "/products/list?limit=1", nil)
	return err
}

// --- Products ---

type promProductsResponse struct {
	Products []promProduct `json:"products"`
	Group    interface{}   `json:"group"`
}

type promProduct struct {
	ID              int             `json:"id"`
	ExternalID      string          `json:"external_id"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	Price           float64         `json:"price"`
	MinimumPrice    float64         `json:"minimum_order_price"`
	Currency        string          `json:"currency"`
	QuantityInStock interface{}     `json:"quantity_in_stock"` // can be string or int
	SKU             string          `json:"sku"`
	URL             string          `json:"prom_url"`
	MainImage       string          `json:"main_image"`
	Images          []promImage     `json:"images"`
	Group           *promGroup      `json:"group"`
	Status          string          `json:"status"`
	Attributes      []promAttribute `json:"attributes"`
}

type promImage struct {
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url"`
}

type promGroup struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type promAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (c *Client) GetProducts(ctx context.Context, page, limit int) ([]connectors.MarketplaceProduct, int, error) {
	offset := (page - 1) * limit
	path := fmt.Sprintf("/products/list?limit=%d&offset=%d", limit, offset)

	data, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, 0, err
	}

	var resp promProductsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, 0, err
	}

	products := make([]connectors.MarketplaceProduct, 0, len(resp.Products))
	for _, p := range resp.Products {
		mp := connectors.MarketplaceProduct{
			ExternalID:  fmt.Sprintf("%d", p.ID),
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			SKU:         p.SKU,
			URL:         p.URL,
			Status:      p.Status,
		}

		// Parse stock
		switch v := p.QuantityInStock.(type) {
		case float64:
			mp.Stock = int(v)
		case string:
			// "unlimited" or similar
			mp.Stock = -1
		}

		// Images
		if p.MainImage != "" {
			mp.Images = append(mp.Images, p.MainImage)
		}
		for _, img := range p.Images {
			if img.URL != "" && img.URL != p.MainImage {
				mp.Images = append(mp.Images, img.URL)
			}
		}

		// Category
		if p.Group != nil {
			mp.CategoryID = fmt.Sprintf("%d", p.Group.ID)
			mp.CategoryName = p.Group.Name
		}

		// Attributes
		if len(p.Attributes) > 0 {
			mp.Attributes = make(map[string]string)
			for _, attr := range p.Attributes {
				mp.Attributes[attr.Name] = attr.Value
			}
		}

		products = append(products, mp)
	}

	return products, len(products), nil
}

func (c *Client) GetProduct(ctx context.Context, externalID string) (*connectors.MarketplaceProduct, error) {
	path := fmt.Sprintf("/products/%s", externalID)
	data, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Product promProduct `json:"product"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	p := resp.Product
	mp := &connectors.MarketplaceProduct{
		ExternalID:  fmt.Sprintf("%d", p.ID),
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		SKU:         p.SKU,
		URL:         p.URL,
		Status:      p.Status,
	}

	switch v := p.QuantityInStock.(type) {
	case float64:
		mp.Stock = int(v)
	case string:
		mp.Stock = -1
	}

	if p.MainImage != "" {
		mp.Images = append(mp.Images, p.MainImage)
	}
	for _, img := range p.Images {
		if img.URL != "" && img.URL != p.MainImage {
			mp.Images = append(mp.Images, img.URL)
		}
	}

	if p.Group != nil {
		mp.CategoryID = fmt.Sprintf("%d", p.Group.ID)
		mp.CategoryName = p.Group.Name
	}

	if len(p.Attributes) > 0 {
		mp.Attributes = make(map[string]string)
		for _, attr := range p.Attributes {
			mp.Attributes[attr.Name] = attr.Value
		}
	}

	return mp, nil
}

func (c *Client) UpdateProduct(ctx context.Context, externalID string, updates map[string]interface{}) error {
	updates["id"] = externalID
	body, err := json.Marshal(updates)
	if err != nil {
		return err
	}
	_, err = c.doRequest(ctx, "POST", "/products/edit", strings.NewReader(string(body)))
	return err
}

// --- Categories ---

type promGroupsResponse struct {
	Groups []promGroupFull `json:"groups"`
}

type promGroupFull struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ParentID int    `json:"parent_group_id"`
}

func (c *Client) GetCategories(ctx context.Context) ([]connectors.MarketplaceCategory, error) {
	data, err := c.doRequest(ctx, "GET", "/groups/list", nil)
	if err != nil {
		return nil, err
	}

	var resp promGroupsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	categories := make([]connectors.MarketplaceCategory, 0, len(resp.Groups))
	for _, g := range resp.Groups {
		cat := connectors.MarketplaceCategory{
			ExternalID: fmt.Sprintf("%d", g.ID),
			Name:       g.Name,
		}
		if g.ParentID > 0 {
			cat.ParentID = fmt.Sprintf("%d", g.ParentID)
		}
		categories = append(categories, cat)
	}

	return categories, nil
}

// --- Chat ---

type promChatRoomsResponse struct {
	ChatRooms []promChatRoom `json:"chat_rooms"`
}

type promChatRoom struct {
	ID           int    `json:"id"`
	Caption      string `json:"caption"`
	CustomerName string `json:"customer_name"`
	CustomerID   int    `json:"customer_id"`
	UpdatedAt    string `json:"updated_at"`
	UnreadCount  int    `json:"unread_count"`
}

func (c *Client) GetChatRooms(ctx context.Context) ([]connectors.MarketplaceChatRoom, error) {
	data, err := c.doRequest(ctx, "GET", "/chat/rooms", nil)
	if err != nil {
		return nil, err
	}

	var resp promChatRoomsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	rooms := make([]connectors.MarketplaceChatRoom, 0, len(resp.ChatRooms))
	for _, r := range resp.ChatRooms {
		rooms = append(rooms, connectors.MarketplaceChatRoom{
			ExternalID:   fmt.Sprintf("%d", r.ID),
			CustomerName: r.CustomerName,
			CustomerID:   fmt.Sprintf("%d", r.CustomerID),
			Subject:      r.Caption,
			UpdatedAt:    r.UpdatedAt,
			UnreadCount:  r.UnreadCount,
		})
	}

	return rooms, nil
}

type promChatMessagesResponse struct {
	Messages []promChatMessage `json:"messages"`
}

type promChatMessage struct {
	ID        int    `json:"id"`
	Body      string `json:"body"`
	Type      string `json:"type"` // "seller" or "customer"
	CreatedAt string `json:"created_at"`
	IsRead    bool   `json:"is_read"`
}

func (c *Client) GetChatMessages(ctx context.Context, chatRoomID string) ([]connectors.MarketplaceMessage, error) {
	path := fmt.Sprintf("/chat/messages_history?chat_id=%s", chatRoomID)
	data, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp promChatMessagesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	messages := make([]connectors.MarketplaceMessage, 0, len(resp.Messages))
	for _, m := range resp.Messages {
		sender := "customer"
		if m.Type == "seller" {
			sender = "seller"
		}
		messages = append(messages, connectors.MarketplaceMessage{
			ExternalID: fmt.Sprintf("%d", m.ID),
			Body:       m.Body,
			SenderType: sender,
			ChatRoomID: chatRoomID,
			CreatedAt:  m.CreatedAt,
			IsRead:     m.IsRead,
		})
	}

	return messages, nil
}

func (c *Client) SendMessage(ctx context.Context, req connectors.SendMessageRequest) error {
	payload := map[string]string{
		"chat_id": req.ChatRoomID,
		"body":    req.Body,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = c.doRequest(ctx, "POST", "/chat/send_message", strings.NewReader(string(body)))
	return err
}

// --- Messages (legacy) ---

type promMessagesResponse struct {
	Messages []promMessage `json:"messages"`
}

type promMessage struct {
	ID        int    `json:"id"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	Status    string `json:"status"`
	ClientID  int    `json:"client_id"`
	CreatedAt string `json:"date_created"`
}

func (c *Client) GetMessages(ctx context.Context, page, limit int) ([]connectors.MarketplaceMessage, error) {
	offset := (page - 1) * limit
	path := fmt.Sprintf("/messages/list?limit=%d&offset=%d", limit, offset)
	data, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp promMessagesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	messages := make([]connectors.MarketplaceMessage, 0, len(resp.Messages))
	for _, m := range resp.Messages {
		messages = append(messages, connectors.MarketplaceMessage{
			ExternalID: fmt.Sprintf("%d", m.ID),
			Subject:    m.Subject,
			Body:       m.Body,
			SenderName: fmt.Sprintf("Client #%d", m.ClientID),
			SenderType: "customer",
			CreatedAt:  m.CreatedAt,
		})
	}

	return messages, nil
}

func (c *Client) ReplyToMessage(ctx context.Context, messageID string, body string) error {
	payload := map[string]string{
		"id":   messageID,
		"body": body,
	}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = c.doRequest(ctx, "POST", "/messages/reply", strings.NewReader(string(jsonBody)))
	return err
}
