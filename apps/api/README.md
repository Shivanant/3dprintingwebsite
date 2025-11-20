# 3DPrint Hub API

Go 1.24 service that powers authentication, pricing, carts/orders, and admin tooling for 3DPrint Hub.

---

## 📦 Requirements
- Go 1.22+ (repo tested on 1.24)
- PostgreSQL 13+ (16 recommended)
- Mailgun account (optional for dev; stdout fallback)

---

## 🔧 Setup

```bash
cp .env.example .env   # populate secrets
go mod tidy            # download modules
go run ./cmd/migrate   # apply database schema
go run ./cmd/api       # start http server on :8080
```

Common environment variables:

| Variable | Description |
|----------|-------------|
| `POSTGRES_DSN` | Connection string, e.g. `postgres://postgres:postgres@localhost:5432/print_hub?sslmode=disable` |
| `JWT_SECRET` | HS256 signing secret (32+ chars recommended) |
| `FRONTEND_URL` | Base URL of the Next.js frontend (CORS + password reset links) |
| `PUBLIC_URL` | Public URL for the API (used in OAuth redirect links) |
| `MAILGUN_DOMAIN` / `MAILGUN_API_KEY` / `MAILGUN_FROM` | Enable real password reset emails (optional) |
| `OAUTH_GOOGLE_CLIENT_ID/SECRET` | Google OAuth app credentials |
| `OAUTH_GITHUB_CLIENT_ID/SECRET` | GitHub OAuth app credentials |
| `STORAGE_UPLOADS_PATH` | Where uploaded STL/OBJ files are persisted |

Any variable omitted in dev uses the safe default defined in `internal/config`.

---

## 🧱 Project Layout

```
cmd/
  api/        # entrypoint server
  migrate/    # migrations runner (AutoMigrate)
internal/
  app/        # application container wiring services together
  auth/       # registration/login/password reset/oauth flows
  cart/       # cart CRUD
  order/      # checkout + admin status updates
  jobs/       # print job persistence
  pricing/    # STL/OBJ heuristics and cost estimation
  http/       # chi router + handlers/middleware
  database/   # GORM models and connection helpers
  token/      # JWT + refresh token utilities
  mailer/     # Mailgun + stdout fallback
  oauth/      # Google/GitHub profile fetching
```

---

## 🧪 Testing

```bash
go test ./...
```

Consider using [testcontainers-go](https://github.com/testcontainers/testcontainers-go) for future integration tests targeting Postgres.

---

## 🔐 Admin Tips

- Promote a user:
  ```sql
  UPDATE users SET role = 'admin' WHERE email = 'you@example.com';
  ```
- Password reset tokens expire in 30 minutes and are stored in `password_resets`.
- Uploaded model metadata is stored in `print_jobs` with JSONB columns for analysis metrics.

---

## 📚 API Overview

All endpoints are namespaced under `/api/v1` (see `internal/http/server.go`).

- `POST /auth/register`, `/auth/login`, `/auth/refresh`, `/auth/forgot-password`, `/auth/reset-password`
- `GET /auth/me`
- `GET /auth/oauth/:provider/start|callback`
- `POST /pricing/estimate`
- `GET/POST/DELETE /cart`, `/cart/items`
- `POST /orders/checkout`, `GET /orders`, `GET /orders/:id`
- Admin-only: `GET /admin/orders`, `PATCH /admin/orders/:id/status`

Auth middleware expects an `Authorization: Bearer <token>` header with the JWT access token.

---

## 🛣 Roadmap Ideas

- Extract migrations to SQL files for deterministic upgrades
- Add Stripe billing & webhook verification
- Queue long-running print jobs with background workers
- Structured logging + OpenTelemetry traces
