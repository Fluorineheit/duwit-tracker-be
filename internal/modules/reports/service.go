package reports

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Fluorineheit/duwit-tracker-be/internal/config"
)

type ReportService struct {
	repo *ReportRepository
	cfg  config.Config
}

func NewReportService(repo *ReportRepository, cfg config.Config) *ReportService {
	return &ReportService{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *ReportService) Daily(ctx context.Context, query DailyReportQuery) ([]DailyReportItem, error) {
	userID, err := s.getCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	from, to, err := parseDateRange(query.DateFrom, query.DateTo)
	if err != nil {
		return nil, err
	}

	return s.repo.Daily(ctx, DailyReportParams{
		UserID: userID,
		From:   from,
		To:     to,
	})
}

func (s *ReportService) Monthly(ctx context.Context, query MonthlyReportQuery) ([]MonthlyReportItem, error) {
	userID, err := s.getCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	from, to, err := parseYear(query.Year)
	if err != nil {
		return nil, err
	}

	return s.repo.Monthly(ctx, MonthlyReportParams{
		UserID: userID,
		From:   from,
		To:     to,
	})
}

func (s *ReportService) Category(ctx context.Context, query CategoryReportQuery) ([]CategoryReportItem, error) {
	userID, err := s.getCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	from, to, err := parseDateRange(query.DateFrom, query.DateTo)
	if err != nil {
		return nil, err
	}

	return s.repo.Category(ctx, CategoryReportParams{
		UserID: userID,
		From:   from,
		To:     to,
	})
}

func (s *ReportService) getCurrentUserID(ctx context.Context) (string, error) {
	if strings.TrimSpace(s.cfg.AppUserEmail) == "" {
		return "", fmt.Errorf("APP_USER_EMAIL is required")
	}

	return s.repo.GetUserIDByEmail(ctx, s.cfg.AppUserEmail)
}

// parseDateRange validates date_from/date_to (YYYY-MM-DD, both required) and returns
// RFC3339 bounds. date_to is inclusive, so the upper bound is the start of the next day.
func parseDateRange(dateFrom string, dateTo string) (string, string, error) {
	from, err := parseDateBound(dateFrom, "date_from")
	if err != nil {
		return "", "", err
	}

	to, err := parseDateBound(dateTo, "date_to")
	if err != nil {
		return "", "", err
	}

	if to.Before(from) {
		return "", "", fmt.Errorf("date_to must be on or after date_from")
	}

	// Make date_to inclusive by extending to the start of the following day.
	to = to.AddDate(0, 0, 1)

	return from.Format(time.RFC3339), to.Format(time.RFC3339), nil
}

func parseDateBound(value string, field string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("%s is required", field)
	}

	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s must use YYYY-MM-DD format", field)
	}

	return parsed, nil
}

// parseYear validates a 4-digit year and returns the [year-01-01, (year+1)-01-01) bounds.
func parseYear(value string) (string, string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", "", fmt.Errorf("year is required")
	}

	year, err := strconv.Atoi(trimmed)
	if err != nil || year < 1000 || year > 9999 {
		return "", "", fmt.Errorf("year must be a 4-digit value")
	}

	from := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(1, 0, 0)

	return from.Format(time.RFC3339), to.Format(time.RFC3339), nil
}
