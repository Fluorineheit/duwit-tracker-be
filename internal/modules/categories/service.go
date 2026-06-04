package categories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Fluorineheit/duwit-tracker-be/internal/config"
	"github.com/google/uuid"
)

type CategoryService struct {
	repo *CategoryRepository
	cfg  config.Config
}

func NewCategoryService(repo *CategoryRepository, cfg config.Config) *CategoryService {
	return &CategoryService{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *CategoryService) FindAll(ctx context.Context, query ListCategoriesQuery) (*ListCategoriesResult, error) {
	userID, err := s.getCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	categoryType, err := normalizeCategoryType(query.Type)
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

	items, nextCursor, hasMore, err := s.repo.FindAll(ctx, ListCategoriesParams{
		UserID: userID,
		Type:   categoryType,
		Limit:  limit,
		Cursor: cursor,
	})
	if err != nil {
		return nil, err
	}

	responses := make([]CategoryResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, item.ToResponse())
	}

	return &ListCategoriesResult{
		Items:      responses,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Limit:      limit,
	}, nil
}

func (s *CategoryService) FindByID(ctx context.Context, id string) (*CategoryResponse, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid category id")
	}

	userID, err := s.getCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	category, err := s.repo.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	response := category.ToResponse()
	return &response, nil
}

func (s *CategoryService) Create(ctx context.Context, req CreateCategoryRequest) (*CategoryResponse, error) {
	userID, err := s.getCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	name, err := normalizeCategoryName(req.Name)
	if err != nil {
		return nil, err
	}

	categoryType, err := normalizeCategoryType(req.Type)
	if err != nil {
		return nil, err
	}

	category, err := s.repo.Create(ctx, CreateCategoryInput{
		UserID: userID,
		Name:   name,
		Icon:   stringPtr(req.Icon),
		Color:  stringPtr(req.Color),
		Type:   categoryType,
	})
	if err != nil {
		return nil, err
	}

	response := category.ToResponse()
	return &response, nil
}

func (s *CategoryService) Update(ctx context.Context, id string, req UpdateCategoryRequest) (*CategoryResponse, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("invalid category id")
	}

	userID, err := s.getCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	var name *string
	if req.Name != nil {
		normalizedName, err := normalizeCategoryName(*req.Name)
		if err != nil {
			return nil, err
		}

		exists, err := s.repo.ExistsByName(ctx, userID, normalizedName, id)
		if err != nil {
			return nil, err
		}

		if exists {
			return nil, ErrCategoryAlreadyExists
		}

		name = &normalizedName
	}

	var categoryType *string
	if req.Type != nil {
		normalizedType, err := normalizeCategoryType(*req.Type)
		if err != nil {
			return nil, err
		}

		categoryType = &normalizedType
	}

	category, err := s.repo.Update(ctx, UpdateCategoryInput{
		ID:     id,
		UserID: userID,
		Name:   name,
		Icon:   req.Icon,
		Color:  req.Color,
		Type:   categoryType,
	})
	if err != nil {
		return nil, err
	}

	response := category.ToResponse()
	return &response, nil
}

func (s *CategoryService) SoftDelete(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid category id")
	}

	userID, err := s.getCurrentUserID(ctx)
	if err != nil {
		return err
	}

	return s.repo.SoftDelete(ctx, userID, id)
}

func (s *CategoryService) getCurrentUserID(ctx context.Context) (string, error) {
	if strings.TrimSpace(s.cfg.AppUserEmail) == "" {
		return "", fmt.Errorf("APP_USER_EMAIL is required")
	}

	return s.repo.GetUserIDByEmail(ctx, s.cfg.AppUserEmail)
}

func normalizeCategoryName(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("category name is required")
	}

	if len(trimmed) > 80 {
		return "", fmt.Errorf("category name must be 80 characters or less")
	}

	return trimmed, nil
}

func normalizeCategoryType(value string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		return "expense", nil
	}

	allowedTypes := map[string]bool{
		"expense": true,
		"income":  true,
	}

	if !allowedTypes[trimmed] {
		return "", fmt.Errorf("category type must be expense or income")
	}

	return trimmed, nil
}

func stringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func IsDuplicateCategoryError(err error) bool {
	return errors.Is(err, ErrCategoryAlreadyExists)
}
