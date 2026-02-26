package dto

type ReviewResponse struct {
	ID             string                   `json:"id"`
	Marketplace    string                   `json:"marketplace"`
	ExternalID     string                   `json:"external_id"`
	ProductID      *string                  `json:"product_id"`
	ProductTitle   string                   `json:"product_title"`
	AuthorName     string                   `json:"author_name"`
	Rating         int                      `json:"rating"`
	Body           string                   `json:"body"`
	Status         string                   `json:"status"`
	Responses      []ReviewResponseItem     `json:"responses"`
	CreatedAt      string                   `json:"created_at"`
}

type ReviewResponseItem struct {
	ID        string `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

type ReviewsListResponse struct {
	Reviews []ReviewResponse `json:"reviews"`
	Total   int64            `json:"total"`
	Page    int              `json:"page"`
	Limit   int              `json:"limit"`
}

type CreateReviewRequest struct {
	Marketplace string  `json:"marketplace" binding:"required"`
	ExternalID  string  `json:"external_id"`
	ProductID   *string `json:"product_id"`
	AuthorName  string  `json:"author_name" binding:"required"`
	Rating      int     `json:"rating" binding:"required,min=1,max=5"`
	Body        string  `json:"body"`
}

type ReplyToReviewRequest struct {
	Body string `json:"body" binding:"required"`
}

type UpdateReviewStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=new responded resolved"`
}

type SyncReviewsRequest struct {
	MarketplaceID string `json:"marketplace_id" binding:"required"`
}

type ReviewStatsResponse struct {
	TotalReviews   int64          `json:"total_reviews"`
	AverageRating  float64        `json:"average_rating"`
	RatingBreakdown map[int]int64 `json:"rating_breakdown"`
	NewCount       int64          `json:"new_count"`
	RespondedCount int64          `json:"responded_count"`
	ResolvedCount  int64          `json:"resolved_count"`
}
