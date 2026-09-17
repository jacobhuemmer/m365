Feature: calendar free
  Scenario: calendar help names free
    Given the CLI is available
    When I run "calendar --help"
    Then the command succeeds
