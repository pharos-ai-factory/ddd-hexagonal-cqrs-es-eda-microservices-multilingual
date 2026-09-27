@fast @loyalty
Feature: Earning and using a reward
  LoyaltyAccount owns stamp accounting and the earned grant.
  A later command creates the independent Reward.

  @LOYALTY_001
  Scenario Outline: Every third collection earns a grant
    Given a customer has no credited collections
    When <collections> different orders are credited
    Then the account has <balance> stamps and <grants> earned grants
    And exactly <grants> private RewardEarned publications are produced
    And no Reward has been created by the credit commands

    Examples:
      | collections | balance | grants |
      | 1           | 1       | 0      |
      | 2           | 2       | 0      |
      | 3           | 0       | 1      |
      | 6           | 0       | 2      |

  @LOYALTY_002
  Scenario: The earned grant fixes the terms for later issuance
    Given a customer has no credited collections
    When 3 different orders are credited
    Then the private grant promises "one free drink" valid for 7 days
    And the recorded grant matches the published grant identity

  @LOYALTY_003
  Scenario: Issuing a reward is a separate command
    Given the customer has earned a grant
    When the grant is handled at "2026-01-01T12:00:00.000Z"
    Then one public RewardIssued publication expires at "2026-01-08T12:00:00.000Z"
    And the LoyaltyAccount is unchanged by the reward command

  @LOYALTY_004
  Scenario: Handling a grant for an existing reward is a no-op
    Given a reward was issued at "2026-01-01T12:00:00.000Z"
    When the grant is handled at "2026-01-02T12:00:00.000Z"
    Then the reward command succeeds
    And the Reward and its outgoing events are unchanged

  @LOYALTY_005
  Scenario: A reward can be redeemed just before its deadline
    Given a reward was issued at "2026-01-01T12:00:00.000Z"
    When the reward is redeemed at "2026-01-08T11:59:59.999Z"
    Then the reward is redeemed for the selected order
    And the LoyaltyAccount is unchanged by the reward command

  @LOYALTY_006
  Scenario Outline: The expiry deadline is exclusive
    Given a reward was issued at "2026-01-01T12:00:00.000Z"
    When the reward is redeemed at "<instant>"
    Then the reward command is rejected with "reward_expired"
    And the Reward and its outgoing events are unchanged

    Examples:
      | instant                  |
      | 2026-01-08T12:00:00.000Z |
      | 2026-01-09T12:00:00.000Z |

  @LOYALTY_007
  Scenario: A reward cannot be redeemed twice
    Given a reward was issued at "2026-01-01T12:00:00.000Z"
    And the reward has been redeemed
    When the reward is redeemed at "2026-01-02T12:00:00.000Z"
    Then the reward command is rejected with "reward_unavailable"
    And the Reward and its outgoing events are unchanged
