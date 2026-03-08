# retrosnack

A Progressive Web App (PWA) ecommerce store for [@retrosnack.shop](https://instagram.com/retrosnack.shop) — a women's thrift clothing store sourcing secondhand brand clothing, accessories, and shoes in good condition for less than retail. Community members can also sell through the store.

---

## Architecture

retrosnack is built as a **modular monolith** — a single Go binary with clean domain-separated packages that mirror microservice boundaries. This gives the simplicity and zero-cost of a single deployment today, with a clear extraction path to true microservices later.

```mermaid
graph TD
    User["User (Browser / Mobile)"]

    subgraph Cloudflare
        CF_CDN["Cloudflare CDN"]
        R2["Cloudflare R2\nObject Storage"]
        Pages["Cloudflare Pages\nSvelteKit PWA"]
    end

    subgraph Render
        RenderAPI["Render\nGo API"]
    end

    subgraph "Go API Modules"
        Auth["auth"]
        Catalog["catalog"]
        Inventory["inventory"]
        Orders["orders"]
        Payments["payments"]
        Instagram["instagram"]
        Media["media"]
    end

    subgraph "Data"
        Neon["Neon\nPostgreSQL"]
    end

    subgraph "External"
        Square["Square\nPayments"]
        IGEmbed["Instagram\noEmbed API"]
    end

    subgraph "CI/CD"
        GHA["GitHub Actions"]
    end

    User --> CF_CDN
    CF_CDN --> Pages
    Pages --> RenderAPI
    RenderAPI --> Auth & Catalog & Inventory & Orders & Payments & Instagram & Media
    Auth & Catalog & Inventory & Orders --> Neon
    Payments --> Square
    Instagram --> IGEmbed
    Media --> R2
    GHA --> Pages
    GHA --> RenderAPI
```

### Domain Modules

| Module | Responsibility |
|---|---|
| `auth` | JWT-based authentication |
| `catalog` | Products, categories, sizing |
| `inventory` | Stock per item/variant (most items are one-of-a-kind) |
| `orders` | Order lifecycle |
| `payments` | Square payment links, webhook fulfillment |
| `instagram` | oEmbed link management per product |
| `media` | Image upload/serve via Cloudflare R2 |

---

## Tech Stack

| Layer | Technology | Rationale | Cost |
|---|---|---|---|
| Frontend | SvelteKit + `vite-plugin-pwa` | Fast, lightweight, installable PWA | Free |
| Frontend hosting | Cloudflare Pages | Global CDN, zero egress fees | Free |
| Backend | Go + chi router | Simple, fast, single binary | Free |
| Backend hosting | Render | Free tier, auto-deploy from GitHub, managed TLS | Free |
| Database | Neon PostgreSQL | Serverless, 0.5 GB free tier | Free |
| Object storage | Cloudflare R2 | 10 GB free, no egress fees | Free |
| Payments | Square | Payment links, unified in-person + online payments | 2.9% + 30¢ |
| ORM / queries | sqlc + goose + pgx/v5 | Type-safe SQL, zero runtime overhead | Free |
| CI/CD | GitHub Actions | Lint, test, build, deploy | Free |

---

## Repository Structure

```
retrosnack/
├── apps/
│   └── frontend/                  # SvelteKit PWA
│       ├── src/
│       │   ├── lib/               # Shared utilities and components
│       │   ├── routes/            # SvelteKit file-based routing
│       │   └── app.html
│       ├── static/
│       │   ├── manifest.json      # PWA manifest
│       │   └── icons/             # App icons for installation
│       ├── package.json
│       ├── svelte.config.js
│       └── vite.config.js         # vite-plugin-pwa configuration
│
├── services/
│   └── api/                       # Go modular monolith
│       ├── cmd/
│       │   └── server/
│       │       └── main.go        # Entry point
│       ├── internal/
│       │   ├── auth/              # JWT authentication
│       │   ├── catalog/           # Products, categories, sizing
│       │   ├── inventory/         # Stock tracking
│       │   ├── orders/            # Order lifecycle
│       │   ├── payments/          # Square payment links
│       │   ├── instagram/         # oEmbed link management
│       │   └── media/             # Image upload via R2
│       ├── db/
│       │   ├── migrations/        # goose SQL migrations
│       │   └── queries/           # sqlc SQL query definitions
│       ├── pkg/
│       │   ├── config/            # Environment configuration
│       │   ├── middleware/        # HTTP middleware (auth, logging, CORS)
│       │   └── httputil/          # Shared HTTP helpers
│       ├── go.mod
│       └── go.sum
│
├── infrastructure/
│   ├── nginx/
│   │   └── nginx.conf             # Reverse proxy config
│   └── docker/
│       └── Dockerfile             # Multi-stage Go build
│
├── docs/
│   └── architecture/
│       └── decisions.md           # Architecture decision records (ADRs)
│
├── .github/
│   └── workflows/
│       ├── ci-api.yml             # Lint, test, build Go API on PRs
│       ├── ci-frontend.yml        # Type check, build frontend on PRs
│       └── deploy-frontend.yml    # Deploy frontend to Cloudflare Pages on main
│
├── docker-compose.yml             # Local development stack (Postgres, API, frontend)
├── render.yaml                    # Render service blueprint
├── sqlc.yaml                      # sqlc code generation config
└── README.md
```

---

## Development Workflow

### Prerequisites

- [Go](https://go.dev/dl/) 1.23+
- [Node.js](https://nodejs.org/) 20+ and [pnpm](https://pnpm.io/)
- [Docker](https://www.docker.com/) and Docker Compose
- [sqlc](https://sqlc.dev/) — `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- [goose](https://github.com/pressly/goose) — `go install github.com/pressly/goose/v3/cmd/goose@latest`

### Local Development

A `Makefile` is provided so you don't need to remember individual commands or pass secrets in the terminal. All env vars are loaded from `.env` automatically.

1. **Clone the repository**

   ```bash
   git clone https://github.com/retrosnack-clothing/retrosnack.git
   cd retrosnack
   ```

2. **Copy environment variables**

   ```bash
   cp .env.example .env
   ```

   The defaults in `.env.example` point at the local Docker Postgres — no changes needed to get started. Fill in Square and R2 keys if you need to test payments or image uploads.

3. **Install dependencies and start the database**

   ```bash
   make install       # install Go and frontend deps
   make db            # start local PostgreSQL via Docker
   ```

4. **Run migrations and generate types**

   ```bash
   make migrate       # apply SQL migrations
   make sqlc          # generate type-safe Go code
   ```

5. **Start the services** (two terminals)

   ```bash
   make api           # terminal 1 — Go API on localhost:8080
   make frontend      # terminal 2 — SvelteKit on localhost:5173
   ```

   Or start everything at once via Docker Compose:

   ```bash
   make dev
   ```

Run `make help` to see all available commands.

---

## Deployment Overview

### Cloudflare Pages (Frontend)

- Push to `main` triggers GitHub Actions to build and deploy the SvelteKit app to Cloudflare Pages.
- Global CDN distribution, zero egress fees, automatic HTTPS.

### Render (Backend API)

- Go binary built via multi-stage Docker build in `infrastructure/docker/Dockerfile`.
- Render handles HTTPS termination natively — no nginx sidecar needed.
- Service blueprint: `render.yaml`.
- Secrets managed via Render dashboard environment variables.
- Free tier: 750 hours/month, auto-deploy from GitHub on push to `main`.
- Sleeps after 15 min idle; kept warm via UptimeRobot ping every 5 min.

### Neon PostgreSQL

- Serverless PostgreSQL, connects from Render via `DATABASE_URL`.
- Free tier: 0.5 GB storage, auto-suspend when idle.

### Cloudflare R2

- Product images uploaded via Go `media` module using the S3-compatible API.
- Served directly from R2 public bucket URL or via Cloudflare CDN.

### Square

- Payments handled via Square payment links (redirect-based).
- Fulfillment triggered by webhook at `POST /api/webhooks/square`.
- Square HMAC signature validated on every webhook event.
- Unified payment provider for both in-person and online sales.

---

## Environment Variables

| Variable | Description | Example |
|---|---|---|
| `DATABASE_URL` | Neon PostgreSQL connection string | `postgres://user:pass@host/db?sslmode=require` |
| `JWT_SECRET` | Secret for signing JWT tokens | `random-32-byte-string` |
| `SQUARE_ACCESS_TOKEN` | Square API access token | `EAAAl...` |
| `SQUARE_LOCATION_ID` | Square location ID | `L...` |
| `SQUARE_WEBHOOK_SIG_KEY` | Square webhook signature key | `...` |
| `SQUARE_WEBHOOK_NOTIF_URL` | Square webhook notification URL | `https://api.retrosnack.shop/api/webhooks/square` |
| `R2_ACCOUNT_ID` | Cloudflare account ID | `abc123...` |
| `R2_ACCESS_KEY_ID` | R2 S3-compatible access key | `...` |
| `R2_SECRET_ACCESS_KEY` | R2 S3-compatible secret key | `...` |
| `R2_BUCKET_NAME` | R2 bucket name for product images | `retrosnack-media` |
| `R2_PUBLIC_URL` | Public base URL for serving images | `https://media.retrosnack.shop` |
| `PORT` | HTTP port for the Go API | `8080` |
| `ENV` | Environment (`development`/`production`) | `production` |

**GitHub Actions secrets (frontend deployment):**

| Secret | Description |
|---|---|
| `CLOUDFLARE_API_TOKEN` | Cloudflare API token with Pages edit permission |
| `CLOUDFLARE_ACCOUNT_ID` | Cloudflare account ID |
| `PUBLIC_API_URL` | Render API URL (e.g. `https://retrosnack-api.onrender.com`) |

Copy `.env.example` to `.env` for local development. Never commit `.env` to version control.
