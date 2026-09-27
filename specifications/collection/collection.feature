@fast @collection
Feature: Collecting ready drinks
  A Pickup owns its collection code and single-use status.
  Loyalty is informed only after a successful collection.

  Background:
    Given drinks are ready for an order

  @COLLECT_001
  Scenario: Opening a pickup provides a stable collection code
    When Collection handles the ready drinks
    Then the pickup is ready with a six-character collection code
    And one PickupOpened publication contains that code, order and customer

  @COLLECT_002
  Scenario: A wrong code cannot collect the order
    Given the pickup has been opened
    When the customer presents the wrong collection code
    Then Collection rejects the command with "incorrect_collection_code"
    And the pickup and its outgoing events are unchanged

  @COLLECT_003
  Scenario: The right code completes collection
    Given the pickup has been opened
    When the customer presents the correct collection code
    Then the pickup is collected
    And one OrderCollected publication identifies the order and customer

  @COLLECT_004
  Scenario: A collected pickup cannot be collected again
    Given the pickup has been collected
    When the customer presents the correct collection code
    Then Collection rejects the command with "pickup_already_collected"
    And the pickup and its outgoing events are unchanged

  @COLLECT_005
  Scenario: Receiving ready drinks again cannot reopen a collected pickup
    Given the pickup has been collected
    When Collection handles the ready drinks
    Then the pickup is collected
    And the pickup and its outgoing events are unchanged
