CREATE SCHEMA cafe;
CREATE TABLE cafe.schema_version(version integer PRIMARY KEY CHECK(version=1), checksum text NOT NULL);
INSERT INTO cafe.schema_version VALUES(1, :'checksum');
CREATE TABLE cafe.aggregates(
 kind text NOT NULL, id uuid NOT NULL, version bigint NOT NULL CHECK(version>0),
 state jsonb NOT NULL CHECK(jsonb_typeof(state)='object'), PRIMARY KEY(kind,id)
);
CREATE FUNCTION cafe.guard_aggregate_write() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE target text; touched text;
BEGIN
 target := NEW.kind || ':' || NEW.id::text;
 IF current_setting('cafe.command_target',true) IS DISTINCT FROM target THEN
  RAISE EXCEPTION 'aggregate write outside the command target';
 END IF;
 touched := NULLIF(current_setting('cafe.touched_aggregate',true),'');
 IF touched IS NOT NULL AND touched <> target THEN
  RAISE EXCEPTION 'one transaction cannot change two aggregates';
 END IF;
 PERFORM set_config('cafe.touched_aggregate',target,true);
 IF (TG_OP='INSERT' AND NEW.version<>1) OR
    (TG_OP='UPDATE' AND (NEW.version<>OLD.version+1 OR NEW.kind<>OLD.kind OR NEW.id<>OLD.id)) THEN
  RAISE EXCEPTION 'aggregate identity or version violation';
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER aggregate_write_guard BEFORE INSERT OR UPDATE ON cafe.aggregates
 FOR EACH ROW EXECUTE FUNCTION cafe.guard_aggregate_write();
CREATE TABLE cafe.command_receipts(
 kind text NOT NULL, aggregate_id uuid NOT NULL, command_name text NOT NULL,
 command_id uuid NOT NULL, fingerprint text NOT NULL, outcome jsonb NOT NULL,
 correlation_id uuid NOT NULL, causation_id uuid,
 created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 PRIMARY KEY(kind,aggregate_id,command_name,command_id)
);
CREATE TABLE cafe.consumer_receipts(
 consumer text NOT NULL, event_id uuid NOT NULL, fingerprint text NOT NULL,
 target text NOT NULL, outcome jsonb NOT NULL,
 completed_at timestamptz NOT NULL DEFAULT clock_timestamp(), PRIMARY KEY(consumer,event_id)
);
CREATE TABLE cafe.outbox_events(
 id uuid PRIMARY KEY, event_name text NOT NULL, visibility text NOT NULL CHECK(visibility IN('domain','integration')),
 aggregate_kind text NOT NULL, aggregate_id uuid NOT NULL, aggregate_version bigint NOT NULL,
 correlation_id uuid NOT NULL, causation_id uuid NOT NULL, body bytea NOT NULL,
 created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 FOREIGN KEY(aggregate_kind,aggregate_id) REFERENCES cafe.aggregates(kind,id)
);
CREATE TABLE cafe.dispatches(
 event_id uuid PRIMARY KEY REFERENCES cafe.outbox_events(id),
 available_at timestamptz NOT NULL DEFAULT clock_timestamp(), attempts integer NOT NULL DEFAULT 0,
 lease_token uuid, lease_until timestamptz, generation bigint NOT NULL DEFAULT 0,
 completed_at timestamptz, last_error text
);
CREATE INDEX pending_dispatch ON cafe.dispatches(available_at) WHERE completed_at IS NULL;
CREATE TABLE cafe.projections(
 name text NOT NULL, key text NOT NULL, revision bigint NOT NULL CHECK(revision>0),
 state jsonb NOT NULL, PRIMARY KEY(name,key)
);
REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON ALL FUNCTIONS IN SCHEMA cafe FROM PUBLIC;
