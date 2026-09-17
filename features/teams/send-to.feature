Feature: teams send --to
  Scenario: dry-run to Ajay
    Given the CLI is available
    When I run "teams send --to Ajay --text ping --dry-run"
    Then the command succeeds
