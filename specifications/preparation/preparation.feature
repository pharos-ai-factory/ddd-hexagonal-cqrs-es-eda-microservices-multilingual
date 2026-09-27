@fast @preparation
Feature: Preparing the ordered drinks
  A preparation ticket keeps the customer's frozen instructions.
  Collection is informed only when preparation completes.

  Background:
    Given an order for 2 "Coffee" drinks has been placed

  @PREP_001
  Scenario: Accepting an order queues its instructions
    When Preparation accepts the placed order
    Then the ticket is "queued" with instructions "2 × Coffee"
    And Preparation produces no outgoing events

  @PREP_002
  Scenario: A queued ticket cannot be completed
    Given Preparation has accepted the placed order
    When the barista completes the ticket
    Then Preparation rejects the command with "preparation_not_started"
    And the ticket and its outgoing events are unchanged

  @PREP_003
  Scenario: Starting preparation does not announce ready drinks
    Given Preparation has accepted the placed order
    When the barista starts the ticket
    Then the ticket is "preparing" with instructions "2 × Coffee"
    And Preparation produces no outgoing events

  @PREP_004
  Scenario: Completion announces the order and customer once
    Given the ticket is being prepared
    When the barista completes the ticket
    Then the ticket is "ready" with instructions "2 × Coffee"
    And one DrinksReady publication identifies the order and customer

  @PREP_005
  Scenario: Starting twice is rejected
    Given the ticket is being prepared
    When the barista starts the ticket
    Then Preparation rejects the command with "ticket_not_queued"
    And the ticket and its outgoing events are unchanged

  @PREP_006
  Scenario: Accepting the same order cannot reset a ready ticket
    Given the ticket has been completed
    When Preparation accepts the placed order
    Then the ticket is "ready" with instructions "2 × Coffee"
    And the ticket and its outgoing events are unchanged
