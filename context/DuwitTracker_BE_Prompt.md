# DuwitTracker — Backend Codex Prompt
> Go Gin • Supabase PostgreSQL • OpenClaw • Repository-Service-Handler

---

## 1. Project Context

You are working on DuwitTracker, a personal expense tracker. The backend is built with Go Gin and Supabase PostgreSQL. It handles REST API, business logic, validation, security, and OpenClaw integration for Telegram input.

---

## 2. Tech Stack

- Go + Gin
- pgxpool — PostgreSQL connection pooling
- godotenv — local environment config
- go-playground/validator — input validation
- Supabase PostgreSQL
- Pattern: Repository-Service-Handler
- OpenClaw — Telegram agent (replaces custom bot)

---

## 3. Directory Structure

```
duwittracker-be/
├── cmd/api/main.go
├── internal/
│   ├── config/
│   ├── database/
│   ├── server/
│   ├── modules/
│   │   ├── expenses/
│   │   ├── categories/
│   │   ├── budgets/
│   │   ├── subscriptions/
│   │   ├── reports/
│   │   └── ai/
│   └── pkg/
│       ├── response/
│       └── logger/
├── migrations/
├── go.mod
└── go.sum
```

---

## 4. Database Schema

### app_users
| Field | Type | Notes |
|---|---|---|
| id | UUID | primary key |
| email | text | |
| telegram_chat_id | text | for OpenClaw identification |
| created_at | timestamptz | |
| updated_at | timestamptz | |

### categories
| Field | Type | Notes |
|---|---|---|
| id | UUID | primary key |
| user_id | UUID | foreign key → app_users |
| name | text | Food, Coffee, Transport, etc. |
| type | text | expense or income |
| color | text | hex color |
| icon | text | icon name |
| created_at | timestamptz | |
| updated_at | timestamptz | |
| deleted_at | timestamptz | soft delete |

### expenses
| Field | Type | Notes |
|---|---|---|
| id | UUID | primary key |
| user_id | UUID | foreign key → app_users |
| category_id | UUID | foreign key → categories |
| subscription_id | UUID | foreign key → subscriptions, nullable |
| amount | numeric(14,2) | never float for money |
| note | text | description |
| spent_at | timestamptz | when expense occurred |
| source | text | manual, openclaw, import, subscription |
| raw_text | text | original OpenClaw input |
| ai_confidence | numeric(3,2) | future AI review score |
| status | text | confirmed, draft |
| created_at | timestamptz | |
| updated_at | timestamptz | |
| deleted_at | timestamptz | soft delete |

### subscriptions
| Field | Type | Notes |
|---|---|---|
| id | UUID | primary key |
| user_id | UUID | foreign key → app_users |
| category_id | UUID | foreign key → categories |
| name | text | e.g. Spotify, Netflix |
| amount | numeric(14,2) | billing amount |
| cycle | text | monthly, yearly |
| billing_day | int | day of month e.g. 15 |
| next_billing_at | date | next scheduled billing date |
| is_active | boolean | default true |
| created_at | timestamptz | |
| updated_at | timestamptz | |

### budgets
| Field | Type | Notes |
|---|---|---|
| id | UUID | primary key |
| user_id | UUID | foreign key → app_users |
| category_id | UUID | foreign key → categories |
| month | date | first day of month e.g. 2026-06-01 |
| limit_amount | numeric(14,2) | monthly limit |
| created_at | timestamptz | |
| updated_at | timestamptz | |

### Design rules
- UUID primary keys on all tables
- user_id on every table — every query scoped by user_id
- numeric(14,2) for all money — never float
- deleted_at soft delete — never hard delete financial records
- raw_text stores original OpenClaw input for audit trail

---

## 5. API Endpoints

### Pagination strategy — cursor (not offset)
- Cursor = `spent_at` timestamp of last item client received
- Query: `WHERE spent_at < $cursor AND user_id = $user_id ORDER BY spent_at DESC LIMIT $limit`
- Default limit: 20, max: 100
- When date_from/date_to filter applied — cursor optional, result naturally bounded
- When response returns fewer items than limit → client reached the end

**List response shape:**
```json
{
  "data": [...],
  "meta": {
    "limit": 20,
    "next_cursor": "2026-05-01T10:00:00Z",
    "has_more": true
  }
}
```

### Expenses
```
GET    /api/v1/expenses         cursor, limit, date_from, date_to, category_id, source, search
POST   /api/v1/expenses
GET    /api/v1/expenses/:id
PUT    /api/v1/expenses/:id
DELETE /api/v1/expenses/:id     soft delete
```

POST body:
```json
{
  "category_id": "uuid",
  "amount": 18000,
  "note": "kopi",
  "spent_at": "2026-06-03T08:00:00Z",
  "source": "manual",
  "raw_text": "original openclaw input if applicable"
}
```

### Categories
```
GET    /api/v1/categories        type=expense|income
POST   /api/v1/categories
GET    /api/v1/categories/:id
PUT    /api/v1/categories/:id
DELETE /api/v1/categories/:id    soft delete
```

### Budgets
```
GET    /api/v1/budgets           month=2026-06-01
POST   /api/v1/budgets
DELETE /api/v1/budgets/:id
```

### Reports
```
GET    /api/v1/reports/daily     date_from, date_to → [{ date, total }]
GET    /api/v1/reports/monthly   year → [{ month, total }]
GET    /api/v1/reports/category  date_from, date_to → [{ category_id, category_name, total }]
```

### Subscriptions
```
GET    /api/v1/subscriptions         — list all subscriptions
POST   /api/v1/subscriptions         — create subscription
PUT    /api/v1/subscriptions/:id     — update subscription
DELETE /api/v1/subscriptions/:id     — deactivate (set is_active=false)
```

POST body:
```json
{
  "name": "Spotify",
  "category_id": "uuid",
  "amount": 49000,
  "cycle": "monthly",
  "billing_day": 15
}
```

### Advisor (future feature)
```
GET    /api/v1/advisor/context   category_id, amount
```
Returns pre-calculated context for LLM advice:
```json
{
  "this_week": 84000,
  "weekly_average": 45000,
  "monthly_budget": 200000,
  "monthly_spent": 168000,
  "anomaly": true,
  "anomaly_reason": "3x above weekly average"
}
```

---

## 6. Response Format

**Success:**
```json
{ "data": { ... } }
```

**Error:**
```json
{ "error": { "code": "VALIDATION_ERROR", "message": "amount is required" } }
```

---

## 7. Security

- Cloudflare Access — restricts dashboard to your email (MVP)
- Every DB query scoped by user_id
- OPENCLAW_API_KEY — OpenClaw skill sends as Bearer token, backend validates
- All secrets in environment variables only — never in frontend

### Environment variables
```
APP_NAME=DuwitTracker API
APP_ENV=development
APP_PORT=8080
APP_USER_EMAIL=your_email
DATABASE_URL=your_supabase_database_url
DB_MAX_CONNS=5
DB_MIN_CONNS=1
OPENCLAW_API_KEY=your_openclaw_secret
```

---

## 8. OpenClaw Integration

OpenClaw replaces the custom Telegram bot, rule-based parser, and Gemini AI fallback.

**Flow:**
```
User: "kopi 18k sama makan siang 25000"
        ↓
OpenClaw (Claude/GPT handles NLP + Indonesian)
        ↓
POST /api/v1/expenses { note: "kopi", amount: 18000, source: "openclaw" }
POST /api/v1/expenses { note: "makan siang", amount: 25000, source: "openclaw" }
        ↓
OpenClaw replies: "Saved! Kopi Rp18.000 + Makan Siang Rp25.000 ✓"
```

**Endpoints OpenClaw uses:**
- `POST /api/v1/expenses`
- `GET  /api/v1/categories`
- `GET  /api/v1/reports/monthly`
- `GET  /api/v1/expenses`
- `GET  /api/v1/advisor/context` (future)

---

## 9. Subscriptions — pg_cron Auto-billing

Subscriptions auto-insert into expenses via Supabase `pg_cron`. No cron job needed in Go — the database handles it directly, so it works even when Render free tier is sleeping.

**Enable in Supabase:**
Dashboard → Database → Extensions → search `pg_cron` → enable

**pg_cron function:**
```sql
CREATE OR REPLACE FUNCTION process_due_subscriptions()
RETURNS void AS $$
BEGIN
  -- Insert due subscriptions as expenses
  INSERT INTO expenses (id, user_id, category_id, subscription_id, amount, note, spent_at, source)
  SELECT
    gen_random_uuid(),
    user_id,
    category_id,
    id,
    amount,
    name || ' - ' || to_char(now(), 'Month YYYY'),
    now(),
    'subscription'
  FROM subscriptions
  WHERE billing_day = EXTRACT(DAY FROM now())
  AND is_active = true;

  -- Update next billing date
  UPDATE subscriptions
  SET next_billing_at = next_billing_at + interval '1 month'
  WHERE billing_day = EXTRACT(DAY FROM now())
  AND is_active = true;
END;
$$ LANGUAGE plpgsql;

-- Schedule: runs every day at midnight
SELECT cron.schedule(
  'process-subscriptions',
  '0 0 * * *',
  'SELECT process_due_subscriptions()'
);
```

**Flow:**
```
pg_cron fires at midnight
        ↓
process_due_subscriptions()
        ↓
INSERT into expenses (source: 'subscription', subscription_id: ...)
UPDATE subscriptions next_billing_at
        ↓
User opens dashboard next morning
        ↓
Spotify Rp49.000 already appears automatically
```

---

## 10. Development Order

1. Setup Go Gin project structure
2. Connect to Supabase PostgreSQL via pgxpool
3. Create database migrations (app_users, categories, expenses, budgets, subscriptions)
4. Build Expenses CRUD API with cursor pagination + filters
5. Build Categories API
6. Build Subscriptions CRUD API
7. Enable pg_cron in Supabase + add process_due_subscriptions function
8. Build Reports API (daily, monthly, category)
9. Build Budgets API
10. Deploy to Render + configure Cloudflare Access + backend security
11. Add OPENCLAW_API_KEY auth middleware
12. Test all endpoints reachable from OpenClaw skill
13. Add Advisor context endpoint (future)
14. Polish and harden API for portfolio

---

## 11. Codex Prompt Template

```
You are working on my DuwitTracker expense tracker project.

Backend stack:
  Go + Gin | pgxpool | Supabase PostgreSQL | godotenv
  Pattern: Repository-Service-Handler

Current status:
  [describe what is already done]

Next goal:
  [describe the one step to implement]

Rules:
  - Implement ONE step only
  - Use standard error envelope: { error: { code, message } }
  - All list endpoints use cursor-based pagination — never offset
  - Cursor = spent_at timestamp, query WHERE spent_at < $cursor
  - Every query must be scoped by user_id
  - Use go-playground/validator for input validation
  - source field on expenses: manual | openclaw | import | subscription
  - Subscriptions auto-insert into expenses via Supabase pg_cron — not Go cron
  - Do not build a Telegram webhook — OpenClaw handles that
  - Do not add Gemini or AI parsing — OpenClaw handles that
  - Do not add auth unless the step is specifically about auth
  - Make the smallest clean changes
  - Explain all files changed after implementation
```
