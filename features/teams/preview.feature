Feature: teams preview
  Scenario: teams preview shows the chat and does not send
    Given the CLI is available
    When I run "teams send chat-1 --text hi --preview"
    Then the command succeeds
    And stdout contains "Chat: chat-1"
    And nothing was sent

  Scenario: preview with --json is a usage error
    Given the CLI is available
    When I run "teams send chat-1 --text hi --preview --json"
    Then exit code 3
