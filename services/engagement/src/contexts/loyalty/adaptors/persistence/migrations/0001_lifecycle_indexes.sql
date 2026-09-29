-- Owner-controlled lifecycle lookup for operator diagnostics.
CREATE INDEX loyalty_reward_status ON cafe.aggregates ((state->>'status'),id) WHERE kind='reward';
