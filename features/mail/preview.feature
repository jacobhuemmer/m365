Feature: mail preview
  Scenario: reply preview shows the header and does not send
    Given the CLI is available
    When I run "mail reply msg-1 --all --body hi --preview"
    Then the command succeeds
    And stdout contains "Reply to message msg-1 (reply all)"
    And nothing was sent

  Scenario: preview with --json is a usage error
    Given the CLI is available
    When I run "mail send --to a@example.com --subject Hi --body hi --preview --json"
    Then exit code 3
    And nothing was sent
