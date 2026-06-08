package subscriptions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Fluorineheit/duwit-tracker-be/internal/config"
	"github.com/google/uuid"
)

type SubscriptionService struct {
	repo *SubscriptionRepository
	cfg  config.Config
}

func NewSubscriptionService(repo *SubscriptionRepository, cfg config.Config) *SubscriptionService {
	return &SubscriptionService{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *SubscriptionService) FindAll(ctx context.Context) (*ListSubscriptionsResult, error) {
	userID, err := s.getCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.FindAll(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]SubscriptionResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, item.ToResponse())
	}

	return &ListSubscriptionsResult{
		Items: responses,
	}, nil
}

func (s *SubscriptionService) Create(ctx context.Context, req CreateSubscriptionRequest) (*SubscriptionResponse, error) {
	userID, err := s.getCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	categoryID, err := normalizeOptionalUUID(req.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("invalid category_id")
	}

	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	cycle, err := normalizeCycle(req.Cycle)
	if err != nil {
		return nil, err
	}

	if req.BillingDay < 1 || req.BillingDay > 31 {
		return nil, fmt.Errorf("billing_day must be between 1 and 31")
	}

	subscription, err := s.repo.Create(ctx, CreateSubscriptionInput{
		UserID:        userID,
		CategoryID:    categoryID,
		Name:          name,
		Amount:        req.Amount,
		Currency:      normalizeCurrency(req.Currency),
		Cycle:         cycle,
		BillingDay:    req.BillingDay,
		NextBillingAt: computeNextBillingAt(req.BillingDay, time.Now()),
	})
	if err != nil {
		return nil, err
	}

	response := subscription.ToResponse()
	return &response, nil
}

func (s *SubscriptionService) Update(ctx context.Context, id string, req UpdateSubscriptionRequest) (*SubscriptionResponse, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid subscription id")
	}

	userID, err := s.getCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	input := UpdateSubscriptionInput{
		ID:       id,
		UserID:   userID,
		IsActive: req.IsActive,
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, fmt.Errorf("name cannot be empty")
		}
		input.Name = &name
	}

	if req.CategoryID != nil {
		categoryID, err := normalizeOptionalUUID(*req.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("invalid category_id")
		}

		if categoryID == nil {
			empty := ""
			input.CategoryID = &empty
		} else {
			input.CategoryID = categoryID
		}
	}

	if req.Amount != nil {
		if *req.Amount <= 0 {
			return nil, fmt.Errorf("amount must be greater than 0")
		}
		input.Amount = req.Amount
	}

	if req.Currency != nil {
		normalized := normalizeCurrency(*req.Currency)
		input.Currency = &normalized
	}

	if req.Cycle != nil {
		cycle, err := normalizeCycle(*req.Cycle)
		if err != nil {
			return nil, err
		}
		input.Cycle = &cycle
	}

	if req.BillingDay != nil {
		if *req.BillingDay < 1 || *req.BillingDay > 31 {
			return nil, fmt.Errorf("billing_day must be between 1 and 31")
		}
		input.BillingDay = req.BillingDay

		// Re-anchor the next billing date when the billing day changes.
		nextBillingAt := computeNextBillingAt(*req.BillingDay, time.Now())
		input.NextBillingAt = &nextBillingAt
	}

	subscription, err := s.repo.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	response := subscription.ToResponse()
	return &response, nil
}

func (s *SubscriptionService) Deactivate(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid subscription id")
	}

	userID, err := s.getCurrentUserID(ctx)
	if err != nil {
		return err
	}

	return s.repo.Deactivate(ctx, userID, id)
}

func (s *SubscriptionService) getCurrentUserID(ctx context.Context) (string, error) {
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

func normalizeCycle(value string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		return "monthly", nil
	}

	if trimmed != "monthly" && trimmed != "yearly" {
		return "", fmt.Errorf("cycle must be monthly or yearly")
	}

	return trimmed, nil
}

// computeNextBillingAt returns the next date on or after `from` whose day of
// month matches billingDay, clamped to the last day for short months.
func computeNextBillingAt(billingDay int, from time.Time) string {
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)

	candidate := clampDay(from.Year(), from.Month(), billingDay)
	if candidate.Before(from) {
		next := from.AddDate(0, 1, 0)
		candidate = clampDay(next.Year(), next.Month(), billingDay)
	}

	return candidate.Format("2006-01-02")
}

func clampDay(year int, month time.Month, day int) time.Time {
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if day > lastDay {
		day = lastDay
	}

	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
