Feature: MCP agent guard flags
  Scenario: read-only rejects opt-in
    Given a signed-in read-only MCP server
    When the client opts in to an MCP mail send
    Then the MCP run is usage
  Scenario: allowlist rejects an unlisted write
    Given a signed-in MCP server allowing only teams.send
    When the client opts in to an MCP mail send
    Then the MCP run is usage
  Scenario: allowlist permits a listed write
    Given a signed-in MCP server allowing only mail.send
    When the client opts in to an MCP mail send
    Then the MCP run succeeds
  Scenario: exact recipients reject fuzzy names
    Given a signed-in exact-recipients MCP server
    When the client opts in to an MCP teams send to Ajay
    Then the MCP run is usage
  Scenario: exact recipients permit email
    Given a signed-in exact-recipients MCP server
    When the client opts in to an MCP teams send to ajay@example.com
    Then the MCP run succeeds
