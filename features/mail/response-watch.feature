Feature: resumable Outlook response triage
  Mail changes are delivered once per changed conversation without exposing message bodies.
  Experimental response classification requires explicit enablement and never authorizes a reply.

  Scenario: First poll establishes a quiet baseline
    Given a signed-in fake mailbox for mail watch
    When I poll mail changes
    Then the mail watch command succeeds
    And no mail change events are emitted
    When I poll mail changes again
    Then the mail watch command succeeds
    And no mail change events are emitted

  Scenario: Existing conversations are emitted only when requested
    Given a signed-in fake mailbox for mail watch
    When I poll mail changes including existing messages
    Then the mail watch command succeeds
    And one body-free mail change event is emitted per changed conversation
    When I poll mail changes again
    Then the mail watch command succeeds
    And no mail change events are emitted

  Scenario: Response classification is disabled by default
    Given a signed-in fake mailbox for mail watch
    When I classify existing mail for "user@example.com"
    Then the mail watch command is rejected as usage
    And no mail watch events are emitted
    And the mail watch checkpoint is unchanged

  Scenario: Enabled classification emits privacy-bounded response assessments
    Given a signed-in fake mailbox with response classification enabled
    When I classify existing mail for "user@example.com"
    Then the mail watch command succeeds
    And one privacy-safe response classification event is emitted per changed conversation
    And response actionability follows the configured threshold
    When I classify mail for "user@example.com" again
    Then the mail watch command succeeds
    And no mail change events are emitted

  Scenario: A waiting-on-target assessment below the threshold is not actionable
    Given a signed-in fake mailbox with response classification enabled above the fake probability
    When I classify existing mail for "user@example.com"
    Then the mail watch command succeeds
    And every response classification event is not actionable
