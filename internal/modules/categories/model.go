package categories

import (
	"database/sql"
	"time"
)

type Category struct {
	ID        string
	UserID    string
	Name      string
	Icon      sql.NullString
	Color     sql.NullString
	Type      string
	IsDefault bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c Category) ToResponse() CategoryResponse {
	return CategoryResponse{
		ID:        c.ID,
		UserID:    c.UserID,
		Name:      c.Name,
		Icon:      nullStringToPtr(c.Icon),
		Color:     nullStringToPtr(c.Color),
		Type:      c.Type,
		IsDefault: c.IsDefault,
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
		UpdatedAt: c.UpdatedAt.Format(time.RFC3339),
	}
}

func nullStringToPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	return &value.String
}