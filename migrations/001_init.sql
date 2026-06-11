-- iag-inventory initial schema (skeleton).
--
-- Domain tables (SKU master, on-hand ledger, stock movements) are added as the
-- inventory domain is implemented. This migration only establishes a marker so
-- the migration runner and schema are exercised from first boot.
CREATE TABLE IF NOT EXISTS inventory_service_meta (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO inventory_service_meta (key, value)
VALUES ('schema_initialized', NOW()::text)
ON CONFLICT (key) DO NOTHING;
