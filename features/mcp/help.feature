Feature: MCP help
  Scenario: help without a session
    Given an MCP server
    When the client calls MCP help for calendar
    Then MCP help names calendar verbs
