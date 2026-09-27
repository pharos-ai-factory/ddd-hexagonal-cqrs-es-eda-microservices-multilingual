@integration @workflows
Feature: Durable workflows across the three languages
  A delivery may repeat with a different envelope identity.
  Business identity and recorded decisions still protect each aggregate.

  Background:
    Given a published menu is available to Ordering

  @FLOW_001
  Scenario: A new delivery identity cannot credit a collection twice
    Given the customer has collected 2 orders
    When the last OrderCollected fact is delivered with a new event identity
    Then the new delivery has a committed consumer receipt
    And the account still has 2 collections, 2 stamps and 0 grants
    And the customer has no reward

  @FLOW_002
  Scenario: A repeated earned grant creates only one reward
    Given the customer has collected 3 orders and received a reward
    When the RewardEarned fact is delivered with a new event identity
    Then the new delivery has a committed consumer receipt
    And the original reward and its single notification are unchanged

  @FLOW_003
  Scenario: A repeated placed order cannot reset a ready ticket
    Given the customer's order is ready for collection
    When the OrderPlaced fact is delivered with a new event identity
    Then the new delivery has a committed consumer receipt
    And the original ready ticket is unchanged

  @FLOW_004
  Scenario: A repeated ready-drinks fact cannot reopen collection
    Given the customer has collected 1 order
    When the DrinksReady fact is delivered with a new event identity
    Then the new delivery has a committed consumer receipt
    And the original collected pickup is unchanged
    And the account still has 1 collections, 1 stamps and 0 grants

  @FLOW_005
  Scenario: A repeated pickup occasion produces one provider effect
    Given the customer's order is ready for collection
    And the pickup notification has been delivered
    When the PickupOpened fact is delivered with a new event identity
    Then the new delivery has a committed consumer receipt
    And the pickup notification and its single provider acceptance are unchanged

  @FLOW_006
  Scenario: Concurrent collections retain every credit
    Given the customer has 6 orders ready for collection
    When all six pickups are collected concurrently
    Then the account eventually has 6 collections, 0 stamps and 2 grants
    And two different earned grants each create one reward

  @FLOW_007
  Scenario: Two commands cannot redeem the same reward version
    Given the customer has collected 3 orders and received a reward
    When two different commands redeem that reward version concurrently
    Then one redemption succeeds and one reports a version conflict
    And the reward is redeemed once without changing the account

  @FLOW_008
  Scenario: Customers never share their stamp balance
    Given the customer has collected 2 orders
    When another customer collects 2 orders
    Then each customer has 2 stamps and no earned grant

  @FLOW_009
  Scenario: Provider acceptance survives a lost response
    Given the next provider acceptance will lose its response
    When the customer's order becomes ready for collection
    Then the pickup notification eventually records delivery
    And the provider has accepted that notification exactly once

  @FLOW_010
  Scenario: A rejected completion keeps its result after preparation starts
    Given the customer's ticket is queued
    When completion is requested before preparation starts
    Then completion is rejected with "preparation_not_started"
    When the barista starts preparation and retries that completion command
    Then the original completion rejection is returned
    And no DrinksReady fact has been committed
    When the barista makes a new completion attempt
    Then the order becomes ready with one DrinksReady fact
