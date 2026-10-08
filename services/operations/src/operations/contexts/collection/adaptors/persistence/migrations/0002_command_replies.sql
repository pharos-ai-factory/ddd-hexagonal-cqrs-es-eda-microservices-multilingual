-- Exact command response bytes are committed with their receiving transaction.
CREATE TABLE cafe.command_replies(
 id uuid PRIMARY KEY, body bytea NOT NULL CHECK(octet_length(body)>0),
 created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 expires_at timestamptz NOT NULL DEFAULT clock_timestamp()+interval '30 seconds');
CREATE TABLE cafe.command_reply_dispatches(
 event_id uuid PRIMARY KEY REFERENCES cafe.command_replies(id),
 available_at timestamptz NOT NULL DEFAULT clock_timestamp(), attempts integer NOT NULL DEFAULT 0,
 lease_token uuid, lease_until timestamptz, generation bigint NOT NULL DEFAULT 0,
 completed_at timestamptz, last_error text);
CREATE INDEX pending_command_reply ON cafe.command_reply_dispatches(available_at) WHERE completed_at IS NULL;
GRANT SELECT,INSERT ON cafe.command_replies TO cafe_collection;
GRANT SELECT,INSERT,UPDATE ON cafe.command_reply_dispatches TO cafe_collection;
