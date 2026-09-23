Feature: MCP lookup recipes
  Scenario: help names recipe topics
    Given an MCP server
    When the client calls MCP help with no topic
    Then MCP help names recipe topics
  Scenario: mail-search recipe
    Given an MCP server
    When the client calls MCP help topic mail-search
    Then MCP help includes from:ajay
  Scenario: prompts stay six and tools stay three
    Given an MCP server
    When the client lists MCP tools
    Then the MCP catalog is compact
    When the client lists MCP prompts
    Then the MCP recipes are six
