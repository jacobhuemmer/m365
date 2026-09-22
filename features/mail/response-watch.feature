Feature: resumable Outlook mail change feed
  Mail changes are delivered once per changed conversation without exposing message bodies,
  and a durable cursor lets later one-shot polls resume from the completed delta round.

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
