package reports

type DailyReportQuery struct {
	DateFrom string
	DateTo   string
}

type MonthlyReportQuery struct {
	Year string
}

type CategoryReportQuery struct {
	DateFrom string
	DateTo   string
}

type DailyReportParams struct {
	UserID string
	From   string
	To     string
}

type MonthlyReportParams struct {
	UserID string
	From   string
	To     string
}

type CategoryReportParams struct {
	UserID string
	From   string
	To     string
}

type DailyReportItem struct {
	Date  string  `json:"date"`
	Total float64 `json:"total"`
}

type MonthlyReportItem struct {
	Month string  `json:"month"`
	Total float64 `json:"total"`
}

type CategoryReportItem struct {
	CategoryID   *string `json:"category_id"`
	CategoryName *string `json:"category_name"`
	Total        float64 `json:"total"`
}
