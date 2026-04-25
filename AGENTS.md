# mgmt-backend — Agent Context

This file is the single source of truth for AI coding agents working in this repository. It covers repo context (purpose, stack, layout, verified commands) plus conventions for handlers, repositories, migrations, OpenFGA ACL, and config.

CE is self-contained OSS. Source, comments, prose docs, and rule files MUST NOT name specific downstream / operator-overlay repos, tooling, or file paths. Use edition-neutral language ("downstream consumers", "the operator overlay") when an external surface must be referenced.

## Purpose

`mgmt-backend` is Instill AI's management service. It owns user-related
resources across the Instill Core platform:

- **Users & organizations** — accounts, profiles, memberships, ownership.
- **Authentication & authorization** — API tokens, auth subjects, OpenFGA
  permission checks (see `pkg/acl/`).
- **Onboarding & metrics** — onboarding status, usage/metrics collection via
  InfluxDB.
- **Subscriptions / plans** — hooks for plan management (extensions
  layered on top by downstream consumers).

It is the source of truth for "who is the user/org" that every other Instill
backend (pipeline, artifact, model, agent, etc.) consults through gRPC.

## Stack

- **Language:** Go (see `go.mod` for the current toolchain version).
- **APIs:** gRPC (primary) + gRPC-Gateway for REST, protos come from
  `github.com/instill-ai/protogen-go`.
- **Persistence:** PostgreSQL via GORM (`gorm.io/gorm`, `driver/postgres`,
  `plugin/dbresolver`); schema migrations with `golang-migrate`.
- **Cache / queues:** Redis (`redis/go-redis/v9`).
- **Workflows:** Temporal (`go.temporal.io/sdk`) — async jobs live in
  `pkg/worker/` and `cmd/worker/`.
- **AuthZ:** OpenFGA (`openfga/go-sdk`) wrapped in `pkg/acl/`.
- **Observability:** OpenTelemetry (traces/metrics/logs) + Zap logging +
  InfluxDB for usage metrics.
- **Config:** `knadh/koanf` reading `config/config.yaml` + `CFG_*` env vars.

## Layout

```
cmd/
  main/       # main API server (public + private gRPC)
  migration/  # runs DB migrations (mgmt-backend-migrate binary)
  init/       # seeds default data (mgmt-backend-init binary)
  worker/     # Temporal worker (mgmt-backend-worker binary)
config/       # config.go loader, config.yaml defaults, models/
pkg/
  acl/          # OpenFGA client + permission helpers
  constant/     # shared constants
  datamodel/    # GORM models (User, Organization, Token, ...)
  db/           # DB connection + migration SQL under db/migration/
  handler/      # gRPC service implementations (public + private)
  middleware/   # gRPC interceptors (auth, logging, tracing)
  repository/   # DB access layer
  service/      # business logic orchestrating repo + acl + worker
  worker/       # Temporal workflows/activities
internal/resource/   # internal resource helpers
integration-test/    # k6 scripts (grpc.js, rest.js)
```

Four binaries are produced by the Dockerfile: the main server, migrate,
init, and worker — and the container entrypoint runs them in that order.

## Verified commands

All commands run from the repo root. Targets below were read directly from
the `Makefile`:

- `make help` — list all targets.
- `make build-dev` / `make build-latest` — build Docker images.
- `make dev` / `make latest` — run dev or released container (requires
  `instill-core` stack running on the `instill-network` docker network;
  start it via `make latest` in the `instill-core` repo first).
- `make logs` / `make stop` / `make rm` / `make top` — container ops.
- `make go-gen` — `go generate ./...`.
- `make coverage` — `go test -race -coverpkg=./... ./...`; set
  `DBTEST=true` to also run DB-backed tests (runs `cmd/migration` first
  against `TEST_DBHOST`/`TEST_DBNAME`). `HTML=true` opens the HTML report.
- `make integration-test` — k6 runs `integration-test/grpc.js` then
  `rest.js`. Override `API_GATEWAY_URL`, `API_GATEWAY_PROTOCOL`, `DB_HOST`
  when running inside the compose network (e.g. `API_GATEWAY_URL=api-gateway:8080
  DB_HOST=pg_sql`).

The `Makefile` requires a populated `.env` (it does `include .env`). The
default development flow is: bring up `instill-core`, then `make latest` or
`make dev` here.

## Conventions

- Use absolute import paths under `github.com/instill-ai/mgmt-backend/...`.
- New DB changes: add a migration under `pkg/db/migration/` and a
  corresponding model/repo/service/handler layer — do not bypass the
  service layer from handlers.
- Permission checks go through `pkg/acl/`, not ad-hoc SQL.
- Keep public vs. private gRPC handlers separate (private endpoints are
  for service-to-service calls inside the cluster and are not exposed by
  `api-gateway`).
- When adding config keys, update `config/config.go` and
  `config/config.yaml` together; env override is `CFG_<SECTION>_<KEY>`.
