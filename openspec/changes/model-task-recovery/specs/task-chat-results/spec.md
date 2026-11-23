## ADDED Requirements

### Requirement: Bind a task to a durable destination conversation

The system SHALL maintain one durable Main conversation per workspace and stable IDs for owned side conversations used by tasks. Task creation SHALL persist its originating conversation as the immutable destination. A task without an originating chat SHALL use Main. Explicit unknown or foreign chat IDs SHALL be rejected rather than redirected. Mail thread and calendar event IDs SHALL remain distinct from conversation IDs.

#### Scenario: Main fallback and side-chat routing

1. **WHEN** Alice creates one task without an originating chat and another from her Workshop side chat
2. **THEN** the first stores Alice's Main conversation ID and the second stores Workshop's ID.
3. **WHEN** Alice signs in from a second session
4. **THEN** both destinations retain the same IDs.

#### Scenario: Concurrent Main creation and foreign destination

1. **WHEN** two sessions first resolve Main at the same time
2. **THEN** they receive one conversation ID for that workspace.
3. **WHEN** Bob supplies that conversation ID to create a task
4. **THEN** creation is denied without saving a task or revealing Alice's messages.

### Requirement: Commit a terminal outcome with recoverable delivery

Each task execution SHALL commit its terminal state, immutable outcome, and pending chat delivery together. Success and failure messages SHALL use the task's saved destination; cancellation and task expiry SHALL also have persisted terminal messages. The message SHALL reference the task ID, execution, and stable result/evidence IDs. A publication failure SHALL leave the task outcome intact and delivery pending.

#### Scenario: Crash after outcome commit

1. **WHEN** a worker commits a successful outcome and pending delivery, then stops before publication while the client is closed
2. **THEN** restarting the dispatcher publishes the outcome to the saved originating chat, and reopening that chat shows it.

#### Scenario: Publication is temporarily unavailable

1. **WHEN** a task finishes but message publication repeatedly fails
2. **THEN** task detail shows its completed outcome and pending delivery, the pending intent survives restart, and the task is not re-executed or failed merely because delivery is delayed.

### Requirement: Publish once per execution and replay saved messages

Publication SHALL use a preallocated stable message identity and deduplicate by terminal outcome. Repeated dispatch or a lost publication acknowledgement SHALL yield one visible message per execution. A manual restart SHALL retain earlier messages and publish a distinct message for its new execution. Authenticated conversation reads SHALL replay ordered, paginated task messages across sessions and restarts without duplicating them in the client.

#### Scenario: Lost publication acknowledgement

1. **WHEN** publication commits but its acknowledgement is lost, and two dispatcher passes retry the same intent
2. **THEN** the destination contains one message with the original message ID and outcome.

#### Scenario: Failure followed by explicit restart

1. **WHEN** a task fails, the user starts it again, and the new execution succeeds
2. **THEN** the chat retains one failure message for the first execution and one success message for the second, both linked to the same task ID and their respective outcomes.

#### Scenario: Reopen a file result

1. **WHEN** a task publishes a registered PDF result, its original preview URL expires, and the owner reopens the chat in a second session
2. **THEN** the message retains the same file ID and obtains a fresh authorized preview for the original version's bytes.

### Requirement: Show task results in their chat and task views

The dashboard SHALL load task messages for the selected conversation, refresh them while visible and on return, and display the saved result or safe failure message. Tasks and Activity SHALL link to that same outcome and destination. A completion SHALL NOT be rerouted to whichever conversation happens to be open. Minimal side-conversation identity SHALL survive reload so the originating chat can be reopened before full Rich Threads integration.

#### Scenario: Another chat is open at completion

1. **WHEN** a task belonging to Workshop finishes while Main is open
2. **THEN** its completion is saved to Workshop, Main does not receive a duplicate, and opening the task's result navigates to Workshop's saved message.

### Requirement: Stop the selected task from the message input

The associated message input SHALL expose a visible, keyboard-accessible Stop task button identifying the selected active task. It SHALL invoke durable cancellation for that task ID and show the committed response or request error. It SHALL preserve the draft and unrelated tasks. Stopping an ordinary chat turn SHALL retain its separate lifecycle and SHALL NOT implicitly cancel delegated tasks. Pause, Resume, and eligible Start again SHALL be available in task detail.

#### Scenario: Stop one of several tasks

1. **WHEN** two tasks are active and the owner presses Stop task for the selected Workshop summary
2. **THEN** only that task is cancelled, its late worker progress is rejected, the other task continues, and the composer draft remains.

#### Scenario: Narrow-screen controls and request failure

1. **WHEN** the owner opens the task-associated composer at a narrow viewport and the cancellation request fails before commit
2. **THEN** Stop task remains reachable by keyboard, the interface shows an actionable error without falsely declaring cancellation, and retrying with the same request key is safe.
