Feature: help
  Scenario: top-level help
    Given the CLI is available
    When I run "--help"
    Then the command succeeds
