Feature: teams format problems block a send
  Scenario: a broken HTML chat message is not sent
    Given the CLI is available
    When I run "teams send chat-1 --html --text <p>hi"
    Then exit code 3
    And nothing was sent

  Scenario: dry-run lists the problem
    Given the CLI is available
    When I run "teams send chat-1 --html --text <p>hi --dry-run"
    Then the command succeeds
    And stdout JSON "format_problems.0.rule" is "broken-html"
