# iag-inventory

Stock levels, SKU master, and replenishment signals for the IAG platform.

| Field | Value |
|-------|-------|
| **Port** | `4006` |
| **Status** | Scaffold |
| **Remote** | [iag-inventory](https://github.com/AlexanderKiyingi/iag-inventory) |

## Planned role

System of record for **on-hand quantity** and SKU attributes. SCM retains legacy inventory APIs for coffee operations; this service will become the canonical inventory API for non-coffee and consolidated views. Platform JWT, Postgres, Kafka `iag.operations`.

## Quick start

```bash
cd services/operations/inventory
# implementation pending
```

Registry: [`subrepos.json`](../../../subrepos.json)
