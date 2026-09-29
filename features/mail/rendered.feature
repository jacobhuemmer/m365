Feature: mail dry-run shows the rendered body
  Scenario: mail send dry-run returns the delivered HTML
    Given the CLI is available
    When I run "mail send --to user@example.com --subject t --body b --dry-run"
    Then the command succeeds
    And stdout JSON "rendered.content_type" is "html"
    And stdout JSON "rendered.content" is "<p>b</p>"
    And stdout JSON "format_problems" is []

  Scenario: mail reply dry-run returns the delivered comment
    Given the CLI is available
    When I run "mail reply msg-1 --body b --dry-run"
    Then the command succeeds
    And stdout JSON "rendered.content" is "<p>b</p>"
    And stdout JSON "format_problems" is []
