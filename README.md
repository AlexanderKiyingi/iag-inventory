# iag-inventory

Stock levels, SKU master, and replenishment signals for the IAG platform.

| Field | Value |
|-------|-------|
| **Port** | `4006` |
| **Status** | Platform skeleton (domain pending) |
| **Audience** | `iag.inventory` |
| **Remote** | [iag-inventory](https://github.com/AlexanderKiyingi/iag-inventory) |

## Planned role

System of record for **on-hand quantity** and SKU attributes. SCM retains legacy inventory APIs for coffee operations; this service will become the canonical inventory API for non-coffee and consolidated views. Platform JWT, Postgres, Kafka `iag.operations`.

## Status

Platform plumbing is in place and builds: gin server, Postgres pool (schema
`inventory`) + embedded migration runner, JWT Bearer+aud verification with JWKS
refresh, OpenTelemetry tracing, boot-time permission registration with
`iag-authentication`, and `/health` + `/ready` probes. The `/api/v1` group is
auth-gated and exposes a placeholder `/overview`.

**Domain not yet implemented** — SKU master, on-hand ledger, stock movements,
and `iag.operations` event emission are the next vertical slice and need a
signed-off data model (SKU attributes, multi-location, finance posting).

## Quick start

```bash
cd services/operations/inventory
cp .env.example .env   # set DATABASE_URL + auth vars
go run ./cmd/server
```

Layout mirrors the platform service template (`cmd/server`, `internal/{config,
db,middleware,handlers,migrate,models}`, `migrations/`).

Registry: [`subrepos.json`](../../../subrepos.json)
