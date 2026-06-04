package expenses

type CreateExpenseRequest struct {
	CategoryID string  `json:"category_id"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Currency   string  `json:"currency"`
	Note       string  `json:"note"`
	RawText    string  `json:"raw_text"`
	Source     string  `json:"source"`
	SpentAt    string  `json:"spent_at"`
}

type UpdateExpenseRequest struct {
	CategoryID *string  `json:"category_id"`
	Amount     *float64 `json:"amount"`
	Currency   *string  `json:"currency"`
	Note       *string  `json:"note"`
	RawText    *string  `json:"raw_text"`
	Source     *string  `json:"source"`
	SpentAt    *string  `json:"spent_at"`
}

type ExpenseResponse struct {
	ID           string   `json:"id"`
	UserID       string   `json:"user_id"`
	CategoryID   *string  `json:"category_id"`
	CategoryName *string  `json:"category_name"`
	Amount       float64  `json:"amount"`
	Currency     string   `json:"currency"`
	Note         *string  `json:"note"`
	RawText      *string  `json:"raw_text"`
	Source       string   `json:"source"`
	SpentAt      string   `json:"spent_at"`
	AIConfidence *float64 `json:"ai_confidence"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

type ListExpensesQuery struct {
	Limit      int
	Cursor     string
	CategoryID string
	From       string
	To         string
	Search     string
}

type ListExpensesParams struct {
	UserID     string
	Limit      int
	Cursor     *string
	CategoryID *string
	From       *string
	To         *string
	Search     string
}

type ListExpensesResult struct {
	Items      []ExpenseResponse `json:"items"`
	NextCursor *string           `json:"next_cursor"`
	HasMore    bool              `json:"has_more"`
	Limit      int               `json:"limit"`
}

type CreateExpenseInput struct {
	UserID     string
	CategoryID *string
	Amount     float64
	Currency   string
	Note       *string
	RawText    *string
	Source     string
	SpentAt    string
}

type UpdateExpenseInput struct {
	ID         string
	UserID     string
	CategoryID *string
	Amount     *float64
	Currency   *string
	Note       *string
	RawText    *string
	Source     *string
	SpentAt    *string
}