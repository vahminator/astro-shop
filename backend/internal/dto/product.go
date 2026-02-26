package dto

type CreateProductRequest struct {
	Title       string            `json:"title" binding:"required"`
	Description string            `json:"description"`
	SKU         string            `json:"sku"`
	EAN         string            `json:"ean"`
	BasePrice   float64           `json:"base_price"`
	SalePrice   *float64          `json:"sale_price"`
	Stock       int               `json:"stock"`
	Status      string            `json:"status"`
	CategoryID  *string           `json:"category_id"`
	Attributes  []AttributeInput  `json:"attributes"`
}

type UpdateProductRequest struct {
	Title       *string           `json:"title"`
	Description *string           `json:"description"`
	SKU         *string           `json:"sku"`
	EAN         *string           `json:"ean"`
	BasePrice   *float64          `json:"base_price"`
	SalePrice   *float64          `json:"sale_price"`
	Stock       *int              `json:"stock"`
	Status      *string           `json:"status"`
	CategoryID  *string           `json:"category_id"`
	Attributes  []AttributeInput  `json:"attributes"`
}

type AttributeInput struct {
	Name  string `json:"name" binding:"required"`
	Value string `json:"value" binding:"required"`
}
