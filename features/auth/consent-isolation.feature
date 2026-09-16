Feature: namespace consent
  Scenario: mail works without teams in help
    Given the CLI is available
    When I run "mail list --help"
    Then the command succeeds
