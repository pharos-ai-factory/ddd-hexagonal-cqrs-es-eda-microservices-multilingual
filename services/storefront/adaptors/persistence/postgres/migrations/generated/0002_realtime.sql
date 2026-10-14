-- Applied by the development migration administrator, never by runtime code.
BEGIN;
CREATE TABLE IF NOT EXISTS cafe.schema_migrations(version integer PRIMARY KEY, checksum text NOT NULL);
INSERT INTO cafe.schema_migrations VALUES(2, :'checksum') ON CONFLICT DO NOTHING;
-- Reapplying the same migration is safe; modifying an applied migration is not.
SELECT 1 / CASE WHEN checksum=:'checksum' THEN 1 ELSE 0 END
 FROM cafe.schema_migrations WHERE version=2;
CREATE TABLE IF NOT EXISTS cafe.realtime_publications(
 id uuid PRIMARY KEY, channel text NOT NULL, aggregate_kind text NOT NULL,
 aggregate_id uuid NOT NULL, revision bigint NOT NULL CHECK(revision>0),
 body bytea NOT NULL, created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 FOREIGN KEY(aggregate_kind,aggregate_id) REFERENCES cafe.aggregates(kind,id),
 UNIQUE(aggregate_kind,aggregate_id,revision)
);
CREATE TABLE IF NOT EXISTS cafe.realtime_dispatches(
 event_id uuid PRIMARY KEY REFERENCES cafe.realtime_publications(id),
 available_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 lease_token uuid, lease_until timestamptz, generation bigint NOT NULL DEFAULT 0,
 completed_at timestamptz, last_error text
);
CREATE INDEX IF NOT EXISTS pending_realtime_dispatch
 ON cafe.realtime_dispatches(available_at) WHERE completed_at IS NULL;
GRANT SELECT,INSERT ON cafe.realtime_publications TO :"role";
GRANT SELECT,INSERT,UPDATE ON cafe.realtime_dispatches TO :"role";
GRANT SELECT ON cafe.schema_migrations TO :"role";
COMMIT;
