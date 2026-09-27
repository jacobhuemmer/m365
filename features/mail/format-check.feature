Feature: mail format problems block a send
  Scenario: a broken HTML mail is not sent
    Given the CLI is available
    When I run "mail send --to user@example.com --subject t --html --body <p>hi"
    Then exit code 3
    And nothing was sent

  Scenario: dry-run lists the problem
    Given the CLI is available
    When I run "mail send --to user@example.com --subject t --html --body <p>hi --dry-run"
    Then the command succeeds
    And stdout JSON "format_problems.0.rule" is "broken-html"
    And stdout JSON "format_problems.0.detail" is "unclosed <p>"
