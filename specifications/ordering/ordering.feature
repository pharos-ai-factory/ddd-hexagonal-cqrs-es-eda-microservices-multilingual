@fast @ordering
Feature: Placing an order
  An Order owns its lines and their combined quantity.
  Customers order the terms of one published edition.

  Background:
    Given Ordering knows a published edition offering "Coffee" for 300 minor units
    And a customer has a draft order for that edition

  @ORDER_001
  Scenario: An empty order cannot be placed
    When the customer places the order
    Then the order command is rejected with "empty_order"
    And the order and its outgoing events are unchanged

  @ORDER_002
  Scenario Outline: The inclusive drink limits can be ordered
    Given the order has a line containing <quantity> drinks
    When the customer places the order
    Then one public order publication contains <quantity> drinks at the frozen price

    Examples:
      | quantity |
      | 1        |
      | 5        |

  @ORDER_003
  Scenario: A second line cannot take the order above five drinks
    Given the order has a line containing 5 drinks
    When the customer adds a second line containing 1 drink
    Then the order command is rejected with "too_many_drinks"
    And the order and its outgoing events are unchanged

  @ORDER_004
  Scenario: Editing either line respects the combined limit
    Given the order has a line containing 2 drinks
    And a second line contains 3 drinks
    When the customer changes the first line to 3 drinks
    Then the order command is rejected with "too_many_drinks"
    And the order and its outgoing events are unchanged

  @ORDER_005
  Scenario: A quantity edit preserves line identity
    Given the order has a line containing 1 drink
    When the customer changes the first line to 2 drinks
    Then the first line keeps its identity and contains 2 drinks

  @ORDER_006
  Scenario: Repeating the current quantity is a no-op
    Given the order has a line containing 2 drinks
    When the customer changes the first line to 2 drinks
    Then the order command succeeds
    And the order and its outgoing events are unchanged

  @ORDER_007
  Scenario Outline: Placement freezes all lines
    Given the order has a line containing 1 drink
    And the order is placed
    When the customer attempts to "<action>" the placed order
    Then the order command is rejected with "order_already_placed"
    And the order and its outgoing events are unchanged

    Examples:
      | action          |
      | change quantity |
      | add a line      |

  @ORDER_008
  Scenario: A line cannot use another edition
    When the customer selects an offer from a different edition
    Then the order command is rejected with "incorrect_edition"
    And the order and its outgoing events are unchanged

  @ORDER_009
  Scenario: An unknown offer cannot be added
    When the customer selects offer "UNKNOWN"
    Then the order command is rejected with "offer_not_found"
    And the order and its outgoing events are unchanged
