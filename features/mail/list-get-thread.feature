Feature: mail list get thread
  Scenario: list help names limits
    Given the CLI is available
    When I run "mail list --help"
    Then the command succeeds
