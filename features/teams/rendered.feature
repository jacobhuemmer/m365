Feature: teams dry-run shows the rendered body
  Scenario: teams send dry-run returns the delivered HTML
    Given the CLI is available
    When I run "teams send chat-1 --text b --dry-run"
    Then the command succeeds
    And stdout JSON "rendered.content_type" is "html"
    And stdout JSON "rendered.content" is "<p>b</p>"
    And stdout JSON "format_problems" is []
