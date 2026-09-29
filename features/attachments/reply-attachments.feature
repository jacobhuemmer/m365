Feature: Email reply attachment delivery
  Scenario: Reply sends a local file with the authored comment
    Given the CLI is available
    When I exercise reply attachments sender
    Then the attachment reply is sent once
  Scenario: Reply all sends a local file with the authored comment
    Given the CLI is available
    When I exercise reply attachments all
    Then the attachment reply is sent once
  Scenario: Reply without files retains the comment-only request
    Given the CLI is available
    When I exercise reply attachments sender without files
    Then the attachment reply is sent once
  Scenario: Reply all without files retains the comment-only request
    Given the CLI is available
    When I exercise reply attachments all without files
    Then the attachment reply is sent once
  Scenario: A later missing file prevents reply
    Given the CLI is available
    When I exercise reply attachments sender missing
    Then the attachment reply fails without fallback usage
  Scenario: A later missing file prevents reply all
    Given the CLI is available
    When I exercise reply attachments all missing
    Then the attachment reply fails without fallback usage
  Scenario: Rejected reply does not report success or fall back
    Given the CLI is available
    When I exercise reply attachments sender rejected
    Then the attachment reply fails without fallback service
  Scenario: Rejected reply all does not report success or fall back
    Given the CLI is available
    When I exercise reply attachments all rejected
    Then the attachment reply fails without fallback service
  Scenario: Reply preview exposes file metadata and sends nothing
    Given the CLI is available
    When I exercise reply attachments sender preview
    Then the attachment reply preview has metadata only
  Scenario: Reply all preview exposes file metadata and sends nothing
    Given the CLI is available
    When I exercise reply attachments all preview
    Then the attachment reply preview has metadata only
