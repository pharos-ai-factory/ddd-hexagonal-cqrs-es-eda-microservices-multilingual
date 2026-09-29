-- Owner-controlled lifecycle lookup for operator diagnostics.
CREATE INDEX menu_edition_status ON cafe.aggregates ((state->>'status'),id) WHERE kind='edition';
