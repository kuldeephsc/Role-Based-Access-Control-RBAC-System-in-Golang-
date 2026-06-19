# RBAC Platform

Enterprise-grade, event-driven Role-Based Access Control platform built with Go, PostgreSQL, Redis, RabbitMQ, and a Next.js dashboard. Demonstrates clean architecture, the transactional outbox pattern, cache-aside with synchronous invalidation, observability (Prometheus + OpenTelemetry/Jaeger), and a credible path from modular monolith to gRPC microservices on Kubernetes.

## Architecture overview

See `rbac-platform-architecture-spec.md` for the full system design document including all architectural decisions, capacity estimates, failure-mode analysis, and sequence diagrams.

**Core flow**: Gin gateway → Auth/RBAC/User services (in-process, Phase 1–4) → PostgreSQL + Redis cache → Transactional outbox → RabbitMQ relay → Audit/Notification/Analytics consumers.

## Quick start (Docker Compose — recommended)

```bash
# 1. Start everything: Postgres, Redis, RabbitMQ, Jaeger, Prometheus, Grafana, app
cp .env.example .env
go mod tidy       # generates go.sum (must be run locally, not in CI without network)
make docker-up

# 2. Run migrations
make migrate-up

# 3. App is now live at http://localhost:8080
curl localhost:8080/health
```

## Quick start (local dev)

```bash
# 1. Start infra only
docker compose up -d postgres redis rabbitmq jaeger prometheus grafana

# 2. Configure
cp .env.example .env

# 3. Dependencies + migrations
go mod tidy
make migrate-up

# 4. Run
make run
```

## Frontend (Next.js)

```bash
cd frontend
npm install
npm run dev    # http://localhost:3000
```

The Next.js dev server proxies `/api/*` to `localhost:8080` automatically.

## Observability UIs

| Service    | URL                        | Notes                           |
|------------|----------------------------|---------------------------------|
| App API    | http://localhost:8080       | `/health`, `/metrics`           |
| Jaeger     | http://localhost:16686      | Trace search by service name    |
| Prometheus | http://localhost:9090       | Query `rbac_*` metrics          |
| Grafana    | http://localhost:3001       | admin/admin, add Prometheus DS  |
| RabbitMQ   | http://localhost:15672      | guest/guest                     |
| Frontend   | http://localhost:3000       | Next.js dashboard               |

## API endpoints

| Method | Path                               | Auth             |
|--------|-------------------------------------|------------------|
| POST   | `/api/v1/auth/signup`               | none             |
| POST   | `/api/v1/auth/login`                | none             |
| POST   | `/api/v1/auth/refresh`              | refresh token    |
| POST   | `/api/v1/auth/logout`               | access token     |
| GET    | `/api/v1/users`                     | `view_users`     |
| GET    | `/api/v1/users/:id`                 | self or `view_profile` |
| PATCH  | `/api/v1/users/:id`                 | self or admin    |
| POST   | `/api/v1/roles`                     | `assign_role`    |
| GET    | `/api/v1/roles`                     | authenticated    |
| DELETE | `/api/v1/roles/:id`                 | `assign_role`    |
| POST   | `/api/v1/permissions`               | `assign_role`    |
| GET    | `/api/v1/permissions`               | authenticated    |
| DELETE | `/api/v1/permissions/:id`           | `assign_role`    |
| POST   | `/api/v1/users/:id/roles`           | `assign_role`    |
| DELETE | `/api/v1/users/:id/roles/:roleId`   | `assign_role`    |
| POST   | `/api/v1/roles/:id/permissions`     | `assign_role`    |
| POST   | `/api/v1/authorize`                 | authenticated    |
| GET    | `/api/v1/audit`                     | `view_audit`     |

## Example curl session

```bash
# Sign up
curl -s -X POST localhost:8080/api/v1/auth/signup \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@test.com","password":"password123","full_name":"Admin User"}'

# Log in
TOKEN=$(curl -s -X POST localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@test.com","password":"password123"}' | jq -r .access_token)

# List roles
curl -s localhost:8080/api/v1/roles -H "Authorization: Bearer $TOKEN"

# Check a permission
curl -s -X POST localhost:8080/api/v1/authorize \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"<USER_ID>","permission":"view_profile"}'
```

## Project structure

```
rbac-platform/
├── cmd/server/main.go                    # Wiring / DI only
├── internal/
│   ├── domain/                           # Entities + repo interfaces (zero deps)
│   ├── auth/                             # Signup, login, refresh, logout
│   ├── rbac/                             # Roles, permissions, assignments, /authorize
│   ├── user/                             # User CRUD with self-or-permission checks
│   ├── audit/                            # Consumer + handler for audit logs
│   ├── notification/                     # Consumer (logs now, email later)
│   ├── analytics/                        # Consumer (logs now, warehouse later)
│   ├── outbox/                           # Transactional outbox relay → RabbitMQ
│   ├── middleware/                        # JWT auth, RBAC gate, rate limit, metrics, tracing
│   ├── repository/postgres/              # GORM implementations of all repo interfaces
│   ├── platform/{cache,jwt,logger,metrics,rabbitmq,redisclient,tracing}/
│   ├── config/                           # Env-based config loader
│   └── httpx/                            # Shared error response helper
├── migrations/                           # golang-migrate SQL files
├── proto/                                # gRPC contracts (Phase 5 target)
├── deployments/{docker,k8s}/             # Prometheus config + K8s manifests
├── frontend/                             # Next.js 14 dashboard
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── go.mod
```

## Key design decisions

1. **PostgreSQL only** — no MongoDB. Audit logs use a JSONB metadata column.
2. **Synchronous cache invalidation** — Redis DEL happens inside the request, after commit, before response. Events are for side effects only, never for cache correctness.
3. **Transactional outbox** — the business write and the event are committed in one Postgres transaction. A separate relay goroutine publishes to RabbitMQ. RabbitMQ being down never loses an event or blocks a request.
4. **Fail-safe Redis** — permission cache misses fall through to Postgres; blacklist checks fail open. A Redis outage degrades latency, never correctness or availability.
5. **Clean architecture** — domain layer has zero infrastructure deps; repository interfaces are the only seam between the service layer and Postgres.

## Seed data

The first migration seeds three roles (`admin`, `manager`, `developer`) with their permission assignments. Signup auto-assigns the `developer` role. To promote a user to admin for testing:

```sql
INSERT INTO user_roles (user_id, role_id, assigned_by)
SELECT u.id, r.id, u.id FROM users u, roles r
WHERE u.email = 'admin@test.com' AND r.name = 'admin';
```

## gRPC (Phase 5)

Proto files are in `proto/`. Generate stubs locally with:

```bash
protoc --go_out=. --go-grpc_out=. proto/*.proto
```

Requires `protoc`, `protoc-gen-go`, and `protoc-gen-go-grpc` installed.
