Feature: teams send
  Scenario: send help
    Given the CLI is available
    When I run "teams send --help"
    Then the command succeeds
