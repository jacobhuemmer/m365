Feature: files list get
  Scenario: files help
    Given the CLI is available
    When I run "files --help"
    Then the command succeeds
