package budgets

import (
	"database/sql"
	"time"
)

type Budget struct {
	ID           string
	UserID       string
	CategoryID   sql.NullString
	CategoryName sql.NullString
	Amount       float64
	Currency     string
	Month        time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (b Budget) ToResponse() BudgetResponse {
	return BudgetResponse{
		ID:           b.ID,
		UserID:       b.UserID,
		CategoryID:   nullStringToPtr(b.CategoryID),
		CategoryName: nullStringToPtr(b.CategoryName),
		Amount:       b.Amount,
		Currency:     b.Currency,
		Month:        b.Month.Format("2006-01-02"),
		CreatedAt:    b.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    b.UpdatedAt.Format(time.RFC3339),
	}
}

func nullStringToPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	return &value.String
}
