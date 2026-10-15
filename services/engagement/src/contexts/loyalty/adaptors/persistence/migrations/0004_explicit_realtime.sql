-- Realtime publications may describe application outcomes without a saved aggregate.
ALTER TABLE cafe.realtime_publications
 DROP CONSTRAINT realtime_publications_aggregate_kind_aggregate_id_fkey,
 DROP CONSTRAINT realtime_publications_aggregate_kind_aggregate_id_revision_key,
 DROP CONSTRAINT realtime_publications_revision_check,
 ADD COLUMN message_kind text NOT NULL DEFAULT 'snapshot',
 ADD CONSTRAINT realtime_payload_revision CHECK (
   (message_kind='snapshot' AND revision>0) OR
   (message_kind='operation_rejected' AND revision>=0));
