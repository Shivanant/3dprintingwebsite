# 3DPrint Hub – Architecture Overview

## High-level layout
- `apps/web` – existing Next.js 15 frontend (currently `/src`).
- `apps/api` – new Go 1.22 service exposing REST + OAuth callbacks.
- Shared services via Postgres 16, Mailgun, and cloud object storage (stubbed locally).

```
┌────────────┐        JWT/REST         ┌────────────┐        SQL         ┌──────────────┐
│ Next.js UI │  <──────────────────►  │  Go API    │  <───────────────►  │   Postgres   │
└────────────┘                        └────────────┘                    └──────────────┘
        ▲                                   │
        │                                   ▼
        └──────── OAuth redirect ◄──────── Google/GitHub
                        │
                        ▼
                     Mailgun
```

## Backend service (`apps/api`)

### Tech stack
- Go 1.22
- Router: `github.com/go-chi/chi/v5`
- Database: `gorm.io/gorm` with `gorm.io/driver/postgres`
- Auth: signed JWT (HS256) via `github.com/golang-jwt/jwt/v5`
- Password hashing: `golang.org/x/crypto/bcrypt`
- OAuth2: `golang.org/x/oauth2` + provider configs
- Email: `github.com/mailgun/mailgun-go/v4`
- File analysis: `github.com/hschendel/stl` (STL), lightweight OBJ parser (custom)
- Storage abstraction: local disk for dev (`storage/uploads`)

### Environment
```
POSTGRES_DSN=postgres://user:pass@host:5432/print?sslmode=disable
JWT_SECRET=...
MAILGUN_DOMAIN=example.com
MAILGUN_API_KEY=key-xxx
MAILGUN_FROM=3DPrint Hub <noreply@example.com>
OAUTH_GOOGLE_CLIENT_ID=...
OAUTH_GOOGLE_CLIENT_SECRET=...
OAUTH_GITHUB_CLIENT_ID=...
OAUTH_GITHUB_CLIENT_SECRET=...
PUBLIC_URL=http://localhost:8080
FRONTEND_URL=http://localhost:3000
```

### Database schema (initial)
- `users`
  - `id UUID PK`
  - `email CITEXT UNIQUE`
  - `password_hash TEXT NULL`
  - `name TEXT`
  - `avatar_url TEXT`
  - `role TEXT default 'user'`
  - `created_at`, `updated_at`
- `oauth_accounts`
  - `id UUID PK`
  - `user_id UUID FK -> users`
  - `provider TEXT`
  - `provider_user_id TEXT`
  - `access_token TEXT`
  - `refresh_token TEXT`
  - `expires_at TIMESTAMP NULL`
  - `created_at`, `updated_at`
- `password_resets`
  - `id UUID PK`
  - `user_id UUID`
  - `token TEXT UNIQUE`
  - `expires_at TIMESTAMP`
  - `used_at TIMESTAMP NULL`
- `carts`
  - `id UUID PK`
  - `user_id UUID UNIQUE`
- `cart_items`
  - `id UUID`
  - `cart_id UUID`
  - `sku TEXT`
  - `display_name TEXT`
  - `quantity INT`
  - `unit_price_cents INT`
  - `metadata JSONB`
  - `created_at`, `updated_at`
- `orders`
  - `id UUID`
  - `user_id UUID`
  - `status TEXT` (`pending`, `paid`, `in_progress`, `shipped`, `cancelled`)
  - `subtotal_cents INT`
  - `tax_cents INT`
  - `total_cents INT`
  - `currency TEXT`
  - `notes TEXT`
  - `created_at`, `updated_at`
- `order_items`
  - `id UUID`
  - `order_id UUID`
  - `name TEXT`
  - `description TEXT`
  - `quantity INT`
  - `unit_price_cents INT`
  - `metadata JSONB`
  - `created_at`, `updated_at`
- `print_jobs`
  - `id UUID`
  - `user_id UUID`
  - `order_item_id UUID NULL`
  - `file_name TEXT`
  - `storage_path TEXT`
  - `material TEXT`
  - `quality TEXT`
  - `estimated_grams DECIMAL`
  - `estimated_hours DECIMAL`
  - `estimated_price_cents INT`
  - `analysis JSONB`
  - `status TEXT`
  - `created_at`, `updated_at`

### Core endpoints (prefixed `/api/v1`)

**Auth**
- `POST /auth/register` – email/password, optional name
- `POST /auth/login`
- `POST /auth/refresh`
- `POST /auth/forgot-password`
- `POST /auth/reset-password`
- `GET /auth/me` – session introspection
- `GET /auth/oauth/:provider/start` – returns redirect URL
- `GET /auth/oauth/:provider/callback` – handles provider callback

**Profile**
- `PUT /users/me` – update profile / shipping info (future)

**Cart & Orders**
- `GET /cart` – user cart
- `POST /cart/items` – add/update
- `DELETE /cart/items/:id`
- `POST /cart/checkout` – create order
- `GET /orders` – list user orders
- `GET /orders/:id` – detail

**Admin**
- `GET /admin/orders` – list paginated
- `PATCH /admin/orders/:id/status`
- `GET /admin/users`

**Pricing & files**
- `POST /pricing/estimate` – multipart upload, returns metrics + price
- `POST /print-jobs` – create job (persist file metadata, optional order)
- `GET /print-jobs/:id` – details (secure)

### Security
- JWT access token (15m), refresh token (14d). Refresh tokens persisted with rotation.
- Middleware verifying `Authorization: Bearer`.
- `RoleMiddleware` gating admin routes.
- CSRF not needed (pure API + Authorization header).

### File handling & pricing heuristic
1. Save upload to temp.
2. Detect extension (`.stl`, `.obj`, `.3mf` fallback).
3. For STL binary, compute triangle mesh volume + bounding box; `density = 1.24 g/cm³`.
4. Est. grams = volume(mm³)/1000 * density.
5. Est. hours = (volume / (print_speed_mm3_per_hr)) + setup factor; calibrate with configurable constants.
6. Pricing = material_cost_per_gram*grams + machine_rate_per_hour*hours + setup_fee.
7. Return classification (material, recommended infill) for UI.

### Testing
- Unit tests for auth usecases, pricing estimator, OBJ parser.
- Integration tests using `testcontainers-go` for Postgres (future).

## Frontend updates
- `src/lib/apiClient.ts` – axios instance targeting `process.env.NEXT_PUBLIC_API_URL`.
- `src/store/auth.ts` – Zustand store for auth state (tokens stored via httpOnly cookie? -> tokens stored in memory + refresh).
- Authentication pages: `/login`, `/register`, `/forgot-password`, `/reset-password`.
- OAuth buttons invoking backend start endpoint (window.location = ...).
- Protected layout for `/account`, `/cart`, `/admin`.
- Admin panel: orders table with status controls, job metrics.
- Lithophane page redesign with canvas preview, gradient background, stepper UI.
- Cart gating: if unauthenticated, show interstitial prompting login.
- Upload flow now hits Go API for estimation and job creation.

## Docker (dev)
- Single `Dockerfile.dev` for Go API with live reload via `air`.
- `docker-compose.dev.yml` bringing up
  - `api` (Go)
  - `db` (postgres:16-alpine)
  - `mailhog` (for local email)
  - `web` (Next.js npm run dev)
- Named volumes for Go cache, node_modules.

## TODO rollout steps
1. Scaffold Go service (`apps/api`) with config, migrations, and health check.
2. Wire database migrations + `go run cmd/migrate/main.go`.
3. Implement auth flows + email.
4. Implement carts/orders.
5. Integrate pricing estimator + storage.
6. Update frontend hooking API.
7. Polish Lithophane and admin.
8. Provide documentation + env example + docker compose.

