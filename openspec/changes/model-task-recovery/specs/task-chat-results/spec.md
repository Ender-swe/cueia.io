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

### Requirement: Read the committed outcome as the chat message

A terminal transition SHALL commit the execution state and exactly one immutable outcome together. The outcome SHALL identify the task, execution, final status, safe message, and any stable result or evidence IDs. Its ID SHALL be the stable chat message ID. Authenticated reads of the task's saved conversation, Tasks, and Activity SHALL expose that same record immediately after commit for success, failure, cancellation, and expiry.

#### Scenario: Worker crashes at commit

1. **WHEN** the worker dies before its terminal transaction commits
2. **THEN** neither a terminal state nor a chat message is visible, and the task is eligible for normal lease recovery.
3. **WHEN** the transaction instead commits but the acknowledgement is lost and the worker stops while the client is closed
4. **THEN** a second session can read the outcome immediately with workers offline. Chat, Tasks, and Activity return the same ID. Restarting the worker adds no second outcome.

#### Scenario: Stop wins the completion race

1. **WHEN** the owner stops a task after its replacement worker claims it and both that worker and its predecessor later submit completion
2. **THEN** the originating chat contains one cancellation for the execution, and both late completions are rejected.

### Requirement: Replay outcomes across executions and sessions

Conversation reads SHALL return the owner's task outcomes for the saved conversation in stable, paginated order. Repeated fetches SHALL identify each outcome by its original ID. An explicit restart SHALL retain the earlier outcome and allow the new execution to record its own message under the same task ID.

#### Scenario: Repeat a read on reopening

1. **WHEN** the owner reopens a side chat, pages through results with the same timestamp, and refreshes the feed
2. **THEN** every result appears once in stable order with its original ID and the same referenced files.

#### Scenario: Failure followed by explicit restart

1. **WHEN** a task fails, the user starts it again, and the new execution succeeds
2. **THEN** the chat retains one failure message for the first execution and one success message for the second, both linked to the same task ID and their respective outcomes.

#### Scenario: Reopen a file result

1. **WHEN** a task finishes with a registered PDF reference, its original preview URL expires, and the owner reopens the chat in a second session
2. **THEN** the message retains the same file ID and obtains a fresh authorized preview for the original version's bytes.

### Requirement: Show task results in their chat and task views

The dashboard SHALL load task messages for the selected conversation, refresh them while visible and on return, and display the saved result or safe failure message. Tasks and Activity SHALL link to that same outcome and destination. A completion SHALL NOT be rerouted to whichever conversation happens to be open. Minimal side-conversation identity SHALL survive reload so the originating chat can be reopened before full Rich Threads integration.

#### Scenario: Another chat is open at completion

1. **WHEN** a task belonging to Workshop finishes while Main is open
2. **THEN** its completion is saved to Workshop, Main does not receive a duplicate, and opening the task's result navigates to Workshop's saved message.

### Requirement: Stop the selected task from the message input

The associated message input SHALL expose a visible, keyboard-accessible Stop task button identifying the selected active task. It SHALL invoke durable cancellation for that task ID and show the committed response or request error. It SHALL preserve the draft and unrelated tasks. Stopping an ordinary chat turn SHALL retain its separate lifecycle and SHALL NOT implicitly cancel delegated tasks. Pause, Resume, and eligible Start again SHALL be available in task detail.

Owner controls SHALL bind to the selected execution number and version and retain their original idempotency key and payload on ambiguous network retries. A conflict SHALL refresh task state without automatically retargeting the control at a newer execution. A replayed response for an earlier execution SHALL NOT mark the current execution stopped.

#### Scenario: Stop one of several tasks

1. **WHEN** two tasks are active and the owner presses Stop task for the selected Workshop summary
2. **THEN** only that task is cancelled, its late worker progress is rejected, the other task continues, and the composer draft remains.

#### Scenario: Narrow-screen controls and request failure

1. **WHEN** the owner opens the task-associated composer at a narrow viewport and the cancellation request fails before commit
2. **THEN** Stop task remains reachable by keyboard, the interface shows an actionable error without falsely declaring cancellation, and retrying with the same request key is safe.

#### Scenario: Control response belongs to an earlier execution

1. **WHEN** a delayed control for execution 1 conflicts after execution 2 starts, or an earlier saved control response is replayed after that restart
2. **THEN** the interface refreshes the current task state, leaves execution 2 active, preserves the draft, and requires an explicit new action to control execution 2.
