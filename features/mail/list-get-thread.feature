Feature: mail list get thread
  Scenario: list help names limits
    Given the CLI is available
    When I run "mail list --help"
    Then the command succeeds

  Scenario: read complete message metadata without attachment bytes
    Given the CLI is available
    When I run "mail get msg-1"
    Then the command succeeds
    And the message includes participants, received time, and read state
    And the message includes attachment metadata without attachment bytes

  Scenario: read every message in a conversation oldest-first
    Given the CLI is available
    When I run "mail thread msg-1 --bodies"
    Then the command succeeds
    And the complete conversation is ordered oldest-first with message ID tie-breakers
    And every message retains its body
