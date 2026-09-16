Feature: mail send reply
  Scenario: send help
    Given the CLI is available
    When I run "mail send --help"
    Then the command succeeds
