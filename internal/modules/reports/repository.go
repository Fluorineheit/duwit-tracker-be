package reports

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ReportRepository struct {
	db *pgxpool.Pool
}

func NewReportRepository(db *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{
		db: db,
	}
}

func (r *ReportRepository) GetUserIDByEmail(ctx context.Context, email string) (string, error) {
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

func (r *ReportRepository) Daily(ctx context.Context, params DailyReportParams) ([]DailyReportItem, error) {
	rows, err := r.db.Query(ctx, `
		select to_char(spent_at, 'YYYY-MM-DD') as date, sum(amount)::float8 as total
		from expenses
		where user_id = $1::uuid
		and deleted_at is null
		and spent_at >= $2::timestamptz
		and spent_at < $3::timestamptz
		group by 1
		order by 1
	`, params.UserID, params.From, params.To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]DailyReportItem, 0)

	for rows.Next() {
		var item DailyReportItem
		if err := rows.Scan(&item.Date, &item.Total); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *ReportRepository) Monthly(ctx context.Context, params MonthlyReportParams) ([]MonthlyReportItem, error) {
	rows, err := r.db.Query(ctx, `
		select to_char(spent_at, 'YYYY-MM') as month, sum(amount)::float8 as total
		from expenses
		where user_id = $1::uuid
		and deleted_at is null
		and spent_at >= $2::timestamptz
		and spent_at < $3::timestamptz
		group by 1
		order by 1
	`, params.UserID, params.From, params.To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]MonthlyReportItem, 0)

	for rows.Next() {
		var item MonthlyReportItem
		if err := rows.Scan(&item.Month, &item.Total); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *ReportRepository) Category(ctx context.Context, params CategoryReportParams) ([]CategoryReportItem, error) {
	rows, err := r.db.Query(ctx, `
		select e.category_id::text, c.name, sum(e.amount)::float8 as total
		from expenses e
		left join categories c on c.id = e.category_id
		where e.user_id = $1::uuid
		and e.deleted_at is null
		and e.spent_at >= $2::timestamptz
		and e.spent_at < $3::timestamptz
		group by e.category_id, c.name
		order by total desc
	`, params.UserID, params.From, params.To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]CategoryReportItem, 0)

	for rows.Next() {
		var (
			categoryID   sql.NullString
			categoryName sql.NullString
			item         CategoryReportItem
		)

		if err := rows.Scan(&categoryID, &categoryName, &item.Total); err != nil {
			return nil, err
		}

		item.CategoryID = nullStringToPtr(categoryID)
		item.CategoryName = nullStringToPtr(categoryName)

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func nullStringToPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	return &value.String
}
