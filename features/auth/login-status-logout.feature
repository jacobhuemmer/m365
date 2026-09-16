Feature: auth login status logout
  Scenario: status while signed out
    Given the CLI is available
    When I run "auth status"
    Then the command succeeds
  Scenario: login then status
    Given the CLI is available
    When I run "auth login"
    Then the command succeeds
