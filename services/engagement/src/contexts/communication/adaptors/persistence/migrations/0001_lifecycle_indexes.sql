-- Owner-controlled lifecycle lookup for operator diagnostics.
CREATE INDEX communication_notification_status ON cafe.aggregates ((state->>'status'),id) WHERE kind='notification';
