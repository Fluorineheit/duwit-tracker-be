package expenses

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Fluorineheit/duwit-tracker-be/internal/config"
	"github.com/google/uuid"
)

type ExpenseService struct {
	repo *ExpenseRepository
	cfg  config.Config
}

func NewExpenseService(repo *ExpenseRepository, cfg config.Config) *ExpenseService {
	return &ExpenseService{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *ExpenseService) Create(ctx context.Context, req CreateExpenseRequest) (*ExpenseResponse, error) {
	userID, err := s.getCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	categoryID, err := normalizeOptionalUUID(req.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("invalid category_id")
	}

	currency := normalizeCurrency(req.Currency)
	source, err := normalizeSource(req.Source)
	if err != nil {
		return nil, err
	}

	spentAt, err := parseSpentAt(req.SpentAt)
	if err != nil {
		return nil, err
	}

	expense, err := s.repo.Create(ctx, CreateExpenseInput{
		UserID:     userID,
		CategoryID: categoryID,
		Amount:     req.Amount,
		Currency:   currency,
		Note:       stringPtr(req.Note),
		RawText:    stringPtr(req.RawText),
		Source:     source,
		SpentAt:    spentAt,
	})
	if err != nil {
		return nil, err
	}

	response := expense.ToResponse()
	return &response, nil
}

func (s *ExpenseService) FindAll(ctx context.Context, query ListExpensesQuery) (*ListExpensesResult, error) {
	userID, err := s.getCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	var cursor *string
	if trimmed := strings.TrimSpace(query.Cursor); trimmed != "" {
		cursor = &trimmed
	}

	categoryID, err := normalizeOptionalUUID(query.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("invalid category_id")
	}

	from, err := parseDateFilter(query.From, false)
	if err != nil {
		return nil, fmt.Errorf("invalid from date")
	}

	to, err := parseDateFilter(query.To, true)
	if err != nil {
		return nil, fmt.Errorf("invalid to date")
	}

	items, nextCursor, hasMore, err := s.repo.FindAll(ctx, ListExpensesParams{
		UserID:     userID,
		Limit:      limit,
		Cursor:     cursor,
		CategoryID: categoryID,
		From:       from,
		To:         to,
		Search:     strings.TrimSpace(query.Search),
	})
	if err != nil {
		return nil, err
	}

	responses := make([]ExpenseResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, item.ToResponse())
	}

	return &ListExpensesResult{
		Items:      responses,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Limit:      limit,
	}, nil
}

func (s *ExpenseService) FindByID(ctx context.Context, id string) (*ExpenseResponse, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid expense id")
	}

	userID, err := s.getCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	expense, err := s.repo.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	response := expense.ToResponse()
	return &response, nil
}

func (s *ExpenseService) Update(ctx context.Context, id string, req UpdateExpenseRequest) (*ExpenseResponse, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid expense id")
	}

	userID, err := s.getCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	var categoryID *string
	if req.CategoryID != nil {
		parsedCategoryID, err := normalizeOptionalUUID(*req.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("invalid category_id")
		}

		if parsedCategoryID == nil {
			empty := ""
			categoryID = &empty
		} else {
			categoryID = parsedCategoryID
		}
	}

	if req.Amount != nil && *req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	var currency *string
	if req.Currency != nil {
		normalized := normalizeCurrency(*req.Currency)
		currency = &normalized
	}

	var source *string
	if req.Source != nil {
		normalizedSource, err := normalizeSource(*req.Source)
		if err != nil {
			return nil, err
		}

		source = &normalizedSource
	}

	var spentAt *string
	if req.SpentAt != nil {
		parsedSpentAt, err := parseSpentAt(*req.SpentAt)
		if err != nil {
			return nil, err
		}

		spentAt = &parsedSpentAt
	}

	expense, err := s.repo.Update(ctx, UpdateExpenseInput{
		ID:         id,
		UserID:     userID,
		CategoryID: categoryID,
		Amount:     req.Amount,
		Currency:   currency,
		Note:       req.Note,
		RawText:    req.RawText,
		Source:     source,
		SpentAt:    spentAt,
	})
	if err != nil {
		return nil, err
	}

	response := expense.ToResponse()
	return &response, nil
}

func (s *ExpenseService) SoftDelete(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid expense id")
	}

	userID, err := s.getCurrentUserID(ctx)
	if err != nil {
		return err
	}

	return s.repo.SoftDelete(ctx, userID, id)
}

func (s *ExpenseService) getCurrentUserID(ctx context.Context) (string, error) {
	if strings.TrimSpace(s.cfg.AppUserEmail) == "" {
		return "", fmt.Errorf("APP_USER_EMAIL is required")
	}

	return s.repo.GetUserIDByEmail(ctx, s.cfg.AppUserEmail)
}

func normalizeOptionalUUID(value string) (*string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}

	if _, err := uuid.Parse(trimmed); err != nil {
		return nil, err
	}

	return &trimmed, nil
}

func normalizeCurrency(value string) string {
	trimmed := strings.ToUpper(strings.TrimSpace(value))
	if trimmed == "" {
		return "IDR"
	}

	return trimmed
}

func normalizeSource(value string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		return "manual", nil
	}

	allowedSources := map[string]bool{
		"manual":   true,
		"telegram": true,
		"import":   true,
		"ai":       true,
	}

	if !allowedSources[trimmed] {
		return "", fmt.Errorf("invalid source")
	}

	return trimmed, nil
}

func parseSpentAt(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Now().Format(time.RFC3339), nil
	}

	if parsed, err := time.Parse("2006-01-02", trimmed); err == nil {
		return parsed.Format(time.RFC3339), nil
	}

	if parsed, err := time.Parse(time.RFC3339, trimmed); err == nil {
		return parsed.Format(time.RFC3339), nil
	}

	return "", fmt.Errorf("spent_at must use YYYY-MM-DD or RFC3339 format")
}

func parseDateFilter(value string, inclusiveEnd bool) (*string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}

	if parsed, err := time.Parse("2006-01-02", trimmed); err == nil {
		if inclusiveEnd {
			parsed = parsed.AddDate(0, 0, 1)
		}

		formatted := parsed.Format(time.RFC3339)
		return &formatted, nil
	}

	if parsed, err := time.Parse(time.RFC3339, trimmed); err == nil {
		formatted := parsed.Format(time.RFC3339)
		return &formatted, nil
	}

	return nil, errors.New("invalid date format")
}
