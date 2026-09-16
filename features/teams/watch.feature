Feature: teams watch
  Scenario: watch help via teams help
    Given the CLI is available
    When I run "teams --help"
    Then the command succeeds
