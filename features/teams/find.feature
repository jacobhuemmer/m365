Feature: teams find
  Scenario: find Ajay
    Given the CLI is available
    When I run "teams find Ajay"
    Then the command succeeds
