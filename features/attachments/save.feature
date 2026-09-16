Feature: save attachment
  Scenario: mail help
    Given the CLI is available
    When I run "mail --help"
    Then the command succeeds
