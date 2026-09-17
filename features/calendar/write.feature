Feature: calendar write
  Scenario: calendar help names dry-run
    Given the CLI is available
    When I run "calendar --help"
    Then the command succeeds
