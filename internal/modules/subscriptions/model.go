package subscriptions

import (
	"database/sql"
	"time"
)

type Subscription struct {
	ID            string
	UserID        string
	CategoryID    sql.NullString
	CategoryName  sql.NullString
	Name          string
	Amount        float64
	Currency      string
	Cycle         string
	BillingDay    int
	NextBillingAt sql.NullTime
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (s Subscription) ToResponse() SubscriptionResponse {
	return SubscriptionResponse{
		ID:            s.ID,
		UserID:        s.UserID,
		CategoryID:    nullStringToPtr(s.CategoryID),
		CategoryName:  nullStringToPtr(s.CategoryName),
		Name:          s.Name,
		Amount:        s.Amount,
		Currency:      s.Currency,
		Cycle:         s.Cycle,
		BillingDay:    s.BillingDay,
		NextBillingAt: nullDateToPtr(s.NextBillingAt),
		IsActive:      s.IsActive,
		CreatedAt:     s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     s.UpdatedAt.Format(time.RFC3339),
	}
}

func nullStringToPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	return &value.String
}

func nullDateToPtr(value sql.NullTime) *string {
	if !value.Valid {
		return nil
	}

	formatted := value.Time.Format("2006-01-02")
	return &formatted
}
