package expenses

import (
	"database/sql"
	"time"
)

type Expense struct {
	ID           string
	UserID       string
	CategoryID   sql.NullString
	CategoryName sql.NullString
	Amount       float64
	Currency     string
	Note         sql.NullString
	RawText      sql.NullString
	Source       string
	SpentAt      time.Time
	AIConfidence sql.NullFloat64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (e Expense) ToResponse() ExpenseResponse {
	return ExpenseResponse{
		ID:           e.ID,
		UserID:       e.UserID,
		CategoryID:   nullStringToPtr(e.CategoryID),
		CategoryName: nullStringToPtr(e.CategoryName),
		Amount:       e.Amount,
		Currency:     e.Currency,
		Note:         nullStringToPtr(e.Note),
		RawText:      nullStringToPtr(e.RawText),
		Source:       e.Source,
		SpentAt:      e.SpentAt.Format(time.RFC3339),
		AIConfidence: nullFloat64ToPtr(e.AIConfidence),
		CreatedAt:    e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    e.UpdatedAt.Format(time.RFC3339),
	}
}

func nullStringToPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	return &value.String
}

func nullFloat64ToPtr(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}

	return &value.Float64
}