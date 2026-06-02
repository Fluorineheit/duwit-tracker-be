package expenses

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrExpenseNotFound = errors.New("expense not found")

type ExpenseRepository struct {
	db *pgxpool.Pool
}

func NewExpenseRepository(db *pgxpool.Pool) *ExpenseRepository {
	return &ExpenseRepository{
		db: db,
	}
}

func (r *ExpenseRepository) GetUserIDByEmail(ctx context.Context, email string) (string, error) {
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

func (r *ExpenseRepository) Create(ctx context.Context, input CreateExpenseInput) (*Expense, error) {
	query := fmt.Sprintf(`
		with inserted as (
			insert into expenses (
				user_id,
				category_id,
				amount,
				currency,
				note,
				raw_text,
				source,
				spent_at
			)
			values ($1, $2, $3, $4, $5, $6, $7, $8::timestamptz)
			returning *
		)
		select %s
		from inserted e
		left join categories c on c.id = e.category_id
	`, expenseSelectColumns("e"))

	return scanExpense(r.db.QueryRow(
		ctx,
		query,
		input.UserID,
		nullableStringArg(input.CategoryID),
		input.Amount,
		input.Currency,
		nullableStringArg(input.Note),
		nullableStringArg(input.RawText),
		input.Source,
		input.SpentAt,
	))
}

func (r *ExpenseRepository) FindAll(ctx context.Context, params ListExpensesParams) ([]Expense, int64, error) {
	args := []any{params.UserID}
	whereClauses := []string{
		"e.user_id = $1",
		"e.deleted_at is null",
	}

	if params.CategoryID != nil {
		args = append(args, *params.CategoryID)
		whereClauses = append(whereClauses, fmt.Sprintf("e.category_id = $%d::uuid", len(args)))
	}

	if params.From != nil {
		args = append(args, *params.From)
		whereClauses = append(whereClauses, fmt.Sprintf("e.spent_at >= $%d::timestamptz", len(args)))
	}

	if params.To != nil {
		args = append(args, *params.To)
		whereClauses = append(whereClauses, fmt.Sprintf("e.spent_at < $%d::timestamptz", len(args)))
	}

	if strings.TrimSpace(params.Search) != "" {
		args = append(args, "%"+strings.TrimSpace(params.Search)+"%")
		whereClauses = append(whereClauses, fmt.Sprintf("(e.note ilike $%d or e.raw_text ilike $%d)", len(args), len(args)))
	}

	whereSQL := strings.Join(whereClauses, " and ")

	countQuery := fmt.Sprintf(`
		select count(*)
		from expenses e
		left join categories c on c.id = e.category_id
		where %s
	`, whereSQL)

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, params.Limit, params.Offset)
	limitIndex := len(args) - 1
	offsetIndex := len(args)

	query := fmt.Sprintf(`
		select %s
		from expenses e
		left join categories c on c.id = e.category_id
		where %s
		order by e.spent_at desc, e.created_at desc
		limit $%d offset $%d
	`, expenseSelectColumns("e"), whereSQL, limitIndex, offsetIndex)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]Expense, 0)

	for rows.Next() {
		expense, err := scanExpense(rows)
		if err != nil {
			return nil, 0, err
		}

		items = append(items, *expense)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *ExpenseRepository) FindByID(ctx context.Context, userID string, id string) (*Expense, error) {
	query := fmt.Sprintf(`
		select %s
		from expenses e
		left join categories c on c.id = e.category_id
		where e.id = $1::uuid
		and e.user_id = $2::uuid
		and e.deleted_at is null
	`, expenseSelectColumns("e"))

	expense, err := scanExpense(r.db.QueryRow(ctx, query, id, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrExpenseNotFound
		}

		return nil, err
	}

	return expense, nil
}

func (r *ExpenseRepository) Update(ctx context.Context, input UpdateExpenseInput) (*Expense, error) {
	setClauses := []string{}
	args := []any{}

	if input.CategoryID != nil {
		args = append(args, nullableStringArg(input.CategoryID))
		setClauses = append(setClauses, fmt.Sprintf("category_id = $%d::uuid", len(args)))
	}

	if input.Amount != nil {
		args = append(args, *input.Amount)
		setClauses = append(setClauses, fmt.Sprintf("amount = $%d", len(args)))
	}

	if input.Currency != nil {
		args = append(args, *input.Currency)
		setClauses = append(setClauses, fmt.Sprintf("currency = $%d", len(args)))
	}

	if input.Note != nil {
		args = append(args, nullableStringArg(input.Note))
		setClauses = append(setClauses, fmt.Sprintf("note = $%d", len(args)))
	}

	if input.RawText != nil {
		args = append(args, nullableStringArg(input.RawText))
		setClauses = append(setClauses, fmt.Sprintf("raw_text = $%d", len(args)))
	}

	if input.Source != nil {
		args = append(args, *input.Source)
		setClauses = append(setClauses, fmt.Sprintf("source = $%d", len(args)))
	}

	if input.SpentAt != nil {
		args = append(args, *input.SpentAt)
		setClauses = append(setClauses, fmt.Sprintf("spent_at = $%d::timestamptz", len(args)))
	}

	if len(setClauses) == 0 {
		return r.FindByID(ctx, input.UserID, input.ID)
	}

	args = append(args, input.ID, input.UserID)
	idIndex := len(args) - 1
	userIDIndex := len(args)

	query := fmt.Sprintf(`
		with updated as (
			update expenses
			set %s
			where id = $%d::uuid
			and user_id = $%d::uuid
			and deleted_at is null
			returning *
		)
		select %s
		from updated e
		left join categories c on c.id = e.category_id
	`, strings.Join(setClauses, ", "), idIndex, userIDIndex, expenseSelectColumns("e"))

	expense, err := scanExpense(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrExpenseNotFound
		}

		return nil, err
	}

	return expense, nil
}

func (r *ExpenseRepository) SoftDelete(ctx context.Context, userID string, id string) error {
	commandTag, err := r.db.Exec(ctx, `
		update expenses
		set deleted_at = now()
		where id = $1::uuid
		and user_id = $2::uuid
		and deleted_at is null
	`, id, userID)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrExpenseNotFound
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanExpense(row rowScanner) (*Expense, error) {
	var expense Expense

	err := row.Scan(
		&expense.ID,
		&expense.UserID,
		&expense.CategoryID,
		&expense.CategoryName,
		&expense.Amount,
		&expense.Currency,
		&expense.Note,
		&expense.RawText,
		&expense.Source,
		&expense.SpentAt,
		&expense.AIConfidence,
		&expense.CreatedAt,
		&expense.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &expense, nil
}

func expenseSelectColumns(alias string) string {
	return fmt.Sprintf(`
		%[1]s.id::text,
		%[1]s.user_id::text,
		%[1]s.category_id::text,
		c.name,
		%[1]s.amount::float8,
		%[1]s.currency,
		%[1]s.note,
		%[1]s.raw_text,
		%[1]s.source,
		%[1]s.spent_at,
		%[1]s.ai_confidence::float8,
		%[1]s.created_at,
		%[1]s.updated_at
	`, alias)
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

func stringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func timeToRFC3339Ptr(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := value.Format(time.RFC3339)
	return &formatted
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
