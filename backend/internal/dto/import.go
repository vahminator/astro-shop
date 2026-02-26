package dto

type StartImportRequest struct {
	MarketplaceID string `json:"marketplace_id" binding:"required"`
}

type ImportStatusResponse struct {
	Total    int    `json:"total"`
	Imported int    `json:"imported"`
	Skipped  int    `json:"skipped"`
	Failed   int    `json:"failed"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}
