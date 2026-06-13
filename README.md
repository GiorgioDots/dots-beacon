# Dots Beacon

TODO: Description

# **_Possible_** Folder Structure

```
dots-beacon/
│
├── apps/ # runnable applications/services
│ ├── api/ # main Go API (REST/OIDC/admin/etc)
│ ├── gateway/ # realtime websocket/SSE gateway
│ ├── agent/ # Raspberry/edge device agent
│ ├── web/ # SolidJS frontend
│ │
│ ├── worker-recognition/ # image/AI processing worker
│ ├── worker-notifications/ # emails/webhooks/realtime notifications
│ ├── worker-storage/ # media/storage processing
│ └── worker-analytics/ # metrics/occupancy/etc
│
├── internal/ # private shared Go code
│ ├── auth/
│ ├── permissions/
│ ├── events/
│ ├── telemetry/
│ ├── storage/
│ ├── messaging/
│ ├── devices/
│ ├── sites/
│ └── database/
│
├── packages/ # shared cross-language packages/contracts
│ ├── sdk-go/
│ ├── sdk-ts/
│ ├── shared-types/
│ ├── event-schemas/
│ └── protobuf/
│
├── deploy/
│ ├── compose/ # docker compose files
│ ├── docker/ # dockerfiles
│ ├── k8s/ # later, optional kubernetes
│ ├── traefik/ # reverse proxy configs
│ └── scripts/ # deployment/bootstrap scripts
│
├── observability/
│ ├── grafana/
│ │ ├── dashboards/
│ │ └── datasources/
│ │
│ ├── prometheus/
│ ├── loki/
│ ├── tempo/
│ └── otel-collector/
│
├── docs/
│ ├── architecture/
│ ├── adr/ # architecture decision records
│ ├── diagrams/
│ ├── event-flow/
│ ├── permissions/
│ └── deployment/
│
├── .github/
│ └── workflows/
│
├── .env.example
├── docker-compose.yml
├── go.work
├── README.md
└── LICENSE
```

# Dev requirements

Requirements for developments:

- [task](https://taskfile.dev/)
- [go >1.26.3](https://go.dev/doc/install)
- [sqlc](https://docs.sqlc.dev/en/latest/index.html)
- [golang-migrate](https://github.com/golang-migrate/migrate)

# Develop

Use [task](https://taskfile.dev/) to run the project locally

# Schema change

- Add a new migration file with this format `{ddMMyyyyHHMM_description.up.sql}` and `*.down.sql`
- Run `sqlc generate` from the root folder
- Run `task migrate-up` to update the local db
  - If it fails, check what failed, fix / drop the migration and re-run it.
