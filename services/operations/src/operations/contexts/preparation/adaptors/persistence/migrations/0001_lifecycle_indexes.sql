-- Owner-controlled lifecycle lookup for operator diagnostics.
CREATE INDEX preparation_ticket_status ON cafe.aggregates ((state->>'status'),id) WHERE kind='ticket';
