Feature: files write
  Scenario: files help names upload cap
    Given the CLI is available
    When I run "files --help"
    Then the command succeeds
