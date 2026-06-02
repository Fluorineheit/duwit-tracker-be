package categories

type CreateCategoryRequest struct {
	Name  string `json:"name" binding:"required"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
	Type  string `json:"type"`
}

type UpdateCategoryRequest struct {
	Name  *string `json:"name"`
	Icon  *string `json:"icon"`
	Color *string `json:"color"`
	Type  *string `json:"type"`
}

type CategoryResponse struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Name      string  `json:"name"`
	Icon      *string `json:"icon"`
	Color     *string `json:"color"`
	Type      string  `json:"type"`
	IsDefault bool    `json:"is_default"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type ListCategoriesQuery struct {
	Type string
}

type CreateCategoryInput struct {
	UserID string
	Name   string
	Icon   *string
	Color  *string
	Type   string
}

type UpdateCategoryInput struct {
	ID     string
	UserID string
	Name   *string
	Icon   *string
	Color  *string
	Type   *string
}