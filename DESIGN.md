# go-project-template — Design Document

> **Purpose**: A reproducible Go project base for backends requiring web APIs, TUI clients,
> OAuth authentication, agentic OS workflows, and a React frontend. Rename via `make setup REPO=<name>`.

---

## 1. Goals & Constraints

| Goal | Constraint |
|---|---|
| Reproducible base for new projects | `make setup REPO=<name>` must be idempotent |
| No framework lock-in | Business logic only uses `http.Handler`; no Gin/Echo context leaking |
| Production-ready auth | OAuth2 (Google, GitHub, Slack, LinkedIn) + JWT + server-side sessions |
| Swappable DB | GORM behind a clean `Store` interface; raw SQL escape hatch always available |
| OS-capable agent | Shell, FS, HTTP, Playwright tools exposed to LLM |
| Single deployable artifact | Go binary embeds the built frontend (`embed.FS`) |
| Testable at every layer | Ginkgo+Gomega unit, testcontainers integration, Playwright E2E |

---

## 2. Directory Layout

```
go-project-template/
├── cmd/
│   ├── server/                  # HTTP server entrypoint
│   │   └── main.go
│   ├── tui/                     # Bubble Tea TUI client
│   │   └── main.go
│   ├── agent/                   # Agent CLI (existing, extended)
│   │   └── main.go
│   └── go-project-template/     # Template placeholder entrypoint
│       └── main.go
│
├── internal/
│   ├── config/
│   │   ├── config.go            # Viper loader (.env + config.yaml)
│   │   └── config_test.go
│   │
│   ├── server/
│   │   ├── server.go            # http.Server bootstrap
│   │   ├── routes.go            # Route registration (Go 1.22 mux)
│   │   ├── embed.go             # Embed frontend dist/
│   │   ├── middleware/
│   │   │   ├── chain.go         # Middleware composition (no framework)
│   │   │   ├── auth.go          # JWT + session validation
│   │   │   ├── cors.go
│   │   │   ├── logging.go       # Structured request logging
│   │   │   └── admin.go         # Admin-only gate + impersonation
│   │   └── handlers/
│   │       ├── health.go        # GET /healthz
│   │       ├── auth.go          # OAuth redirect + callback
│   │       ├── users.go         # CRUD /api/v1/users
│   │       ├── admin.go         # /api/v1/admin/* (assume role, user list)
│   │       ├── payments.go      # Stripe payment intent + webhook
│   │       └── agent.go         # Agent prompt + SSE streaming
│   │
│   ├── auth/
│   │   ├── oauth.go             # Provider registry + state management
│   │   ├── providers/
│   │   │   ├── google.go
│   │   │   ├── github.go
│   │   │   ├── slack.go
│   │   │   └── linkedin.go
│   │   ├── jwt.go               # Sign / verify JWT (HS256, configurable secret)
│   │   └── session.go           # DB-backed session CRUD
│   │
│   ├── db/
│   │   ├── db.go                # GORM connection factory
│   │   ├── store.go             # Store interface (repository + raw SQL)
│   │   ├── migrate.go           # golang-migrate runner
│   │   ├── models/
│   │   │   ├── user.go          # UUID PK, provider, role, paid_at
│   │   │   ├── session.go       # Token, user_id, expires_at
│   │   │   ├── payment.go       # Stripe payment record
│   │   │   └── audit.go         # Optional: impersonation audit log
│   │   └── repositories/
│   │       ├── user_repo.go
│   │       ├── session_repo.go
│   │       └── payment_repo.go
│   │
│   ├── agent/
│   │   ├── agent.go             # Extended agent wiring
│   │   └── tools/
│   │       ├── shell.go         # exec.Command with timeout + output
│   │       ├── filesystem.go    # os.ReadFile / WriteFile / Stat / Glob
│   │       ├── httptool.go      # http.Get / POST with redirect follow
│   │       └── playwright.go    # playwright-go browser automation
│   │
│   └── stripe/                  # (existing, unchanged)
│       ├── stripe.go
│       └── stripe_test.go
│
├── pkg/
│   ├── agent.go                 # (existing multi-LLM agent)
│   └── agent_test.go
│
├── frontend/                    # TypeScript + React 18 + Vite + Tailwind v3
│   ├── src/
│   │   ├── main.tsx
│   │   ├── App.tsx
│   │   ├── api/
│   │   │   └── client.ts        # Typed fetch wrapper (all API calls)
│   │   ├── contexts/
│   │   │   └── AuthContext.tsx  # JWT/session state + impersonation
│   │   ├── hooks/
│   │   │   ├── useAuth.ts
│   │   │   ├── useApi.ts
│   │   │   └── useStream.ts     # SSE consumer for agent chat
│   │   ├── pages/
│   │   │   ├── Login.tsx        # OAuth provider buttons
│   │   │   ├── Dashboard.tsx    # Authenticated home
│   │   │   ├── Admin.tsx        # User table + paid status + assume role
│   │   │   ├── Payment.tsx      # Stripe Elements payment demo
│   │   │   └── AgentChat.tsx    # Streaming agent chat
│   │   └── components/
│   │       ├── Layout.tsx
│   │       ├── NavBar.tsx
│   │       ├── ProtectedRoute.tsx
│   │       ├── AdminRoute.tsx
│   │       ├── UserTable.tsx
│   │       ├── ChatWindow.tsx
│   │       └── PaymentForm.tsx
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   └── tailwind.config.ts
│
├── migrations/                  # golang-migrate SQL files
│   ├── 000001_create_users.up.sql
│   ├── 000001_create_users.down.sql
│   ├── 000002_create_sessions.up.sql
│   ├── 000002_create_sessions.down.sql
│   ├── 000003_create_payments.up.sql
│   └── 000003_create_payments.down.sql
│
├── test/
│   ├── integration/             # testcontainers Postgres tests
│   │   └── db_test.go
│   └── e2e/                     # Playwright specs (TS or Go)
│
├── scripts/
│   ├── entrypoint.sh            # (existing)
│   └── rename-repo.sh           # (existing — replaced by make setup)
│
├── .env.example                 # All env vars, no values
├── config.yaml                  # App config (non-secret)
├── Dockerfile                   # 3-stage: frontend → go → release
├── go.mod
├── go.sum
├── Makefile
└── DESIGN.md                    # This file
```

---

## 3. HTTP Layer

### Go 1.22+ ServeMux (no framework)

```go
mux := http.NewServeMux()

// Health
mux.HandleFunc("GET /healthz", handlers.Health)

// Auth
mux.HandleFunc("GET /auth/{provider}",          handlers.OAuthRedirect)
mux.HandleFunc("GET /auth/{provider}/callback", handlers.OAuthCallback)
mux.HandleFunc("POST /auth/logout",             handlers.Logout)

// API — authenticated
mux.Handle("GET /api/v1/me",           chain(handlers.GetMe, mw.Auth))
mux.Handle("GET /api/v1/users",        chain(handlers.ListUsers, mw.Auth, mw.Admin))
mux.Handle("POST /api/v1/admin/assume/{id}", chain(handlers.AssumeUser, mw.Auth, mw.Admin))

// Payments
mux.Handle("POST /api/v1/payments/intent",  chain(handlers.CreateIntent, mw.Auth))
mux.Handle("POST /api/v1/webhooks/stripe",  handlers.StripeWebhook)

// Agent
mux.Handle("POST /api/v1/agent/prompt", chain(handlers.AgentPrompt, mw.Auth))
// SSE streaming
mux.Handle("GET /api/v1/agent/stream",  chain(handlers.AgentStream, mw.Auth))

// Frontend (catch-all, serve embedded SPA)
mux.Handle("/", handlers.Frontend)
```

### Middleware Composition

```go
// internal/server/middleware/chain.go
type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, mws ...Middleware) http.Handler {
    for i := len(mws) - 1; i >= 0; i-- {
        h = mws[i](h)
    }
    return h
}
```

No framework context — request-scoped values use standard `context.WithValue`.

---

## 4. Authentication

### Dual Strategy

| Client | Mechanism | Storage |
|---|---|---|
| Browser (SPA) | Signed cookie → session token | Postgres `sessions` table |
| API client (TUI, curl) | `Authorization: Bearer <jwt>` | Stateless (verified locally) |

Both are validated by the same `mw.Auth` middleware — it checks the `Authorization` header first,
then falls back to the `session` cookie. The resolved `*models.User` is placed on the context.

### OAuth Flow

```
1. GET /auth/google
   → generate state (CSRF), store in cookie, redirect to Google

2. GET /auth/google/callback?code=...&state=...
   → validate state, exchange code for token
   → fetch user profile from provider
   → upsert user in DB (email as unique key, provider + provider_id)
   → create session record OR sign JWT
   → redirect to /dashboard
```

### Admin Assume-Role

```
POST /api/v1/admin/assume/{userID}
  → validates caller is admin
  → writes audit log entry (admin_id, target_user_id, timestamp)
  → returns short-lived JWT with claims:
    { sub: targetUserID, admin_sub: adminID, role: "user", impersonated: true }
  → frontend swaps token; shows "Acting as <user>" banner
```

---

## 5. Database

### Store Interface

```go
// internal/db/store.go
type Store interface {
    Users()    UserRepository
    Sessions() SessionRepository
    Payments() PaymentRepository

    // Raw SQL escape hatch — always available
    Exec(ctx context.Context, sql string, args ...any) error
    QueryInto(ctx context.Context, dest any, sql string, args ...any) error
    Transaction(ctx context.Context, fn func(Store) error) error

    Close() error
}
```

### Repository Pattern

```go
type UserRepository interface {
    Create(ctx context.Context, u *models.User) error
    GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
    GetByEmail(ctx context.Context, email string) (*models.User, error)
    GetByProvider(ctx context.Context, provider, providerID string) (*models.User, error)
    List(ctx context.Context, opts ListOptions) ([]*models.User, int64, error)
    Update(ctx context.Context, u *models.User) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

### Models

```go
type User struct {
    ID         uuid.UUID  `gorm:"type:uuid;primaryKey"`
    Email      string     `gorm:"uniqueIndex;not null"`
    Name       string
    AvatarURL  string
    Provider   string     `gorm:"not null"`  // google | github | slack | linkedin
    ProviderID string     `gorm:"not null"`
    Role       string     `gorm:"default:'user'"` // user | admin
    PaidAt     *time.Time
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

type Session struct {
    ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
    UserID    uuid.UUID  `gorm:"type:uuid;not null;index"`
    Token     string     `gorm:"uniqueIndex;not null"`
    ExpiresAt time.Time  `gorm:"not null"`
    CreatedAt time.Time
}

type Payment struct {
    ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
    UserID          uuid.UUID `gorm:"type:uuid;not null;index"`
    StripePaymentID string    `gorm:"uniqueIndex"`
    Amount          int64
    Currency        string    `gorm:"default:'usd'"`
    Status          string    // pending | succeeded | failed
    CreatedAt       time.Time
}
```

### Configuration (DB host)

```yaml
# config.yaml
database:
  host: "127.0.0.1"     # set to prod DNS/IP via env override
  port: 5432
  name: "app_db"
  user: "app_user"
  sslmode: "disable"    # set to "require" in prod
```

Env override: `DATABASE_HOST=prod.db.internal` overrides `config.yaml`.

---

## 6. Configuration

Loaded in priority order (highest wins):
1. Environment variables
2. `.env` file (via godotenv, loaded before viper)
3. `config.yaml`
4. Defaults in code

```go
// internal/config/config.go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Auth     AuthConfig
    Stripe   StripeConfig
    Agent    AgentConfig
}
```

---

## 7. Agent (OS-Capable)

### Tool Interface

```go
// internal/agent/tools/tool.go
type Tool interface {
    Name()        string
    Description() string
    Parameters()  map[string]any   // JSON Schema for LLM
    Execute(ctx context.Context, input map[string]any) (string, error)
}
```

### Tools Implemented

| Tool | Capability | Safety Guard |
|---|---|---|
| `shell` | `exec.Command` with timeout | Configurable allowlist / denylist |
| `file_read` | `os.ReadFile` + `filepath.Glob` | Path must be within configured root |
| `file_write` | `os.WriteFile` | Path must be within configured root |
| `http_fetch` | GET/POST with redirect follow | URL allowlist optional |
| `playwright` | Browser automation via playwright-go | Headless, sandboxed |

---

## 8. Frontend

### Tech Stack

| Layer | Choice |
|---|---|
| Framework | React 18 |
| Build | Vite 5 |
| Styling | Tailwind CSS v3 |
| Routing | React Router v6 |
| HTTP | `fetch` (typed wrapper in `src/api/client.ts`) |
| Payments | `@stripe/react-stripe-js` |
| State | React Context (auth) + local state (simple, no Redux) |
| Streaming | `EventSource` (SSE) for agent chat |

### Pages

| Route | Page | Auth required |
|---|---|---|
| `/` | Login | No |
| `/dashboard` | Dashboard | Yes |
| `/admin` | Admin (user list, assume role) | Admin only |
| `/payment` | Stripe payment demo | Yes |
| `/agent` | Agent chat (streaming) | Yes |

### Production Embedding

```go
// internal/server/embed.go
//go:embed all:frontend/dist
var frontendFS embed.FS

func FrontendHandler() http.Handler {
    sub, _ := fs.Sub(frontendFS, "frontend/dist")
    fsHandler := http.FileServer(http.FS(sub))
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // SPA fallback: serve index.html for unknown routes
        if _, err := sub.Open(r.URL.Path); err != nil {
            http.ServeFileFS(w, r, sub, "index.html")
            return
        }
        fsHandler.ServeHTTP(w, r)
    })
}
```

---

## 9. TUI

### Stack: Bubble Tea (Charm)

```
cmd/tui/main.go
  └── model (bubbletea.Model)
      ├── views/
      │   ├── login.go    (select OAuth provider → opens browser)
      │   ├── dashboard.go (shows current user + menu)
      │   ├── users.go    (table of users — admin only)
      │   └── agent.go    (chat window with streaming)
      └── api/
          └── client.go   (thin wrapper around backend API)
```

Auth flow in TUI: user selects provider → TUI starts local HTTP callback listener
on a free port → opens browser to `/auth/{provider}?redirect_uri=127.0.0.1:{port}` →
backend redirects to local listener with JWT → TUI stores JWT for subsequent API calls.

---

## 10. Dockerfile (Single, Multi-Stage)

```
Stage 1  node:20-alpine        Build frontend → /app/frontend/dist
Stage 2  golang:1.22-alpine    Build Go binary (includes dist via embed)
Stage 3  alpine:latest         Final: binary + config.yaml + CA certs
```

The Go binary serves everything. No separate nginx needed.

---

## 11. Testing Strategy

| Layer | Framework | What |
|---|---|---|
| Unit | Ginkgo + Gomega | All internal packages; mocks via `testify/mock` |
| Integration | Ginkgo + testcontainers-go | Postgres store, auth flows |
| E2E | Playwright (TS specs) | Login → dashboard → payment → admin → agent |
| CI | GitHub Actions | Run all three tiers on PR |

---

## 12. Makefile Targets

```makefile
make setup REPO=<name>    # Idempotent: copy .env.example, rename placeholders, install tools
make build                # Build all cmd/* binaries
make run                  # Run server (requires .env)
make test                 # All tests (unit + integration)
make test-unit            # Unit tests only
make test-integration     # Integration tests (starts testcontainer)
make test-e2e             # Playwright E2E (requires running server)
make migrate-up           # Apply all pending migrations
make migrate-down         # Rollback last migration
make migrate-create NAME= # Create new migration pair
make lint                 # golangci-lint
make docker-build         # Build Docker image
make docker-push          # Push to GHCR
make clean                # Remove build artifacts
```

---

## 13. Implementation Order

1. **Config** (`internal/config`) — foundation everything else depends on
2. **Database** (`internal/db`) — models, store interface, GORM impl, migrations
3. **Auth** (`internal/auth`) — OAuth providers, JWT, session store
4. **HTTP Server** (`internal/server`) — routes, middleware, handlers (stubs)
5. **Agent** (`internal/agent`) — OS tools, extended agent
6. **Frontend** (`frontend/`) — React + Vite + Tailwind; all pages
7. **TUI** (`cmd/tui`) — Bubble Tea client
8. **Dockerfile** — update multi-stage to include frontend build
9. **Makefile** — `make setup` + all new targets
10. **Tests** — Ginkgo suites for each package, Playwright E2E
