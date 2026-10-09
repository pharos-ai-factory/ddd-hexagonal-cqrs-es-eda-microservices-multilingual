-- A receiving receipt and exact command bytes share one immutable row.
CREATE TABLE cafe.internal_commands(
 id uuid PRIMARY KEY, consumer text NOT NULL, event_id uuid NOT NULL,
 fingerprint text NOT NULL, target uuid NOT NULL, body bytea NOT NULL CHECK(octet_length(body)>0),
 created_at timestamptz NOT NULL DEFAULT clock_timestamp(), UNIQUE(consumer,event_id));
CREATE TABLE cafe.internal_command_dispatches(
 event_id uuid PRIMARY KEY REFERENCES cafe.internal_commands(id),
 available_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 lease_token uuid, lease_until timestamptz, generation bigint NOT NULL DEFAULT 0,
 completed_at timestamptz, last_error text);
CREATE INDEX pending_internal_command ON cafe.internal_command_dispatches(available_at) WHERE completed_at IS NULL;
GRANT SELECT,INSERT ON cafe.internal_commands TO cafe_preparation;
GRANT SELECT,INSERT,UPDATE ON cafe.internal_command_dispatches TO cafe_preparation;
