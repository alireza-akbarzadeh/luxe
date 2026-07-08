# Deploy luxe API to Render (Neon Postgres)

Use this when the database already lives on **Neon** and you only need to host the Go API on Render.

## Architecture

```
Browser / Vercel (luxe-front)
        ↓  NEXT_PUBLIC_API_URL
Render Web Service (luxe-backend Docker)
        ↓  DATABASE_URL
Neon Postgres (already configured)
```

Redis is **optional**. Without `REDIS_URL`, background jobs use an in-memory queue (fine for early deploys; add Render Redis later for durability).

---

## 1. Prerequisites

- GitHub repo for `luxe-backend` connected to Render
- Neon project with connection string (use the **pooled** URL, `sslmode=require`)
- Strong `JWT_SECRET` (Render can generate one)
- Storefront URL for CORS (e.g. `https://your-app.vercel.app`)

---

## 2. Create the Web Service

### Option A — Blueprint (recommended)

1. Render Dashboard → **New** → **Blueprint**
2. Point at `luxe-backend` repo (uses root `render.yaml`)
3. After deploy, open the service → **Environment** and set:

| Variable | Example / notes |
|----------|-----------------|
| `DATABASE_URL` | `postgresql://user:pass@ep-xxx-pooler.aws.neon.tech/neondb?sslmode=require` |
| `JWT_SECRET` | Long random string (or use generated value) |
| `FRONTEND_URL` | `https://your-storefront.vercel.app` |
| `CORS_ALLOW_ORIGINS` | Same as frontend URL, comma-separated if multiple |
| `STRIPE_*` | If payments enabled |
| `R2_*` | If media uploads enabled |
| `REDIS_URL` | Optional — Render Redis internal URL |

### Option B — Manual Docker service

1. **New** → **Web Service** → connect repo
2. **Language:** Docker
3. **Health check path:** `/health/live`
4. Set env vars from the table above
5. `APP_ENV=production`, `GIN_MODE=release`

---

## 3. Migrations

The Docker image runs Goose on startup (`scripts/docker-entrypoint.sh`) when `DATABASE_URL` is set.

To skip auto-migrate (e.g. run migrations manually):

```bash
SKIP_MIGRATIONS=1
```

Manual migrate from your machine:

```bash
export DATABASE_URL="postgresql://...neon...?sslmode=require"
make migrate-up
```

---

## 4. Verify API

After deploy (URL like `https://luxe-api-xxxx.onrender.com`):

```bash
curl -s https://YOUR-SERVICE.onrender.com/health/live
curl -s https://YOUR-SERVICE.onrender.com/api/v1/health/ready
```

OpenAPI (for frontend codegen): `https://YOUR-SERVICE.onrender.com/openapi`

---

## 5. Connect luxe-front on Vercel

1. Vercel → import `luxe-front` repo (framework: **Next.js**)
2. **Settings → Environment Variables** (Production):

| Variable | Value |
|----------|--------|
| `NEXT_PUBLIC_API_URL` | `https://luxe-3pvz.onrender.com/api/v1` |
| `BACKEND_API_URL` | `https://luxe-3pvz.onrender.com/api/v1` |
| `NEXT_PUBLIC_SITE_URL` | `https://your-project.vercel.app` (or custom domain) |

3. **Redeploy** the frontend after saving env vars.

The Vercel build uses committed `openapi3.json` for Orval — no live API needed at build time.

On **Render** (API), set CORS to match Vercel:

```env
FRONTEND_URL=https://your-project.vercel.app
CORS_ALLOW_ORIGINS=https://your-project.vercel.app
```

If you add a custom domain on Vercel, include it in `CORS_ALLOW_ORIGINS` (comma-separated, no trailing slashes).

Full Vercel checklist: `luxe-front/documentation/DEPLOY_VERCEL.md`

---

## 6. Stripe webhooks (if used)

In Stripe Dashboard → Webhooks, endpoint:

```
https://YOUR-SERVICE.onrender.com/api/v1/webhooks/stripe
```

Set `STRIPE_WEBHOOK_SECRET` on Render. Required when Stripe keys are set in production.

---

## 7. Optional: Redis on Render

1. **New** → **Redis** (or uncomment `luxe-redis` in `render.yaml`)
2. Copy **Internal Redis URL** → `REDIS_URL` on the API service
3. Redeploy API

---

## 8. Troubleshooting

| Symptom | Fix |
|---------|-----|
| Service won't start | Check logs — `JWT_SECRET` and `DATABASE_URL` required |
| `sslmode=disable` error | Neon URL must include `sslmode=require` |
| CORS errors in browser | Set `CORS_ALLOW_ORIGINS` to exact frontend origin (no trailing slash). `FRONTEND_URL` is merged automatically. |
| Swagger "Try it out" fails on Render | Spec host was `localhost:8080` — redeploy after `configureSwaggerInfo` (uses `RENDER_EXTERNAL_URL`). Or set `SWAGGER_HOST=luxe-3pvz.onrender.com`. |
| 502 on cold start | Free/starter tier spins down; first request may take ~30s |
| Migrations failed | Run `make migrate-up` locally with same `DATABASE_URL` |
