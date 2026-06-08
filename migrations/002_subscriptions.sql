-- CatatDuit subscriptions schema + expenses linkage

-- =========================================================
-- 1. Subscriptions
-- =========================================================
create table if not exists subscriptions (
    id uuid primary key default gen_random_uuid(),

    user_id uuid not null references app_users(id) on delete cascade,
    category_id uuid references categories(id) on delete set null,

    name text not null,

    amount numeric(14, 2) not null check (amount > 0),
    currency text not null default 'IDR',

    -- How often the subscription bills.
    cycle text not null default 'monthly'
        check (cycle in ('monthly', 'yearly')),

    -- Day of month the subscription bills, e.g. 15.
    billing_day int not null check (billing_day between 1 and 31),

    -- Next scheduled billing date, advanced by pg_cron after each charge.
    next_billing_at date,

    is_active boolean not null default true,

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

-- =========================================================
-- 2. Link expenses to the subscription that generated them
-- =========================================================
alter table expenses
    add column if not exists subscription_id uuid
        references subscriptions(id) on delete set null;

-- Allow pg_cron auto-billing to record 'subscription' as an expense source.
alter table expenses drop constraint if exists expenses_source_check;
alter table expenses
    add constraint expenses_source_check
    check (source in ('manual', 'telegram', 'import', 'ai', 'subscription'));

-- =========================================================
-- 3. updated_at trigger
-- =========================================================
drop trigger if exists set_subscriptions_updated_at on subscriptions;
create trigger set_subscriptions_updated_at
before update on subscriptions
for each row
execute function set_updated_at();

-- =========================================================
-- 4. Indexes
-- =========================================================
create index if not exists idx_subscriptions_user_id
on subscriptions(user_id)
where is_active = true;

create index if not exists idx_subscriptions_billing
on subscriptions(billing_day)
where is_active = true;

create index if not exists idx_expenses_subscription_id
on expenses(subscription_id)
where deleted_at is null;
