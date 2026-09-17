Feature: calendar files help
  Scenario: root help lists namespaces
    Given the CLI is available
    When I run "--help"
    Then the command succeeds
