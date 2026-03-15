# URL Shortener (Go)

This service is a Go rewrite of the URL shortener API using:

- `chi` for HTTP routing
- `sqlx` for query ergonomics
- `pgxpool` for PostgreSQL connection pooling

## Endpoints

### `POST /shorten`

Create a new short URL key.

Request body:

```json
{ "url": "https://example.com/some-path" }
```

Response:

```json
{ "url": "abc123X" }
```

Errors:
- `400` for missing or invalid request URL
- `500` for internal/database errors

### `GET /{key}`

Resolve a short key by redirecting to the original URL.

Response:
- `302 Found` with `Location: <long_url>` header

Errors:
- `404` when key is not found
- `500` for internal/database errors

### `GET /healthz`

Readiness endpoint that verifies database connectivity.

---

## Tech Stack

- Go `1.26`
- Router: `github.com/go-chi/chi/v5`
- CORS middleware: `github.com/go-chi/cors`
- PostgreSQL pool: `github.com/jackc/pgx/v5/pgxpool`
- SQL helper: `github.com/jmoiron/sqlx`

---

## Configuration

Environment variables:

- `HTTP_ADDR` (default: `:8000`)
- `LOG_ENABLED` (default: `true`)
- `PROFILING_ENABLED` (default: `false`)
- `PROFILING_ADDR` (default: `127.0.0.1:6060`)
- `PROFILING_RUNTIME_STATS_ENABLED` (default: `true`)
- `PROFILING_RUNTIME_STATS_INTERVAL` (default: `15s`)
- `DATABASE_URL` (default: `postgresql://postgres:postgres@localhost:5432/url_shortener?sslmode=disable`)
- `REQUEST_TIMEOUT` (default: `2s`)
- `READ_TIMEOUT` (default: `15s`)
- `WRITE_TIMEOUT` (default: `15s`)
- `IDLE_TIMEOUT` (default: `60s`)
- `CORS_ALLOW_ORIGIN` (default: `*`)
- `DB_MAX_CONNS` (default: `10`)
- `DB_MIN_CONNS` (default: `2`)
- `REDIS_ENABLED` (default: `false`)
- `REDIS_ADDR` (default: `localhost:6379`)
- `REDIS_PASSWORD` (default: empty)
- `REDIS_DB` (default: `0`)
- `REDIS_TTL` (default: `24h`)
- `REDIS_DIAL_TIMEOUT` (default: `2s`)
- `REDIS_READ_TIMEOUT` (default: `500ms`)
- `REDIS_WRITE_TIMEOUT` (default: `500ms`)
- `CB_MAX_REQUESTS` (default: `3`)
- `CB_INTERVAL` (default: `30s`)
- `CB_TIMEOUT` (default: `10s`)
- `CB_MIN_REQUESTS` (default: `5`)
- `CB_FAILURE_RATIO` (default: `0.6`)

### Redis cache + circuit breaker

When `REDIS_ENABLED=true`, the app uses cache-aside for URL resolution:

- Cache key: `url:<short_code>`
- `GET /{key}`: Redis lookup first, fallback to PostgreSQL on miss/failure
- `POST /shorten`: best-effort write-through cache set after DB save

Redis is guarded by a circuit breaker (`sony/gobreaker`). If Redis degrades,
the breaker opens and requests continue against PostgreSQL.

---

## Database

Migrations are SQL files embedded into the binary:

- `internal/infrastructure/db/migrations/0001_init_urls.sql`

Current schema:

```sql
CREATE TABLE IF NOT EXISTS urls (
    id BIGINT PRIMARY KEY,
    long_url TEXT NOT NULL,
    short_url TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## Run Locally

1. Start PostgreSQL and create DB `url_shortener`.
2. Export database URL:

```bash
export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/url_shortener?sslmode=disable"
```

3. Run server:

```bash
go run ./cmd/server
```

Server starts at `http://localhost:8000` by default.

---

## Build and Test

Run tests:

```bash
go test ./...
```

Build binary:

```bash
go build ./cmd/server
```

---

## Load Testing (k6)

k6 script path:

```text
tests/load/k6-url-shortener.js
```

Run balanced profile:

```bash
BASE_URL="http://localhost:8000" k6 run tests/load/k6-url-shortener.js
```

Run read-heavy profile:

```bash
BASE_URL="http://localhost:8000" \
PROFILE="read-heavy" \
SEED_COUNT="200" \
DURATION="60s" \
RATE="50" \
PRE_ALLOCATED_VUS="50" \
MAX_VUS="200" \
k6 run tests/load/k6-url-shortener.js
```

---

## Docker

Build image:

```bash
docker build -t url-shortener-go .
```

Run container:

```bash
docker run --rm -p 8000:8000 \
  -e HTTP_ADDR=":8000" \
  -e DATABASE_URL="postgresql://postgres:postgres@host.docker.internal:5432/url_shortener?sslmode=disable" \
  url-shortener-go
```

---

## Profiling (CPU, Memory, Runtime)

Enable profiling via env:

```bash
export PROFILING_ENABLED="true"
export PROFILING_ADDR="127.0.0.1:6060"
```

When enabled, profiling endpoints are exposed on a separate debug server:

- `GET /debug/pprof/`
- `GET /debug/pprof/profile` (CPU profile)
- `GET /debug/pprof/heap` (heap profile)
- `GET /debug/pprof/goroutine`
- `GET /debug/vars` (expvar)
- `GET /debug/stats` (JSON runtime stats)

Examples:

```bash
# CPU profile (30s)
go tool pprof http://127.0.0.1:6060/debug/pprof/profile?seconds=30

# Heap profile
go tool pprof http://127.0.0.1:6060/debug/pprof/heap

# Runtime stats JSON
curl http://127.0.0.1:6060/debug/stats
```

Optional runtime stats logging:

- `PROFILING_RUNTIME_STATS_ENABLED=true|false`
- `PROFILING_RUNTIME_STATS_INTERVAL=15s` (or any valid duration)

---

## Kubernetes (Minikube) - Traefik + PgBouncer + Redis

This repo includes Kubernetes manifests under `k8s/` for:

- App deployment with rolling updates and health probes
- Horizontal Pod Autoscaler (CPU + memory)
- Traefik ingress routing
- PgBouncer between app and PostgreSQL
- Redis for cache-aside lookup

### Prerequisites

1. Minikube running
2. Metrics server enabled (for HPA)
3. Traefik ingress controller installed in cluster

Example:

```bash
minikube start
minikube addons enable metrics-server
```

Install Traefik (Helm):

```bash
helm repo add traefik https://traefik.github.io/charts
helm repo update
helm install traefik traefik/traefik --namespace traefik --create-namespace
```

### Build image into Minikube Docker

```bash
eval $(minikube docker-env)
docker build -t url-shortener-go:latest .
```

### Deploy

```bash
kubectl apply -k k8s/
kubectl -n url-shortener get pods,svc,ingress,hpa
```

### Access app

Add local host mapping:

```bash
echo "$(minikube ip) url-shortener.local" | sudo tee -a /etc/hosts
```

Then call:

```bash
curl -i http://url-shortener.local/healthz
```

### Rolling deployment

Update image and rollout:

```bash
kubectl -n url-shortener set image deployment/url-shortener app=url-shortener-go:latest
kubectl -n url-shortener rollout status deployment/url-shortener
```

### Elastic scaling

HPA is configured (`min=2`, `max=10`).
Manual scale is also available:

```bash
kubectl -n url-shortener scale deployment/url-shortener --replicas=5
```
