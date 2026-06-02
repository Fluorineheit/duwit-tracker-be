-- CatatDuit initial database schema

-- Use UUIDs as primary keys.
-- gen_random_uuid() is available in modern PostgreSQL/Supabase projects.
create extension if not exists pgcrypto;

-- =========================================================
-- 1. App users
-- =========================================================
create table if not exists app_users (
    id uuid primary key default gen_random_uuid(),

    email text unique not null,
    display_name text,

    -- For future Supabase Auth integration.
    -- We keep it nullable because MVP uses Cloudflare Access first.
    auth_user_id uuid unique,

    -- Telegram chat ID is used to allowlist your personal Telegram account.
    telegram_chat_id bigint unique,

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

-- =========================================================
-- 2. Categories
-- =========================================================
create table if not exists categories (
    id uuid primary key default gen_random_uuid(),

    user_id uuid not null references app_users(id) on delete cascade,

    name text not null,
    icon text,
    color text,

    -- expense now, income later if you want.
    type text not null default 'expense'
        check (type in ('expense', 'income')),

    is_default boolean not null default false,

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz,

    unique (user_id, name)
);

-- =========================================================
-- 3. Expenses
-- =========================================================
create table if not exists expenses (
    id uuid primary key default gen_random_uuid(),

    user_id uuid not null references app_users(id) on delete cascade,
    category_id uuid references categories(id) on delete set null,

    amount numeric(14, 2) not null check (amount > 0),
    currency text not null default 'IDR',

    note text,
    raw_text text,

    -- manual, telegram, import, ai, etc.
    source text not null default 'manual'
        check (source in ('manual', 'telegram', 'import', 'ai')),

    spent_at timestamptz not null default now(),

    -- For later AI parsing/categorization confidence.
    ai_confidence numeric(5, 4)
        check (ai_confidence is null or (ai_confidence >= 0 and ai_confidence <= 1)),

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

-- =========================================================
-- 4. Budgets
-- =========================================================
create table if not exists budgets (
    id uuid primary key default gen_random_uuid(),

    user_id uuid not null references app_users(id) on delete cascade,
    category_id uuid references categories(id) on delete cascade,

    amount numeric(14, 2) not null check (amount > 0),
    currency text not null default 'IDR',

    -- Store the month as the first day of the month.
    -- Example: 2026-06-01 means June 2026 budget.
    month date not null,

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz,

    unique (user_id, category_id, month)
);

-- =========================================================
-- 5. updated_at trigger
-- =========================================================
create or replace function set_updated_at()
returns trigger as $$
begin
    new.updated_at = now();
    return new;
end;
$$ language plpgsql;

drop trigger if exists set_app_users_updated_at on app_users;
create trigger set_app_users_updated_at
before update on app_users
for each row
execute function set_updated_at();

drop trigger if exists set_categories_updated_at on categories;
create trigger set_categories_updated_at
before update on categories
for each row
execute function set_updated_at();

drop trigger if exists set_expenses_updated_at on expenses;
create trigger set_expenses_updated_at
before update on expenses
for each row
execute function set_updated_at();

drop trigger if exists set_budgets_updated_at on budgets;
create trigger set_budgets_updated_at
before update on budgets
for each row
execute function set_updated_at();

-- =========================================================
-- 6. Indexes
-- =========================================================
create index if not exists idx_categories_user_id
on categories(user_id)
where deleted_at is null;

create index if not exists idx_expenses_user_id_spent_at
on expenses(user_id, spent_at desc)
where deleted_at is null;

create index if not exists idx_expenses_category_id
on expenses(category_id)
where deleted_at is null;

create index if not exists idx_expenses_source
on expenses(source)
where deleted_at is null;

create index if not exists idx_budgets_user_id_month
on budgets(user_id, month)
where deleted_at is null;

-- =========================================================
-- 7. Seed your first app user
-- =========================================================
insert into app_users (email, display_name)
values ('farhanadika7@gmail.com', 'Farhan')
on conflict (email) do nothing;

-- =========================================================
-- 8. Seed default categories
-- =========================================================
insert into categories (user_id, name, icon, color, is_default)
select
    u.id,
    c.name,
    c.icon,
    c.color,
    true
from app_users u
cross join (
    values
        ('Food', 'utensils', '#ef4444'),
        ('Coffee', 'coffee', '#92400e'),
        ('Transport', 'car', '#3b82f6'),
        ('Shopping', 'shopping-bag', '#a855f7'),
        ('Bills', 'receipt', '#f97316'),
        ('Entertainment', 'gamepad-2', '#22c55e'),
        ('Health', 'heart-pulse', '#ec4899'),
        ('Gym', 'dumbbell', '#14b8a6'),
        ('Other', 'circle', '#6b7280')
) as c(name, icon, color)
where u.email = 'farhanadika7@gmail.com'
on conflict (user_id, name) do nothing;