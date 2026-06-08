package subscriptions

type CreateSubscriptionRequest struct {
	Name       string  `json:"name" binding:"required"`
	CategoryID string  `json:"category_id"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Currency   string  `json:"currency"`
	Cycle      string  `json:"cycle"`
	BillingDay int     `json:"billing_day" binding:"required,gte=1,lte=31"`
}

type UpdateSubscriptionRequest struct {
	Name       *string  `json:"name"`
	CategoryID *string  `json:"category_id"`
	Amount     *float64 `json:"amount"`
	Currency   *string  `json:"currency"`
	Cycle      *string  `json:"cycle"`
	BillingDay *int     `json:"billing_day"`
	IsActive   *bool    `json:"is_active"`
}

type SubscriptionResponse struct {
	ID            string  `json:"id"`
	UserID        string  `json:"user_id"`
	CategoryID    *string `json:"category_id"`
	CategoryName  *string `json:"category_name"`
	Name          string  `json:"name"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Cycle         string  `json:"cycle"`
	BillingDay    int     `json:"billing_day"`
	NextBillingAt *string `json:"next_billing_at"`
	IsActive      bool    `json:"is_active"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type ListSubscriptionsResult struct {
	Items []SubscriptionResponse `json:"items"`
}

type CreateSubscriptionInput struct {
	UserID        string
	CategoryID    *string
	Name          string
	Amount        float64
	Currency      string
	Cycle         string
	BillingDay    int
	NextBillingAt string
}

type UpdateSubscriptionInput struct {
	ID            string
	UserID        string
	CategoryID    *string
	Name          *string
	Amount        *float64
	Currency      *string
	Cycle         *string
	BillingDay    *int
	NextBillingAt *string
	IsActive      *bool
}
