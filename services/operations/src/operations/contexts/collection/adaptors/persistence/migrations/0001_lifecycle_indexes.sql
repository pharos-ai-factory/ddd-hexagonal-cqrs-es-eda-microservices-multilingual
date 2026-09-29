-- Owner-controlled lifecycle lookup for operator diagnostics.
CREATE INDEX collection_pickup_status ON cafe.aggregates ((state->>'status'),id) WHERE kind='pickup';
