# Architecture Decision Records

## ADR-001: Modular Monolith over Microservices

**Status:** Accepted
**Date:** 2026-03-05

**Context:** retrosnack is an early-stage project with a small team. True microservices add operational
complexity (service discovery, distributed tracing, inter-service auth) without benefit at this scale.

**Decision:** One Go binary with domain-separated packages (`internal/auth`, `internal/catalog`, etc.)
that mirror microservice boundaries. Single Render deployment.

**Consequences:** Simple deployment and debugging. Can be extracted to true microservices later by
promoting each `internal/` package to its own repo and adding gRPC/HTTP transport.

---

## ADR-002: SvelteKit PWA on Cloudflare Pages

**Status:** Accepted
**Date:** 2026-03-05

**Context:** Mobile-first shopping experience. Customers browse on phones; installability improves
retention. Cloudflare Pages is free with global CDN and no egress fees.

**Decision:** SvelteKit with `vite-plugin-pwa`, deployed to Cloudflare Pages via `adapter-cloudflare`.

**Consequences:** Zero hosting cost. App is installable on Android/iOS home screens. Service worker
enables offline browsing of cached pages.

---

## ADR-003: Neon PostgreSQL

**Status:** Accepted
**Date:** 2026-03-05

**Context:** Need a managed PostgreSQL with a free tier that works well with Fly.io.

**Decision:** Neon serverless PostgreSQL. Free tier: 0.5 GB, auto-suspend when idle.

**Consequences:** Cold-start latency on first connection after idle period. Acceptable for low-traffic
MVP. Uses pgx/v5 connection pool with `pgxpool` to amortize connection cost.

---

## ADR-004: Cloudflare R2 for Media

**Status:** Accepted
**Date:** 2026-03-05

**Context:** Product images need reliable object storage. S3 charges egress; Cloudflare R2 does not.

**Decision:** Cloudflare R2 with S3-compatible API. 10 GB free storage, no egress fees.

**Consequences:** Images served directly from R2 public URL or via Cloudflare CDN. Go `media` module
uses `aws-sdk-go-v2` with a custom R2 endpoint.

---

## ADR-005: Square for Payments

**Status:** Accepted (supersedes original Stripe decision)
**Date:** 2026-03-06

**Context:** retrosnack uses Square for in-person sales. Using the same provider for online payments
unifies transaction management, reporting, and inventory across both channels. Custom card forms
require PCI compliance overhead.

**Decision:** Square payment links (redirect-based). Webhook at `POST /api/webhooks/square` fulfills
orders. Square HMAC signature validates every webhook event. Uses `square-go-sdk` Go client.

**Consequences:** No PCI scope on our servers. Order fulfillment is event-driven, not synchronous.
Single payment provider for in-person and online sales simplifies bookkeeping and reconciliation.

---

## ADR-006: Instagram oEmbed (No API Auth)

**Status:** Accepted
**Date:** 2026-03-05

**Context:** retrosnack's Instagram (@retrosnack.shop) is the primary product discovery channel.
Each product should link to its Instagram post. Instagram's Graph API requires app review for
advanced access; oEmbed works for public posts with no auth.

**Decision:** Store `instagram_post_url` per product. Render post embed using Instagram oEmbed API
on product pages. Cache `embed_html` in `instagram_links` table. Fall back to direct link if
oEmbed is unavailable.

**Consequences:** Works immediately, no API approval needed. If Instagram deprecates oEmbed for
unauthenticated use, we fall back gracefully to a link.

---

## ADR-007: sqlc for Type-Safe SQL

**Status:** Accepted
**Date:** 2026-03-05

**Context:** ORMs add magic and performance overhead. Raw SQL is verbose and error-prone.

**Decision:** sqlc generates type-safe Go code from SQL queries. goose manages migrations. pgx/v5
is the driver.

**Consequences:** SQL is the source of truth. Schema changes require migration + sqlc regeneration.
No runtime reflection overhead.

---

## ADR-008: Render over Google Cloud Run for API Hosting

**Status:** Accepted (supersedes previous Cloud Run and Fly.io decisions)
**Date:** 2026-03-07

**Context:** Fly.io removed its free tier. Google Cloud Run's Workload Identity Federation proved
problematic — new GCP accounts with non-Gmail emails are frequently flagged by automated ToS
enforcement, making reliable setup difficult. A simpler, truly free alternative is needed.

**Decision:** Render free tier. 750 hours/month, auto-deploy from GitHub on push to `main`.
Docker-based deployment using the existing multi-stage Dockerfile. Managed HTTPS, no nginx
sidecar required. Secrets managed via Render dashboard. UptimeRobot pings `/health` every 5 min
to prevent idle sleep.

**Consequences:**
- Simplest deployment setup of all options — connect GitHub repo, set env vars, done.
- Free tier sleeps after 15 min idle; UptimeRobot keep-warm eliminates this for active hours.
- No GCP project, Artifact Registry, Secret Manager, or WIF configuration needed.
- Cold starts after sleep are ~30s (Go binary boot + Neon connection); acceptable for low traffic.
- Service blueprint defined in `render.yaml` at project root.
