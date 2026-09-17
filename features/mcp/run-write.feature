Feature: MCP write opt-in
  Scenario: send without opt-in is dry-run
    Given a signed-in MCP server
    When the client runs MCP mail send without opt-in
    Then the MCP run is dry-run
  Scenario: calendar create without opt-in
    Given a signed-in MCP server
    When the client runs MCP calendar create without opt-in
    Then the MCP run is dry-run
