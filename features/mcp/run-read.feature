Feature: MCP run reads
  Scenario: mail list through run
    Given a signed-in MCP server
    When the client runs MCP mail list
    Then the MCP run succeeds
  Scenario: unknown verb
    Given a signed-in MCP server
    When the client runs MCP mail nope
    Then the MCP run is usage
  Scenario: missing teams consent
    Given an MCP server without teams consent
    When the client runs MCP teams list
    Then the MCP run is auth
