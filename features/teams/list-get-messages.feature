Feature: teams list
  Scenario: list help
    Given the CLI is available
    When I run "teams list --help"
    Then the command succeeds

  Scenario: messages help
    Given the CLI is available
    When I run "teams messages --help"
    Then the command succeeds
