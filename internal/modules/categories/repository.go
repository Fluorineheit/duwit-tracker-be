package categories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrCategoryNotFound = errors.New("category not found")
var ErrCategoryAlreadyExists = errors.New("category already exists")

type CategoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{
		db: db,
	}
}

func (r *CategoryRepository) GetUserIDByEmail(ctx context.Context, email string) (string, error) {
	var userID string

	err := r.db.QueryRow(ctx, `
		select id::text
		from app_users
		where email = $1
		and deleted_at is null
	`, email).Scan(&userID)

	if err != nil {
		return "", err
	}

	return userID, nil
}

func (r *CategoryRepository) FindAll(ctx context.Context, userID string, categoryType string) ([]Category, error) {
	args := []any{userID}
	whereClauses := []string{
		"user_id = $1::uuid",
		"deleted_at is null",
	}

	if strings.TrimSpace(categoryType) != "" {
		args = append(args, strings.TrimSpace(categoryType))
		whereClauses = append(whereClauses, fmt.Sprintf("type = $%d", len(args)))
	}

	query := fmt.Sprintf(`
		select %s
		from categories
		where %s
		order by is_default desc, name asc
	`, categorySelectColumns(), strings.Join(whereClauses, " and "))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Category, 0)

	for rows.Next() {
		category, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, *category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *CategoryRepository) FindByID(ctx context.Context, userID string, id string) (*Category, error) {
	query := fmt.Sprintf(`
		select %s
		from categories
		where id = $1::uuid
		and user_id = $2::uuid
		and deleted_at is null
	`, categorySelectColumns())

	category, err := scanCategory(r.db.QueryRow(ctx, query, id, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}

		return nil, err
	}

	return category, nil
}

func (r *CategoryRepository) Create(ctx context.Context, input CreateCategoryInput) (*Category, error) {
	query := fmt.Sprintf(`
		with existing as (
			select id, deleted_at
			from categories
			where user_id = $1::uuid
			and lower(name) = lower($2)
			limit 1
		),
		restored as (
			update categories
			set
				name = $2,
				icon = $3,
				color = $4,
				type = $5,
				is_default = false,
				deleted_at = null
			where id = (select id from existing where deleted_at is not null)
			returning *
		),
		inserted as (
			insert into categories (
				user_id,
				name,
				icon,
				color,
				type,
				is_default
			)
			select
				$1::uuid,
				$2,
				$3,
				$4,
				$5,
				false
			where not exists (select 1 from existing)
			returning *
		)
		select %s from restored
		union all
		select %s from inserted
	`, categorySelectColumns(), categorySelectColumns())

	category, err := scanCategory(r.db.QueryRow(
		ctx,
		query,
		input.UserID,
		input.Name,
		nullableStringArg(input.Icon),
		nullableStringArg(input.Color),
		input.Type,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCategoryAlreadyExists
		}

		return nil, err
	}

	return category, nil
}

func (r *CategoryRepository) ExistsByName(ctx context.Context, userID string, name string, excludeID string) (bool, error) {
	args := []any{userID, name}
	whereClauses := []string{
		"user_id = $1::uuid",
		"lower(name) = lower($2)",
		"deleted_at is null",
	}

	if strings.TrimSpace(excludeID) != "" {
		args = append(args, excludeID)
		whereClauses = append(whereClauses, fmt.Sprintf("id != $%d::uuid", len(args)))
	}

	query := fmt.Sprintf(`
		select exists (
			select 1
			from categories
			where %s
		)
	`, strings.Join(whereClauses, " and "))

	var exists bool
	if err := r.db.QueryRow(ctx, query, args...).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *CategoryRepository) Update(ctx context.Context, input UpdateCategoryInput) (*Category, error) {
	setClauses := []string{}
	args := []any{}

	if input.Name != nil {
		args = append(args, strings.TrimSpace(*input.Name))
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", len(args)))
	}

	if input.Icon != nil {
		args = append(args, nullableStringArg(input.Icon))
		setClauses = append(setClauses, fmt.Sprintf("icon = $%d", len(args)))
	}

	if input.Color != nil {
		args = append(args, nullableStringArg(input.Color))
		setClauses = append(setClauses, fmt.Sprintf("color = $%d", len(args)))
	}

	if input.Type != nil {
		args = append(args, strings.TrimSpace(*input.Type))
		setClauses = append(setClauses, fmt.Sprintf("type = $%d", len(args)))
	}

	if len(setClauses) == 0 {
		return r.FindByID(ctx, input.UserID, input.ID)
	}

	args = append(args, input.ID, input.UserID)
	idIndex := len(args) - 1
	userIDIndex := len(args)

	query := fmt.Sprintf(`
		update categories
		set %s
		where id = $%d::uuid
		and user_id = $%d::uuid
		and deleted_at is null
		returning %s
	`, strings.Join(setClauses, ", "), idIndex, userIDIndex, categorySelectColumns())

	category, err := scanCategory(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}

		return nil, err
	}

	return category, nil
}

func (r *CategoryRepository) SoftDelete(ctx context.Context, userID string, id string) error {
	commandTag, err := r.db.Exec(ctx, `
		update categories
		set deleted_at = now()
		where id = $1::uuid
		and user_id = $2::uuid
		and deleted_at is null
	`, id, userID)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanCategory(row rowScanner) (*Category, error) {
	var category Category

	err := row.Scan(
		&category.ID,
		&category.UserID,
		&category.Name,
		&category.Icon,
		&category.Color,
		&category.Type,
		&category.IsDefault,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &category, nil
}

func categorySelectColumns() string {
	return `
		id::text,
		user_id::text,
		name,
		icon,
		color,
		type,
		is_default,
		created_at,
		updated_at
	`
}

func nullableStringArg(value *string) any {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return trimmed
}

func sqlNullString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}

	return sql.NullString{
		String: *value,
		Valid:  true,
	}
}
