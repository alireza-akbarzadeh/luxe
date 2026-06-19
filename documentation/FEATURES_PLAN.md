# Luxe E-Commerce — Features Plan & Roadmap

> **Purpose:** Single source of truth for what exists today, what is missing, and what to build next — across **luxe-backend**, **luxe-front**, and **luxe-mobile**.  
> **Use this doc** when starting a new sprint or AI session: read “Current state” first, then pick work from “Next steps” in priority order.  
> **Related:** `roadmap.md` (historical phases), `documentation/architecture.md` (technical design), `AGENTS.md` (agent rules).

**Last reviewed:** June 2026

---

## 1. Executive summary

Luxe is a **production-shaped** e-commerce platform: catalog, cart, checkout (Stripe / mock / wallet), orders, shipments, returns, invoices, coupons, inventory, workflows, and a broad admin dashboard. The **storefront** and **mobile app** cover the core buy flow.

What remains is mostly **admin polish**, **menu-linked stubs**, **operational tooling**, **test hardening**, and a **minimal AI layer** (free tier / local) for merchant productivity — not a full “AI product” yet.

| Layer | Maturity | Notes |
|-------|----------|--------|
| Backend API | **Strong** | Go + Gin + GORM, workflows, Stripe, jobs, audit |
| Admin (web) | **Good, uneven** | ~20 real domains; 6 placeholder routes |
| Storefront (web) | **Strong** | Full funnel; gift cards & social auth stubbed |
| Mobile | **Good** | Core shop/cart/checkout; no admin |
| AI | **None** | No LLM integration yet |

---

## 2. What we have today

### 2.1 Backend (`luxe-backend`)

**Core commerce**

| Domain | API | Notes |
|--------|-----|--------|
| Auth & users | `/auth/*`, `/users`, `/account` | JWT, refresh, email verify, password reset |
| Catalog | `/products`, `/categories`, `/brands`, `/collections` | PDP service, search (PostgreSQL FTS), compare, likes |
| Cart & checkout | `/cart`, `/checkout` | Stock checks, coupons, multi-payment |
| Orders | `/orders` | Cancel, bulk admin status, CSV export, workflow |
| Payments | Stripe webhook, mock, wallet | Wallet deposit + pay |
| Fulfillment | `/shipments`, shipping providers | Async jobs, workflow |
| Returns | `/returns`, `/admin/returns` | Refund workflow |
| Invoices | `/admin/invoices` | Auto on payment, PDF, email |
| Coupons | `/coupons`, `/admin/coupons` | Workflow lifecycle |
| Inventory | `/admin/inventory` | Adjustments, low-stock alerts, bulk receive |
| Reviews | `/reviews` | Customer-facing; **no moderation API surface in admin UI** |

**Platform**

| Domain | API | Notes |
|--------|-----|--------|
| Workflows | `/workflows/*`, `/admin/workflows/*` | DB-driven state machine (order, product, shipment, return, coupon, brand, category, collection, user) |
| Admin analytics | `/admin/stats`, `/admin/dashboard/overview`, `/admin/reports/revenue`, `/admin/sales-feed/snapshot` | KPIs + daily revenue report |
| Roles & permissions | `/admin/roles`, module guards | RBAC on admin routes |
| Audit | `/admin/audit-logs` | |
| Settings | `/settings` | Key/value system config |
| Menus | Dashboard menu + site nav | DB-driven |
| Stores | `/stores`, `/admin/stores` | Admin CRUD UI at `/dashboard/stores` |
| Upload (R2) | Presigned uploads | Used in forms |
| Webhooks log | `/admin/webhooks` | Admin viewer at `/dashboard/settings/webhooks` |
| Wallet admin | `/admin/wallet` adjust | UI at `/dashboard/settings/wallet` |
| Import | `/admin/import/*` | Categories/products via dialogs |
| Observability | Sentry, OTEL, `/metrics`, health probes | |
| Jobs | Asynq (Redis) or in-memory | Order process, shipment, email |

**Dev & data**

- Migrations: Goose (`make migrate-up`)
- Seeds: `make seed-dev`, orders/returns, invoices, coupons, shipping providers
- Tests: unit (services), integration (Postgres), E2E placeholder
- Swagger: `make swagger` (gitignored `docs/`)

---

### 2.2 Admin frontend (`luxe-front` → `/dashboard`)

**Implemented domains** (page wired to real UI + API)

| Module | Route | Domain folder |
|--------|-------|----------------|
| Dashboard | `/dashboard` | `domains/dashboard/` |
| Live sales feed | `/dashboard/live` | `domains/sales-feed/` |
| Users | `/dashboard/users` | `domains/users/` |
| Roles | `/dashboard/roles` | `domains/roles/` |
| Audit logs | `/dashboard/audit-logs` | `domains/audit/` |
| Products | `/dashboard/products` | `domains/product-dashboard/` |
| Categories | `/dashboard/categories` | `domains/categories/` |
| Brands | `/dashboard/brands` | `domains/brands/` |
| Collections | `/dashboard/collections` | `domains/collections-admin/` |
| Inventory | `/dashboard/inventory` | `domains/inventory-admin/` |
| Orders | `/dashboard/orders` | `domains/orders/` |
| Shipments | `/dashboard/shipments` | `domains/shipments-admin/` |
| Returns | `/dashboard/returns` | `domains/returns-admin/` |
| Invoices | `/dashboard/invoices` | `domains/invoices-admin/` |
| Discounts / coupons | `/dashboard/discounts` | `domains/discounts/` |
| Revenue report | `/dashboard/reports/revenue` | `domains/revenue-report/` |
| Workflows editor | `/dashboard/workflows` | `domains/workflows/` |
| System settings | `/dashboard/settings/system` | `domains/systems/` |
| Menus | `/dashboard/menus` | `domains/menus/` |
| Notifications (admin) | `/dashboard/notifications` | `domains/admin/` |
| Shipping providers | `/dashboard/shipping-providers` | `domains/shipping-providers/` (**CRUD complete**) |

**Workflow UI** is integrated on: products, orders, shipments, returns, brands, categories, collections, coupons (edit).

**Placeholder pages** — replaced with real or roadmap UIs (June 2026)

| Route | Status |
|-------|--------|
| `/dashboard/marketing/newsletters` | Roadmap panel (API pending) |
| `/dashboard/reports/traffic` | Roadmap panel + link to revenue |
| `/dashboard/settings/gateways` | **Live** — Stripe + payment methods status |
| `/dashboard/settings/shipping` | **Live** — hub → providers & shipments |
| `/dashboard/suppliers` | Roadmap panel (API pending) |
| `/dashboard/staff` | **Live** — admin users table |

**Known admin gaps**

- Shipping providers: ~~no create/edit/detail routes~~ **done** (admin list includes inactive; create/edit/delete UI)
- API client barrels (`-admin-invoices`, etc.) **must not** be hand-written — `pnpm api:gen` wipes `src/services/`; import Orval files directly (e.g. `-admin-invoices-get.ts`)
- Discounts admin: ~~polish (status tabs, delete, KPI cards)~~ **done**
- Legacy `PUT .../status` still exists alongside workflow transitions (deprecate when UI is 100% on transitions)

---

### 2.3 Storefront (`luxe-front` → `(site)`)

**Implemented**

- Home, shop, product listing, PDP (`/product/[slug]`), search, collections, stores
- Cart, checkout (multi-step), order confirmed, order tracking
- Account: profile, addresses, orders, wallet, wishlist tab
- Wishlist, compare, notifications
- Auth: login, register, forgot/reset password, verify email
- Help & legal: static content pages

**Stubbed / partial**

- Gift cards (`/gift-cards`) — UI only, no backend entity
- Social OAuth — “coming soon” toasts
- Newsletter preferences in account — “not available yet”

**E2E:** Playwright smoke + auth + catalog flows exist (`pnpm test:smoke`).

---

### 2.4 Mobile (`luxe-mobile`)

**Implemented:** tabs (home, shop, stores, search, collections), cart, checkout, order tracking, account (incl. wishlist/wallet), compare, notifications, auth.

**Missing vs web:** help/legal, gift cards, standalone wishlist route, **no admin app**.

---

## 3. What we still need (by priority)

### P0 — Stabilize & complete in-flight admin (1–2 weeks)

These unblock daily operations and prevent regressions.

- [x] **Fix API import pattern** — audit all `@/services/-admin-*` barrel imports; use generated `-*-get.ts` / `-*-post.ts` files only
- [x] **Shipping providers admin** — create/edit/detail, fix navigation, wire Orval mutations
- [x] **Discounts admin polish** — KPI cards, status filter tabs, delete coupon, loading/error routes
- [x] **Deprecate legacy status PUT** — bulk order status removed from list UI; shipment/order status PUT marked deprecated in swagger; workflow panel is sole path on detail
- [ ] **Run `pnpm api:gen`** after backend swagger changes; commit Orval output or document regen in CI (requires running API at `OPENAPI_BASE_URL`)
- [x] **Integration tests** — coupon workflow (pause/resume), invoice on paid order, admin wallet adjust (`tests/integration/`)

### P1 — Admin surfaces for existing APIs (2–3 weeks)

Backend already supports these; build UI only.

- [x] **Store management** — `/dashboard/stores` (CRUD for multi-store / vendor pages)
- [x] **Webhook events viewer** — `/dashboard/settings/webhooks` (read-only list from `GET /admin/webhooks`)
- [x] **Wallet adjustments** — small admin tool for support credits (`POST /admin/wallet/adjust`)
- [ ] **Review moderation** — list/approve/reject (`/dashboard/reviews` + backend admin routes if missing)

### P2 — Menu stubs → real features (3–4 weeks)

Requires **new backend** work unless noted.

| Feature | Backend needed? | Suggested scope |
|---------|-----------------|-----------------|
| **Traffic report** | Yes — analytics events or Plausible/Umami integration | Start with Umami/PostHog embed + `GET /admin/reports/traffic` aggregate |
| **Newsletters** | Yes — subscribers, campaigns, send via job queue | MVP: subscriber list + export; no drag-and-drop builder yet |
| **Payment gateways** | Partial — settings UI over existing Stripe env + mock toggle | Read-only status + link to env docs; no secrets in UI |
| **Shipping settings** | Partial — default provider, zones (future) | Link to shipping-providers + order defaults |
| **Suppliers** | Yes — supplier entity, POs (later) | Phase 1: supplier directory only |
| **Staff** | Partial — extend users with staff roles / shifts | Reuse users + roles; optional staff-specific fields |

### P3 — Storefront & mobile parity (2–3 weeks)

- [ ] Gift cards — backend entity + checkout redemption
- [ ] Newsletter signup — footer + account prefs wired to P2 API
- [ ] Social login — Google OAuth (backend + front + mobile)
- [ ] Mobile — help/legal WebView or lightweight screens
- [ ] Expand Playwright — checkout happy path, admin order transition smoke

### P4 — Production hardening (ongoing)

- [ ] Email templates — order shipped, invoice, return approved (consistent branding)
- [ ] Rate limits & abuse — review public endpoints
- [ ] Backup / restore runbook for Postgres
- [ ] Staging environment + seeded demo for sales demos
- [ ] Mobile release pipeline (EAS / TestFlight)

---

## 4. Minimal AI integration (free / low-cost)

**Goal:** Merchant-facing AI that helps run the store — **not** a chatbot gimmick on day one. Start local or free-tier to avoid cost until product-market fit.

### 4.1 Recommended stack (pick one to start)

| Option | Cost | Best for | Integration |
|--------|------|----------|-------------|
| **Ollama** (local) | Free | Dev + privacy; product copy, internal tools | Backend calls `http://localhost:11434/api/generate` |
| **Groq** free tier | Free tier limits | Fast cloud inference, low latency | `GROQ_API_KEY` in backend only |
| **Google AI Studio (Gemini)** | Free tier | Good JSON + long context | `GEMINI_API_KEY` in backend |
| **OpenRouter** | Some free models | Model switching | Single API key |

**Architecture rule:** All LLM calls go through **`luxe-backend`** (`internal/integrations/ai/`). Never expose API keys to the browser or mobile app.

```
Admin UI / Storefront
    → POST /api/v1/ai/{feature}
        → ai.Service (prompt templates, guardrails)
            → Provider (Ollama | Groq | Gemini)
        → audit log (prompt hash, user id, no PII in logs)
```

### 4.2 Phase AI-1 — Admin copilot (MVP, ~1 week)

**Features**

1. **Product description generator** — on product create/edit: button “Generate description” from title, brand, category, attributes
2. **SEO meta helper** — slug + meta description suggestions
3. **Coupon naming** — suggest code + description from discount rules

**Backend**

- [ ] `internal/integrations/ai/client.go` — interface + Ollama/Groq implementation
- [ ] `internal/config` — `AI_PROVIDER`, `AI_BASE_URL`, `AI_API_KEY`, `AI_MODEL` (default `llama3.2` or `gemini-2.0-flash`)
- [ ] `POST /admin/ai/generate` — body: `{ task, context }` → `{ text }`
- [ ] Rate limit: 20 req/hour per admin user
- [ ] Feature flag: `AI_ENABLED=false` by default in production until tested

**Frontend**

- [ ] `domains/ai/` — `useAiGenerate()` hook calling admin endpoint
- [ ] Wire into `product-dashboard` form and `discounts` form only (smallest surface)

### 4.3 Phase AI-2 — Customer support (optional, ~2 weeks)

- [ ] **FAQ assistant** on `/help` — RAG over static FAQ markdown (no hallucination: cite sources)
- [ ] Backend: embed FAQ chunks (simple: precomputed JSON + keyword match first; upgrade to embeddings later)
- [ ] **Order status Q&A** — authenticated: “Where is my order?” → fetch order + shipment, LLM formats answer (facts from DB only)

### 4.4 Phase AI-3 — Smarter catalog (later)

- [ ] Semantic search — vector column on products (pgvector) + embedding job on publish
- [ ] “Similar products” beyond rule-based related
- [ ] Inventory forecasting suggestions (read-only insights)

### 4.5 What NOT to do early

- Autonomous agents changing prices or inventory without human approval
- Customer-facing open-ended chat without grounding
- Sending customer PII to third-party LLMs without disclosure + DPA

---

## 5. Suggested execution order (follow this sequence)

Use this as a **checklist**. Complete each block before jumping ahead unless blocked.

### Block A — Housekeeping (now)

1. Fix remaining `-admin-*` barrel imports across front — **done** (admin imports use per-endpoint `@/services/-admin-*` files)
2. `make swagger` + `pnpm api:gen` after any backend route change
3. `pnpm check` + `make test` before merging admin features
4. Document env vars in `.env.example` (both repos)

### Block B — Finish admin operations (P0)

5. ~~Shipping providers CRUD UI~~ **done**
6. Discounts admin polish + coupon workflow QA
7. Invoice + revenue report smoke test on seeded data
8. Workflow transition as sole path on order/shipment detail (remove duplicate status dropdowns) — **done** (bulk list actions removed; detail uses `EntityWorkflowPanel`)

### Block C — Wire existing APIs (P1)

9. ~~Admin stores page~~ **done**
10. Webhook events viewer
11. Wallet adjust tool (support)
12. Review moderation (if reviews are public on PDP)

### Block D — Reports & marketing (P2)

13. Traffic report MVP (embed or lightweight backend)
14. Newsletter subscribers MVP
15. Gateway/shipping **settings** pages (configuration visibility, not full payment UI)

### Block E — AI MVP (AI-1)

16. Backend `ai` integration package + config
17. `POST /admin/ai/generate` + audit
18. Product description button in admin
19. Local Ollama setup in `README` / `make dev-setup` notes

### Block F — Storefront gaps (P3)

20. Gift cards backend + UI
21. Newsletter on storefront + account
22. Google OAuth

### Block G — Mobile & release (P4)

23. Parity gaps on mobile
24. CI: backend integration + front smoke on PR

---

## 6. Definition of done (per feature)

A feature is **done** when:

- [ ] Backend: controller → service → GORM, swagger comments, `make swagger`, relevant tests
- [ ] Frontend: uses Orval clients, `useAppForm` / `DataTable` patterns, loading + error routes
- [ ] Admin: KPI or workflow column where siblings have them
- [ ] Seeds or scripts for local demo (`scripts/seed-*.sql`, makefile target)
- [ ] `pnpm check` (front) and `go build` / `make test` (back) pass
- [ ] No hand-written files in `src/services/` except documented exceptions

---

## 7. Repository map (quick reference)

```
luxe-backend/
  cmd/api/                 # main
  internal/
    services/              # business logic + GORM
    controllers/           # HTTP
    routes/                # route groups
    migrations/            # Goose SQL
    integrations/          # stripe, r2, (future: ai)
  scripts/                 # seed SQL
  documentation/
    architecture.md
    FEATURES_PLAN.md       # ← this file
    roadmap.md             # historical phases

luxe-front/
  src/app/(site)/          # storefront routes
  src/app/(admin)/dashboard/  # admin routes
  src/domains/             # feature modules
  src/services/            # Orval-generated API (do not hand-edit barrels)

luxe-mobile/
  app/                     # Expo routes
  src/features/            # feature modules
```

---

## 8. Notes for future sessions (AI / developers)

1. **Read this file + `architecture.md`** before large features.
2. **Prefer extending existing domains** over new patterns (orders-admin, invoices-admin, revenue-report are templates).
3. **Workflow first** for lifecycle entities — add `workflow_state_id` + seed transitions (see `brand_workflow.sql`, `coupon` workflow).
4. **Admin list APIs** belong under `/admin/<resource>` with `ListAdmin` — public `List()` stays customer-scoped.
5. **AI:** backend proxy only; start with Ollama locally; ship one button (product description) before building chat UI.

---

## 9. Changelog

| Date | Change |
|------|--------|
| 2026-06 | Initial plan: inventory of stack, gaps, P0–P4 priorities, AI-1–3 roadmap |
