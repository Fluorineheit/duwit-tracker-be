package budgets

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Fluorineheit/duwit-tracker-be/internal/config"
	"github.com/google/uuid"
)

type BudgetService struct {
	repo *BudgetRepository
	cfg  config.Config
}

func NewBudgetService(repo *BudgetRepository, cfg config.Config) *BudgetService {
	return &BudgetService{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *BudgetService) FindAll(ctx context.Context, query ListBudgetsQuery) (*ListBudgetsResult, error) {
	userID, err := s.getCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	var month *string
	if strings.TrimSpace(query.Month) != "" {
		normalized, err := normalizeMonth(query.Month)
		if err != nil {
			return nil, err
		}

		month = &normalized
	}

	items, err := s.repo.FindAll(ctx, ListBudgetsParams{
		UserID: userID,
		Month:  month,
	})
	if err != nil {
		return nil, err
	}

	responses := make([]BudgetResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, item.ToResponse())
	}

	return &ListBudgetsResult{
		Items: responses,
	}, nil
}

func (s *BudgetService) Create(ctx context.Context, req CreateBudgetRequest) (*BudgetResponse, error) {
	userID, err := s.getCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := uuid.Parse(strings.TrimSpace(req.CategoryID)); err != nil {
		return nil, fmt.Errorf("invalid category id")
	}

	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}

	month, err := normalizeMonth(req.Month)
	if err != nil {
		return nil, err
	}

	budget, err := s.repo.Create(ctx, CreateBudgetInput{
		UserID:     userID,
		CategoryID: strings.TrimSpace(req.CategoryID),
		Amount:     req.Amount,
		Currency:   normalizeCurrency(req.Currency),
		Month:      month,
	})
	if err != nil {
		return nil, err
	}

	response := budget.ToResponse()
	return &response, nil
}

func (s *BudgetService) SoftDelete(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid budget id")
	}

	userID, err := s.getCurrentUserID(ctx)
	if err != nil {
		return err
	}

	return s.repo.SoftDelete(ctx, userID, id)
}

func (s *BudgetService) getCurrentUserID(ctx context.Context) (string, error) {
	if strings.TrimSpace(s.cfg.AppUserEmail) == "" {
		return "", fmt.Errorf("APP_USER_EMAIL is required")
	}

	return s.repo.GetUserIDByEmail(ctx, s.cfg.AppUserEmail)
}

// normalizeMonth accepts YYYY-MM-DD or YYYY-MM and returns the first day of that
// month as YYYY-MM-DD, since budgets are stored per calendar month.
func normalizeMonth(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("month is required")
	}

	if parsed, err := time.Parse("2006-01-02", trimmed); err == nil {
		return firstOfMonth(parsed), nil
	}

	if parsed, err := time.Parse("2006-01", trimmed); err == nil {
		return firstOfMonth(parsed), nil
	}

	return "", fmt.Errorf("month must use YYYY-MM-DD or YYYY-MM format")
}

func firstOfMonth(value time.Time) string {
	return time.Date(value.Year(), value.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
}

func normalizeCurrency(value string) string {
	trimmed := strings.ToUpper(strings.TrimSpace(value))
	if trimmed == "" {
		return "IDR"
	}

	return trimmed
}
