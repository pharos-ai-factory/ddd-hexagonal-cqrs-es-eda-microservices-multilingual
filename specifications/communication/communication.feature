@fast @communication
Feature: Delivering customer notifications
  Requesting a notification records a private delivery hand-off.
  Provider acceptance is recorded by a later command.

  @COMMS_001
  Scenario: Ready drinks request a pickup notification
    When Communication handles an opened pickup with code "ABC123"
    Then the requested notification tells the customer "Collect your order using code ABC123"
    And one private NotificationRequested publication identifies the notification
    And no provider call has been made

  @COMMS_002
  Scenario: An issued reward requests a reward notification
    When Communication handles a reward valid until "2026-01-08T12:00:00.000Z"
    Then the requested notification tells the customer "one free drink; valid until 2026-01-08T12:00:00.000Z"
    And one private NotificationRequested publication identifies the notification

  @COMMS_003
  Scenario: A repeated pickup occasion cannot replace its notification
    Given a pickup notification has been requested
    When Communication handles an opened pickup with code "ABC123"
    Then the Notification and its outgoing events are unchanged

  @COMMS_004
  Scenario: Successful delivery records the provider receipt
    Given a pickup notification has been requested
    When the delivery worker handles the request
    Then the notification is sent with provider receipt "accepted-1"
    And the provider has received exactly 1 call
    And delivery produces no outgoing business events

  @COMMS_005
  Scenario: Provider failure leaves the request available for retry
    Given a pickup notification has been requested
    And the provider is unavailable
    When the delivery worker handles the request
    Then the provider failure is available to the retry mechanism
    And the notification is still requested at its original version
    And no delivery command has committed

  @COMMS_006
  Scenario: Redelivery of a sent notification does not call the provider again
    Given a pickup notification has been delivered
    When the delivery worker handles the request
    Then the provider has received exactly 1 call
    And the Notification and its outgoing events are unchanged
