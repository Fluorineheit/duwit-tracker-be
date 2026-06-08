package budgets

type CreateBudgetRequest struct {
	CategoryID string  `json:"category_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Currency   string  `json:"currency"`
	Month      string  `json:"month" binding:"required"`
}

type BudgetResponse struct {
	ID           string  `json:"id"`
	UserID       string  `json:"user_id"`
	CategoryID   *string `json:"category_id"`
	CategoryName *string `json:"category_name"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	Month        string  `json:"month"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type ListBudgetsQuery struct {
	Month string
}

type ListBudgetsParams struct {
	UserID string
	Month  *string
}

type ListBudgetsResult struct {
	Items []BudgetResponse `json:"items"`
}

type CreateBudgetInput struct {
	UserID     string
	CategoryID string
	Amount     float64
	Currency   string
	Month      string
}
