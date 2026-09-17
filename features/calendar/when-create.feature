Feature: calendar when create
  Scenario: calendar help names when
    Given the CLI is available
    When I run "calendar --help"
    Then the command succeeds
