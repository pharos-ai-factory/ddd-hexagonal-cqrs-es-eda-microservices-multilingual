@postgres @ordering
Feature: Remembering a decision while menu information catches up
  A rejected command keeps its recorded result.
  A new attempt can use a projection that arrived later.

  @ORDER_010
  Scenario: A missing menu remains a rejection on identical retry
    Given a published menu has not reached Ordering
    When the customer requests an order
    Then the decision is recorded as "menu_pending" without creating an order
    When the published menu reaches Ordering
    And the customer retries the identical command
    Then the original rejection is returned without creating an order
    When the customer makes a new attempt
    Then one draft order and one browser publication are committed

  @ORDER_011
  Scenario: A command identity cannot be reused with a different customer
    Given a published menu has reached Ordering
    When the customer requests an order
    And the same command identity is reused for another customer
    Then the new input is rejected as "idempotency_conflict"
    And the original order and its publication remain unchanged
