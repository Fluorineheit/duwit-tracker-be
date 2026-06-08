package subscriptions

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSubscriptionNotFound = errors.New("subscription not found")

type SubscriptionRepository struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepository(db *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{
		db: db,
	}
}

func (r *SubscriptionRepository) GetUserIDByEmail(ctx context.Context, email string) (string, error) {
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

func (r *SubscriptionRepository) FindAll(ctx context.Context, userID string) ([]Subscription, error) {
	query := fmt.Sprintf(`
		select %s
		from subscriptions s
		left join categories c on c.id = s.category_id
		where s.user_id = $1::uuid
		order by s.is_active desc, s.billing_day asc, s.name asc, s.id asc
	`, subscriptionSelectColumns("s"))

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Subscription, 0)

	for rows.Next() {
		subscription, err := scanSubscription(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, *subscription)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *SubscriptionRepository) FindByID(ctx context.Context, userID string, id string) (*Subscription, error) {
	query := fmt.Sprintf(`
		select %s
		from subscriptions s
		left join categories c on c.id = s.category_id
		where s.id = $1::uuid
		and s.user_id = $2::uuid
	`, subscriptionSelectColumns("s"))

	subscription, err := scanSubscription(r.db.QueryRow(ctx, query, id, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSubscriptionNotFound
		}

		return nil, err
	}

	return subscription, nil
}

func (r *SubscriptionRepository) Create(ctx context.Context, input CreateSubscriptionInput) (*Subscription, error) {
	query := fmt.Sprintf(`
		with inserted as (
			insert into subscriptions (
				user_id,
				category_id,
				name,
				amount,
				currency,
				cycle,
				billing_day,
				next_billing_at
			)
			values ($1, $2, $3, $4, $5, $6, $7, $8::date)
			returning *
		)
		select %s
		from inserted s
		left join categories c on c.id = s.category_id
	`, subscriptionSelectColumns("s"))

	return scanSubscription(r.db.QueryRow(
		ctx,
		query,
		input.UserID,
		nullableStringArg(input.CategoryID),
		input.Name,
		input.Amount,
		input.Currency,
		input.Cycle,
		input.BillingDay,
		nullableStringValue(input.NextBillingAt),
	))
}

func (r *SubscriptionRepository) Update(ctx context.Context, input UpdateSubscriptionInput) (*Subscription, error) {
	setClauses := []string{}
	args := []any{}

	if input.CategoryID != nil {
		args = append(args, nullableStringArg(input.CategoryID))
		setClauses = append(setClauses, fmt.Sprintf("category_id = $%d::uuid", len(args)))
	}

	if input.Name != nil {
		args = append(args, *input.Name)
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", len(args)))
	}

	if input.Amount != nil {
		args = append(args, *input.Amount)
		setClauses = append(setClauses, fmt.Sprintf("amount = $%d", len(args)))
	}

	if input.Currency != nil {
		args = append(args, *input.Currency)
		setClauses = append(setClauses, fmt.Sprintf("currency = $%d", len(args)))
	}

	if input.Cycle != nil {
		args = append(args, *input.Cycle)
		setClauses = append(setClauses, fmt.Sprintf("cycle = $%d", len(args)))
	}

	if input.BillingDay != nil {
		args = append(args, *input.BillingDay)
		setClauses = append(setClauses, fmt.Sprintf("billing_day = $%d", len(args)))
	}

	if input.NextBillingAt != nil {
		args = append(args, *input.NextBillingAt)
		setClauses = append(setClauses, fmt.Sprintf("next_billing_at = $%d::date", len(args)))
	}

	if input.IsActive != nil {
		args = append(args, *input.IsActive)
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", len(args)))
	}

	if len(setClauses) == 0 {
		return r.FindByID(ctx, input.UserID, input.ID)
	}

	args = append(args, input.ID, input.UserID)
	idIndex := len(args) - 1
	userIDIndex := len(args)

	query := fmt.Sprintf(`
		with updated as (
			update subscriptions
			set %s
			where id = $%d::uuid
			and user_id = $%d::uuid
			returning *
		)
		select %s
		from updated s
		left join categories c on c.id = s.category_id
	`, strings.Join(setClauses, ", "), idIndex, userIDIndex, subscriptionSelectColumns("s"))

	subscription, err := scanSubscription(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSubscriptionNotFound
		}

		return nil, err
	}

	return subscription, nil
}

// Deactivate sets is_active=false. Subscriptions are never hard-deleted so the
// expenses they generated keep a valid subscription_id reference.
func (r *SubscriptionRepository) Deactivate(ctx context.Context, userID string, id string) error {
	commandTag, err := r.db.Exec(ctx, `
		update subscriptions
		set is_active = false
		where id = $1::uuid
		and user_id = $2::uuid
		and is_active = true
	`, id, userID)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrSubscriptionNotFound
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSubscription(row rowScanner) (*Subscription, error) {
	var subscription Subscription

	err := row.Scan(
		&subscription.ID,
		&subscription.UserID,
		&subscription.CategoryID,
		&subscription.CategoryName,
		&subscription.Name,
		&subscription.Amount,
		&subscription.Currency,
		&subscription.Cycle,
		&subscription.BillingDay,
		&subscription.NextBillingAt,
		&subscription.IsActive,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &subscription, nil
}

func subscriptionSelectColumns(alias string) string {
	return fmt.Sprintf(`
		%[1]s.id::text,
		%[1]s.user_id::text,
		%[1]s.category_id::text,
		c.name,
		%[1]s.name,
		%[1]s.amount::float8,
		%[1]s.currency,
		%[1]s.cycle,
		%[1]s.billing_day,
		%[1]s.next_billing_at,
		%[1]s.is_active,
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

func nullableStringValue(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	return trimmed
}
