Feature: attach on send
  Scenario: send help names caps
    Given the CLI is available
    When I run "mail send --help"
    Then the command succeeds
