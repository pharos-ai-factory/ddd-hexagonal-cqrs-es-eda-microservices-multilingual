@fast @menu
Feature: Publishing a menu edition
  A menu edition freezes the offers that customers can order.
  Publishing a drink does not silently rewrite an existing edition.

  Background:
    Given a published drink named "Coffee"
    And a draft menu edition in "EUR"

  @MENU_001
  Scenario: An empty edition cannot be published
    When the operator publishes the menu edition
    Then the menu command is rejected with "empty_menu"
    And the edition and its outgoing events are unchanged

  @MENU_002
  Scenario: Publishing freezes the complete offer
    Given the edition offers "C1" for 300 minor units
    When the operator publishes the menu edition
    Then one public menu publication contains "Coffee" at 300 minor units in "EUR"
    And the published offer refers to drink revision 1

  @MENU_003
  Scenario: Offer codes are unique within an edition
    Given the edition offers "C1" for 300 minor units
    When the operator adds offer "C1" for 400 minor units
    Then the menu command is rejected with "duplicate_offer_code"
    And the edition and its outgoing events are unchanged

  @MENU_004
  Scenario Outline: Drink revisions preserve Unicode names and leave existing offers unchanged
    Given the edition offers "C1" for 300 minor units
    When the drink is renamed to "<name>" and published again
    Then the edition still offers "Coffee" at drink revision 1
    And the new drink revision produces one private drink publication named "<name>"

    Examples:
      | name                                                                             |
      | Large coffee                                                                     |
      | κκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκκ |

  @MENU_005
  Scenario Outline: Published editions cannot be edited or published again
    Given the edition offers "C1" for 300 minor units
    And the menu edition is published
    When the operator attempts to "<action>" the published edition
    Then the menu command is rejected with "edition_already_published"
    And the edition and its outgoing events are unchanged

    Examples:
      | action        |
      | change price  |
      | add an offer  |
      | publish again |

  @MENU_006
  Scenario: Repeating the current price is a no-op
    Given the edition offers "C1" for 300 minor units
    When the operator changes offer "C1" to 300 minor units
    Then the menu command succeeds
    And the edition and its outgoing events are unchanged

  @MENU_007
  Scenario Outline: A draft price can change before publication, including an explicit free offer
    Given the edition offers "C1" for 300 minor units
    When the operator changes offer "C1" to <minor> minor units
    Then the menu command succeeds
    And the draft edition offers "Coffee" for <minor> minor units
    And no public menu publication is produced

    Examples:
      | minor |
      | 350   |
      | 0     |

  @MENU_008
  Scenario: An offer waits for its published drink revision
    Given the drink revision has not reached the menu directory
    When the operator adds offer "C1" for 300 minor units
    Then the menu command is rejected with "drink_revision_pending"
    And the edition and its outgoing events are unchanged
