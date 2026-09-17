Feature: calendar list get
  Scenario: calendar help
    Given the CLI is available
    When I run "calendar --help"
    Then the command succeeds
