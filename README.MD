Below is a **production‑ready README.md** for the shopping platform backend. It covers local development, Docker, migrations, testing, and deployment guidelines.

```markdown
# Shopping Platform Backend

Production‑grade e‑commerce backend API built with Go, Gin, GORM, PostgreSQL, and JWT authentication.

## Features

- ✅ RESTful API with Gin framework
- ✅ JWT‑based authentication
- ✅ PostgreSQL with connection pooling
- ✅ Structured logging (JSON format)
- ✅ Environment‑based configuration (Viper)
- ✅ Database migrations with Goose
- ✅ Docker & Docker Compose support
- ✅ Graceful shutdown
- ✅ CORS middleware
- ✅ Health check endpoint

## Tech Stack

| Component       | Technology                                     |
|----------------|------------------------------------------------|
| Language       | Go 1.23+                                       |
| HTTP Framework | Gin                                            |
| ORM            | GORM                                           |
| Migrations     | Goose                                          |
| Auth           | golang-jwt/jwt v5                              |
| Validation     | go-playground/validator v10                    |
| Database       | PostgreSQL 16+                                 |
| Config         | Viper (.env + env vars)                        |
| Logging        | Logrus (JSON format)                           |
| Container      | Docker + Docker Compose                        |

## Prerequisites

- Go 1.23 or later
- PostgreSQL 16 (or use Docker)
- Make
- Goose CLI (will be installed automatically via `make install-tools`)
- [Air](https://github.com/cosmtrek/air) (optional, for hot reload – installed via `make install-tools`)
- Docker & Docker Compose (optional)

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/yourusername/luxe.git
cd luxe
```

### 2. Environment setup

Copy the example environment file and adjust values as needed:

```bash
cp .env.example .env
```

Edit `.env` to set your database credentials, JWT secret, etc.

### 3. Local Development (without Docker)

#### Install tools and dependencies

```bash
make install-tools   # installs goose and air
make deps            # downloads Go modules
```

#### Run database migrations

```bash
make migrate-up
```

#### Run the application

```bash
make run
```

The server will start on `http://localhost:8080`.

#### Hot reload (development)

```bash
make watch
```

### 4. Using Docker (recommended for production‑like environment)

Build and start all services:

```bash
docker-compose up -d
```

Check logs:

```bash
docker-compose logs -f app
```

Run migrations inside the app container:

```bash
docker-compose exec app goose -dir /app/migrations postgres "postgres://postgres:postgres@postgres:5432/shopping_platform?sslmode=disable" up
```

Stop and remove containers (including the database volume):

```bash
docker-compose down -v
```

## Database Migrations

This project uses the external [Goose](https://github.com/pressly/goose) CLI. Migration files live in `internal/migrations/`.

Install Goose: `make install-tools` (pins `v3.22.1`).

### Commands

| Command                                     | Description                                    |
|---------------------------------------------|------------------------------------------------|
| `make migrate-create name=<migration_name>` | Create a new migration file (up and down)     |
| `make migrate-up`                           | Apply all pending migrations (host `goose`)   |
| `make migrate-up-docker`                    | Run `goose` inside Docker network (Windows Docker fallback) |
| `make migrate-down`                         | Rollback the last migration                   |
| `make migrate-reset`                        | Rollback all migrations (dangerous)           |
| `make migrate-status`                       | Show current migration status                 |
| `make migrate-version`                      | Show the current migration version            |
| `make migrate-force version=<version>`      | Force set migration version (use with caution)|

**Windows + Docker Desktop:** if `make migrate-up` fails with `unexpected EOF` on `localhost:5433`, use:

```bash
make migrate-up-docker
```

Or full setup:

```bash
make dev-setup   # docker-up → wait postgres → migrate-up (auto-fallback)
```

Then edit the generated `.sql` files in `internal/migrations/` and run `make migrate-up`.

## Environment Variables

| Variable        | Description                     | Default                          |
|----------------|---------------------------------|----------------------------------|
| `SERVER_PORT`  | HTTP server port                | `8080`                           |
| `GIN_MODE`     | Gin mode (`debug`, `release`)   | `debug`                          |
| `DB_HOST`      | PostgreSQL host                 | `localhost` (Docker: `postgres`) |
| `DB_PORT`      | PostgreSQL port                 | `5433` (Docker maps `5433→5432`) |
| `DB_USER`      | Database user                   | `postgres`                       |
| `DB_PASSWORD`  | Database password               | `postgres`                       |
| `DB_NAME`      | Database name                   | `shopping_platform`              |
| `DB_SSLMODE`   | PostgreSQL SSL mode             | `disable`                        |
| `JWT_SECRET`   | Secret key for JWT signing      | **required**                     |
| `LOG_LEVEL`    | Log level (`debug`, `info`, …)  | `info`                           |
| `AI_ENABLED`   | Enable LLM features             | `false`                          |
| `AI_PROVIDER`  | Provider label (`ollama`, `groq`, `gemini`) | `ollama`              |
| `AI_BASE_URL`  | OpenAI-compatible API base URL  | `http://localhost:11434/v1`      |
| `AI_API_KEY`   | API key (empty for local Ollama)| *(empty)*                        |
| `AI_MODEL`     | Model name                      | `llama3.2`                       |
| `AI_MAX_REQUESTS_PER_HOUR` | Admin generate rate limit | `20`                   |
| `AI_CHAT_MAX_REQUESTS_PER_HOUR` | Shopper chat rate limit | `30`            |

All variables can be set in a `.env` file or as system environment variables. When using Docker Compose, they are defined in the `environment` section of the `app` service.

### Local AI with Ollama (free, no API key)

1. Install [Ollama](https://ollama.com) and pull a model:

```bash
ollama pull llama3.2
```

2. In `.env`:

```bash
AI_ENABLED=true
AI_PROVIDER=ollama
AI_BASE_URL=http://localhost:11434/v1
AI_MODEL=llama3.2
```

3. Restart the API (`make run`). Endpoints:

- `POST /api/v1/admin/ai/generate` — admin product description / SEO copy
- `POST /api/v1/ai/chat` — grounded shopper product chat
- `GET /api/v1/admin/ai/status` — provider status for admin UI

**Swap to a hosted provider** without code changes — set `AI_PROVIDER`, `AI_BASE_URL`, `AI_API_KEY`, and `AI_MODEL`. Examples:

| Provider | `AI_BASE_URL` | Model example |
|----------|---------------|---------------|
| Groq | `https://api.groq.com/openai/v1` | `llama-3.3-70b-versatile` |
| Gemini (OpenAI compat) | `https://generativelanguage.googleapis.com/v1beta/openai` | `gemini-2.0-flash` |

API keys stay on the server only — never expose `AI_API_KEY` in frontend or mobile env.

## Testing

Run all tests:

```bash
make test
```

Run tests with coverage and generate an HTML report:

```bash
make test-coverage
# open coverage.html in your browser
```

Run integration tests (requires Postgres):

```bash
DATABASE_URL=postgresql://postgres:postgres@localhost:5433/shopping_platform?sslmode=disable \
JWT_SECRET=ci_test_secret \
go test ./tests/integration/...
```

## Dev demo data

Schema migrations are schema-only. For local PDP demo stores/products, after `make migrate-up`:

```bash
make seed-dev
```

Requires `psql` and `DATABASE_URL`. Do not run in production.

## OpenTelemetry tracing (optional)

Tracing is optional. The API runs fine with `OTEL_ENABLED=false` or unset.

1. Start Jaeger when Docker Hub is reachable:

```bash
make docker-up-jaeger
# or: make docker-up-all
```

If pull fails (`auth.docker.io` / DNS), fix VPN/network or skip Jaeger — the app still starts with OTEL enabled; traces export only if a collector is listening.

2. In `.env`:

```bash
OTEL_ENABLED=true
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318
```

3. Run the API (`make run`) and send requests.

## Prometheus metrics

The API exposes Prometheus metrics at `GET /metrics` (Go runtime and default HTTP metrics). No authentication is required by default — restrict access in production (reverse proxy or network policy).

```bash
curl http://localhost:8080/metrics
```

4. Open the Jaeger UI: [http://localhost:16686](http://localhost:16686) — search service `luxe-api` (or your `SERVICE_NAME`).

## Project Structure

```
.
├── cmd/
│   └── api/               # Application entry point (main.go)
├── config/                # Configuration loading (Viper)
├── controllers/           # HTTP handlers (request/response logic)
├── middleware/            # Custom middleware (CORS, auth, logging)
├── models/                # GORM data models
├── routes/                # Route definitions
├── services/              # Business logic (to be added)
├── utils/                 # Helpers (logger, validation, etc.)
├── migrations/            # Goose migration files (*.sql)
├── .env.example           # Example environment variables
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── go.mod
```

## API Endpoints (current)

| Method | Endpoint         | Description           | Auth |
|--------|------------------|-----------------------|------|
| GET    | `/api/v1/health` | Health check          | None |

More endpoints will be added in future steps (user registration, login, products, orders, etc.).

## Deployment Notes (Production)

1. **Never** run with `GIN_MODE=debug` in production. Set `GIN_MODE=release`.
2. **Use a strong, unique JWT_SECRET** – generate one with `openssl rand -base64 32`.
3. **Disable `AutoMigrate`** – the application does not perform automatic schema changes. Use Goose migrations exclusively.
4. **Set appropriate PostgreSQL pool sizes** (already configured: `MaxIdleConns=10`, `MaxOpenConns=100`).
5. **Enable SSL/TLS** – place the application behind a reverse proxy like Nginx or use `ListenAndServeTLS`.
6. **Use environment secrets** – never commit `.env` to version control.
7. **Run database migrations separately** before deploying a new version (e.g., as a Kubernetes init container or CI step).
8. **Monitor logs** – the application outputs JSON logs, suitable for log aggregation tools (ELK, Loki, etc.).
9. **Health check endpoint** – use `/api/v1/health` for liveness/readiness probes.

## Contributing

1. Create a feature branch
2. Make your changes
3. Run tests (`make test`)
4. Create a pull request

## License

[MIT](LICENSE)
```

---

This README is now ready for your shopping platform project. It covers everything from setup to production considerations. You can adjust the GitHub repository URL and license file later.

**Next step:** After you approve the README, I'll proceed to **Step 2 – User Registration & JWT Generation**.



# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"12345678","first_name":"John","last_name":"Doe"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"12345678"}'


[//]: # (goose -dir ./migrations create add_orders_indexes sql)

## Wallet deposits (Stripe)

When `STRIPE_SECRET_KEY` is set, `POST /api/v1/wallet/deposit` returns a Stripe Checkout URL (`checkout_url`) and leaves the transaction `pending` until payment succeeds.

```bash
curl -X POST http://localhost:8080/api/v1/wallet/deposit \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"amount": 50}'
```

Webhook events: `checkout.session.completed` (credit wallet), `checkout.session.expired` (mark failed). Users can also cancel pending deposits via `POST /api/v1/wallet/deposit/{id}/cancel`.

Without Stripe keys, deposits are auto-confirmed for local dev and CI.

# Redis — durable background jobs
REDIS_URL=redis://localhost:6379/0