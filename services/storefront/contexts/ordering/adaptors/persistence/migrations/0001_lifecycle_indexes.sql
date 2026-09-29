-- Owner-controlled lifecycle lookup for operator diagnostics.
CREATE INDEX ordering_order_status ON cafe.aggregates ((state->>'status'),id) WHERE kind='order';
