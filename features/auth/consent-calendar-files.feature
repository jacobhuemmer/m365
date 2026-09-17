Feature: calendar files consent
  Scenario: help still works signed out
    Given the CLI is available
    When I run "auth status"
    Then the command succeeds
