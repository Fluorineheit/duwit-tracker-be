package budgets

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrBudgetNotFound = errors.New("budget not found")
var ErrBudgetAlreadyExists = errors.New("budget already exists")

type BudgetRepository struct {
	db *pgxpool.Pool
}

func NewBudgetRepository(db *pgxpool.Pool) *BudgetRepository {
	return &BudgetRepository{
		db: db,
	}
}

func (r *BudgetRepository) GetUserIDByEmail(ctx context.Context, email string) (string, error) {
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

func (r *BudgetRepository) FindAll(ctx context.Context, params ListBudgetsParams) ([]Budget, error) {
	args := []any{params.UserID}
	whereClauses := []string{
		"b.user_id = $1::uuid",
		"b.deleted_at is null",
	}

	if params.Month != nil {
		args = append(args, *params.Month)
		whereClauses = append(whereClauses, fmt.Sprintf("b.month = $%d::date", len(args)))
	}

	query := fmt.Sprintf(`
		select %s
		from budgets b
		left join categories c on c.id = b.category_id
		where %s
		order by b.month desc, c.name asc, b.id asc
	`, budgetSelectColumns("b"), strings.Join(whereClauses, " and "))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Budget, 0)

	for rows.Next() {
		budget, err := scanBudget(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, *budget)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// Create inserts a new budget, restoring a soft-deleted one for the same
// (user_id, category_id, month) if it exists. An active duplicate yields
// ErrBudgetAlreadyExists.
func (r *BudgetRepository) Create(ctx context.Context, input CreateBudgetInput) (*Budget, error) {
	query := fmt.Sprintf(`
		with existing as (
			select id, deleted_at
			from budgets
			where user_id = $1::uuid
			and category_id = $2::uuid
			and month = $3::date
			limit 1
		),
		restored as (
			update budgets
			set
				amount = $4,
				currency = $5,
				deleted_at = null
			where id = (select id from existing where deleted_at is not null)
			returning *
		),
		inserted as (
			insert into budgets (
				user_id,
				category_id,
				amount,
				currency,
				month
			)
			select
				$1::uuid,
				$2::uuid,
				$4,
				$5,
				$3::date
			where not exists (select 1 from existing)
			returning *
		),
		combined as (
			select * from restored
			union all
			select * from inserted
		)
		select %s
		from combined b
		left join categories c on c.id = b.category_id
	`, budgetSelectColumns("b"))

	budget, err := scanBudget(r.db.QueryRow(
		ctx,
		query,
		input.UserID,
		input.CategoryID,
		input.Month,
		input.Amount,
		input.Currency,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBudgetAlreadyExists
		}

		return nil, err
	}

	return budget, nil
}

func (r *BudgetRepository) SoftDelete(ctx context.Context, userID string, id string) error {
	commandTag, err := r.db.Exec(ctx, `
		update budgets
		set deleted_at = now()
		where id = $1::uuid
		and user_id = $2::uuid
		and deleted_at is null
	`, id, userID)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrBudgetNotFound
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanBudget(row rowScanner) (*Budget, error) {
	var budget Budget

	err := row.Scan(
		&budget.ID,
		&budget.UserID,
		&budget.CategoryID,
		&budget.CategoryName,
		&budget.Amount,
		&budget.Currency,
		&budget.Month,
		&budget.CreatedAt,
		&budget.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &budget, nil
}

func budgetSelectColumns(alias string) string {
	return fmt.Sprintf(`
		%[1]s.id::text,
		%[1]s.user_id::text,
		%[1]s.category_id::text,
		c.name,
		%[1]s.amount::float8,
		%[1]s.currency,
		%[1]s.month,
		%[1]s.created_at,
		%[1]s.updated_at
	`, alias)
}
