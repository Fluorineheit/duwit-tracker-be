# Duwit Tracker API

Duwit Tracker API is a Go backend for tracking personal money records. It currently supports expense and category management, backed by PostgreSQL.

The API is built as a small modular monolith: each feature module owns its HTTP handler, service logic, data transfer objects, model mapping, and repository queries.

## Tech Stack

- Go
- Gin for HTTP routing
- PostgreSQL
- pgx for PostgreSQL access
- godotenv for local environment loading

## Project Structure

```text
cmd/api/main.go                 Application entrypoint
internal/config                 Environment configuration
internal/database               PostgreSQL connection setup
internal/response               Shared API response helpers
internal/server                 Router and dependency wiring
internal/modules/categories     Category feature module
internal/modules/expenses       Expense feature module
migrations                      Database schema migrations
```

Feature modules follow this pattern:

```text
handler.go      HTTP request/response layer
service.go      Validation and application logic
repository.go   Database queries
model.go        Internal database-facing model
dto.go          Request, response, and input structs
```

Request flow:

```text
HTTP request
  -> Gin router
  -> module handler
  -> module service
  -> module repository
  -> PostgreSQL
```

## Getting Started

### 1. Install dependencies

```powershell
go mod download
```

### 2. Configure environment

Create a `.env` file in the project root:

```env
APP_NAME=DuwitTrackerApp
APP_ENV=development
APP_PORT=8080
APP_USER_EMAIL=your-email@example.com

DATABASE_URL=postgresql://user:password@host:5432/database
DB_MAX_CONNS=5
DB_MIN_CONNS=1
```

`APP_USER_EMAIL` must match an existing row in the `app_users` table. The current MVP uses this email to resolve the active user for category and expense operations.

### 3. Run database migration

Apply the SQL in:

```text
migrations/001_init_schema.sql
```

Example with `psql`:

```powershell
psql "postgresql://user:password@host:5432/database" -f migrations/001_init_schema.sql
```

### 4. Start the API

For local development on Windows, use the dev launcher. It stops an old Go-built API process on the configured port before starting a fresh server:

```powershell
.\scripts\dev.ps1
```

You can also run the API directly:

```powershell
go run ./cmd/api
```

By default the server starts at:

```text
http://localhost:8080
```

## API Routes

All API routes are under `/api/v1`.

### Health

```http
GET /
GET /api/v1/health
GET /api/v1/health/database
```

### Categories

```http
GET    /api/v1/categories
GET    /api/v1/categories?type=expense
GET    /api/v1/categories?limit=20&cursor=<next_cursor>
GET    /api/v1/categories/:id
POST   /api/v1/categories
PUT    /api/v1/categories/:id
DELETE /api/v1/categories/:id
```

Create category body:

```json
{
  "name": "Food",
  "icon": "utensils",
  "color": "#ef4444",
  "type": "expense"
}
```

Category type defaults to `expense` and must be either `expense` or `income`.

### Expenses

```http
GET    /api/v1/expenses
GET    /api/v1/expenses?limit=20&cursor=<next_cursor>
GET    /api/v1/expenses?category_id=<uuid>&from=2026-06-01&to=2026-06-30&search=coffee
GET    /api/v1/expenses/:id
POST   /api/v1/expenses
PUT    /api/v1/expenses/:id
DELETE /api/v1/expenses/:id
```

Create expense body:

```json
{
  "category_id": "category-uuid",
  "amount": 25000,
  "currency": "IDR",
  "note": "Lunch",
  "raw_text": "lunch 25k",
  "source": "manual",
  "spent_at": "2026-06-02"
}
```

Expense notes:

- `amount` must be greater than `0`.
- `currency` defaults to `IDR`.
- `source` defaults to `manual` and must be one of `manual`, `telegram`, `import`, or `ai`.
- `spent_at` accepts `YYYY-MM-DD` or RFC3339. If omitted, it defaults to the current time.

List endpoints (`GET /categories`, `GET /expenses`) use cursor (keyset) pagination:

- `limit` defaults to `20` and is capped at `100`.
- `cursor` is an opaque token; omit it for the first page.
- The response `data` contains `items`, `has_more`, and `next_cursor` (the value to pass as `cursor` for the next page; `null` when there are no more pages).

## Response Format

Successful responses use:

```json
{
  "success": true,
  "message": "Message",
  "data": {}
}
```

Error responses use:

```json
{
  "success": false,
  "message": "Error message"
}
```

## Troubleshooting

### Port 8080 is already in use

If you see:

```text
listen tcp :8080: bind: Only one usage of each socket address is normally permitted
```

Find the process:

```powershell
netstat -ano | Select-String ':8080'
```

Then stop it:

```powershell
Stop-Process -Id <PID>
```

Or change `APP_PORT` in `.env` to another port, such as `8081`.

For local development, prefer:

```powershell
.\scripts\dev.ps1
```

The script reads `APP_PORT` from `.env`, stops an old `api.exe` created by `go run`, and then starts the API again.

### Database connection failed

Check that:

- `DATABASE_URL` is set correctly.
- The database is reachable from your machine.
- The migration has been applied.
- `APP_USER_EMAIL` exists in the `app_users` table.
