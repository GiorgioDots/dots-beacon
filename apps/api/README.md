# dots-beacon API

The HTTP API service. This README explains **where code goes** so you can add
features confidently.

## Layout

```
apps/api/
├── cmd/
│   └── api/
│       └── main.go          # composition root — wires everything, starts the server
└── internal/                # private to this app (Go enforces this)
    ├── config/              # Config struct + Load() from env
    ├── server/              # gin engine, shared middleware, base routes, lifecycle
    └── site/                # a FEATURE slice (domain + repo + service + handler)
```

Cross-service code (telemetry, auth, database) lives in the **shared module**
[`internal/`](../../internal/) at the repo root, not here.

## The pattern: feature slices

Each business concept (sites, devices, users, …) is **one package** under
`internal/`, containing the full vertical slice. The request flows in one
direction:

```
HTTP request → handler → service → repository → database
                (gin)    (logic)   (sqlc)
```

Using `site` as the template:

| File | Layer | Responsibility |
|------|-------|----------------|
| [`site/site.go`](internal/site/site.go) | domain | The model the API exposes (`Site`), with JSON tags. Decoupled from the DB row. |
| [`site/repository.go`](internal/site/repository.go) | data | Wraps the sqlc `*db.Queries`; maps DB rows ↔ domain models. The only layer that knows about pgx/sqlc. |
| [`site/service.go`](internal/site/service.go) | business | Validation, rules, orchestration. Handlers stay thin; this is where logic lives. |
| [`site/handler.go`](internal/site/handler.go) | transport | gin handlers + `RegisterRoutes`. Parses input, calls the service, writes JSON. |

Dependencies point inward and are passed in via constructors (`NewRepository`,
`NewService`, `NewHandler`) — no globals, no DI framework. They're wired once in
[`cmd/api/main.go`](cmd/api/main.go).

## Adding a feature (e.g. `device`)

1. **SQL** — add queries to [`internal/database/queries/`](../../internal/database/queries/)
   (shared module) and run `sqlc generate` from the repo root.
2. **Package** — create `internal/device/` with `device.go`, `repository.go`,
   `service.go`, `handler.go` mirroring `site/`.
3. **Routes** — implement `RegisterRoutes(r gin.IRouter)` on the handler.
4. **Wire** — in `main.go`, add:
   ```go
   devices := device.NewHandler(device.NewService(device.NewRepository(queries)))
   srv := server.New(cfg, authenticator, sites, devices) // add to the list
   ```
That's it — telemetry, auth, and graceful shutdown apply automatically.

## Routes & auth

`server.New` mounts:
- `GET /healthz` — public.
- everything else under one group that requires a valid Keycloak token when
  auth is enabled (`KEYCLOAK_ISSUER_URL` set). `GET /me` and all feature routes
  live here.

To add a **public** route, register it on the engine before the group (edit
`server.go`); by default feature routes are authenticated.

## Config

Add settings as fields on `config.Config` with an `env:"..."` tag; read them
where needed (passed in from `main`). See [`internal/config/config.go`](internal/config/config.go).

## Run

```bash
task dev-api          # builds ./cmd/api and runs it (brings up the stack first)
```

Endpoints (dev): `http://localhost:8080` — `/healthz`, `/sites`, `/me`.
Auth/telemetry/storage details: [docs/auth](../../docs/auth/README.md),
[docs/telemetry](../../docs/telemetry/README.md).

## Testing (when you want it)

The service layer is the sweet spot for unit tests. To mock the DB, define a
small interface for what the service needs and have `Service` accept it:

```go
type repository interface {
    List(ctx context.Context) ([]Site, error)
}
type Service struct { repo repository }
```

Then pass a fake `repository` in tests and the real `*Repository` in `main`.
("Accept interfaces, return structs.")
