## ADDED Requirements

### Requirement: Persist an owned task before acknowledging creation

The system SHALL derive task ownership from the authenticated workspace and persist the task ID, type, request, initial plan, destination conversation, initial execution, state, and any due time before returning success. Source documents and evidence SHALL be validated against that workspace. Task list and detail operations SHALL return persisted state and reject access by another owner. A delegated task SHALL continue independently of the creating session's lifetime.

#### Scenario: Commit before confirmation

1. **WHEN** task creation is held before database commit
2. **THEN** no success response or scheduling confirmation is emitted.
3. **WHEN** the transaction commits and the owner opens a second session
4. **THEN** the response and task list identify the same task, destination, and due time.

#### Scenario: Sign-out and owner isolation

1. **WHEN** Alice creates a task, signs out of its creating session, and Bob requests its detail, mutation, or file references
2. **THEN** Alice's saved task remains eligible, Bob receives no private task data and cannot change it, and Alice's other active session can inspect it.

### Requirement: Replay a committed mutation after response loss

Task creation and owner control mutations SHALL require a workspace-scoped idempotency key and persist the response in the same transaction as the mutation. An authorized retry with the same scope, key, and validated payload SHALL return the saved status/body without repeating the mutation. A different payload using that key SHALL return a conflict. Scope SHALL be identical across supported API prefixes. Replays SHALL recheck authentication and ownership.

#### Scenario: Response is lost after commit

1. **WHEN** task creation commits but its response is lost, and two requests repeat its key and payload
2. **THEN** both return the original saved task ID and response, with one task and one initial execution.

#### Scenario: Conflicting payload and revoked session

1. **WHEN** a caller reuses a committed key with a different due time
2. **THEN** the API returns 409 without creating or changing a task.
3. **WHEN** the original request is replayed with a revoked session
4. **THEN** it returns 401 without revealing the saved response.

### Requirement: Acquire and maintain a single current lease

The worker SHALL atomically check current execution, version, state, due time, retry eligibility, and expiry when claiming work. A winning claim SHALL create an attempt with an original start time and a unique expiring lease. Workers SHALL renew leases during long work. Every progress write, checkpoint, completion, and effect authorization SHALL verify the current running execution and unexpired lease. Losing ownership SHALL stop local execution and prevent stale durable changes.

#### Scenario: Competing processes

1. **WHEN** two separate Go workers against Docker PostgreSQL are released at a barrier against the same eligible task
2. **THEN** one enters its handler under a valid lease, one attempt is created, and one durable outcome is recorded.

#### Scenario: Renewal and stale completion

1. **WHEN** a long operation renews its lease before expiry
2. **THEN** a competing worker cannot reclaim it.
3. **WHEN** renewal later fails and another worker acquires the expired lease
4. **THEN** the old worker cannot save progress, authorize an effect, or complete the task.

### Requirement: Recover versioned checkpoints and immutable document inputs

The system SHALL save versioned execution checkpoints containing completed and remaining work, accepted input references, stable operation keys, evidence, and registered file IDs. Document reads SHALL reuse the existing document/workspace services and the task's permanent workspace. Recovery SHALL preserve the task ID, retain the interrupted attempt and its original start time, and create a new attempt from the last compatible checkpoint. An unsupported handler or checkpoint version SHALL fail visibly before effects.

#### Scenario: Resume the original file on another host

1. **WHEN** a worker checkpoints a PDF by `fileId` and stops, the owner edits and registers that path again, and another owner stores different bytes at the same path
2. **THEN** recovery on a separate host after lease expiry reads the first registered version's exact bytes and retains the interrupted attempt.
3. **WHEN** the old worker submits a late checkpoint
4. **THEN** it changes neither progress nor the recovered outcome.

#### Scenario: Incompatible saved state

1. **WHEN** a claimed task contains a checkpoint or handler version unsupported by the installed worker
2. **THEN** the task records a safe incompatibility failure and chat delivery intent, with zero handler effects and no silent restart from empty state.

### Requirement: Reconcile operations before repeating effects

Each effectful step SHALL have a stable operation identity and normalized input fingerprint. Recovery and manual restart SHALL reuse confirmed results and reconcile receipts created after the last checkpoint. A mismatched fingerprint SHALL be rejected. A dispatch with an uncertain result SHALL NOT be automatically repeated. Review decisions SHALL be consumed idempotently and revalidated against workspace, task eligibility, payload, connection generation, and expiry before dispatch. Cancellation SHALL invalidate pending reviews and prevent new authorization; it SHALL NOT claim to retract an already-dispatched provider effect.

#### Scenario: Effect completed after the checkpoint

1. **WHEN** an operation records a result file or confirmed receipt and the worker stops before its next checkpoint
2. **THEN** recovery adopts that result under the same operation identity and does not repeat the completed effect.

#### Scenario: Cancellation wins before dispatch authorization

1. **WHEN** a worker is held after claim but before effect authorization and the owner cancels the task
2. **THEN** releasing the worker produces zero external effects and preserves cancelled state.

#### Scenario: Provider acceptance is uncertain

1. **WHEN** a controlled provider accepts a send but its response is lost
2. **THEN** the receipt remains uncertain and the task reports that uncertainty in a failure outcome.
3. **WHEN** the service restarts or the owner requests another execution
4. **THEN** the same operation is reconciled without another automatic send.

### Requirement: Bound safe automatic retries

Workers SHALL automatically retry transient failures only when repetition is safe. The first such failure SHALL start a persisted two-minute recovery budget, using backoff and a persisted next-attempt time. Active recovery and retry backoff, including worker downtime, SHALL consume that budget. Scheduled waits, explicit pauses, and input/review waits SHALL preserve its remaining duration without consuming or resetting it. Exhaustion SHALL invalidate the lease, save a failure message, and enqueue chat delivery. Initial healthy execution SHALL NOT be limited to two minutes solely by this retry policy.

#### Scenario: Restart during retry backoff

1. **WHEN** a safe operation repeatedly fails, the worker restarts during backoff, and the controlled clock reaches the original recovery deadline
2. **THEN** no additional retry begins, the task becomes failed, and its saved message explains that two minutes of retries were exhausted and that the user can start it again.

#### Scenario: Wait preserves remaining budget

1. **WHEN** 30 seconds of recovery have elapsed and the task is paused for five minutes
2. **THEN** its remaining recovery budget stays at 90 seconds.
3. **WHEN** the owner explicitly resumes it and safe failures continue
4. **THEN** it fails after those remaining 90 seconds rather than receiving a new two minutes.

#### Scenario: A retry stalls

1. **WHEN** a retry handler remains blocked past the persisted recovery deadline
2. **THEN** the task becomes failed through deadline enforcement and a late handler result cannot overwrite that outcome.

### Requirement: Pause and stop invalidate pending execution and answers

The owner SHALL be able to pause or cancel an eligible active task. Pause SHALL retain committed progress and accepted answers, release the execution slot, invalidate its lease and unresolved input request IDs, and require explicit resume. New answers submitted while paused or cancelled, or to closed request IDs, SHALL be rejected without storing the answer or queuing work. Resume SHALL reissue still-needed input under a new request ID. Cancellation SHALL end eligibility for that execution and close pending reviews. Races with completion SHALL resolve atomically to one terminal outcome.

#### Scenario: Late answer after pause and resume

1. **WHEN** the owner pauses a task awaiting input and an answer to its old request arrives
2. **THEN** the answer is rejected without storage and the task remains paused, including after restart.
3. **WHEN** the owner resumes and then answers the old request ID again
4. **THEN** the old answer is still rejected; only the new current request can be answered.

#### Scenario: Stop races completion

1. **WHEN** cancellation and completion race on a running task
2. **THEN** exactly one terminal transition wins and produces an outcome; a stale worker cannot resurrect cancelled work.

### Requirement: Wait without blocking unrelated tasks

A task needing input or review SHALL atomically save its checkpoint and pending request/review before releasing its lease and execution slot. Valid answers and decisions SHALL advance it once. A paused task SHALL NOT resume automatically when a review decision arrives; explicit resume SHALL revalidate the saved decision. Denial, expiry, or invalid connection identity SHALL NOT permit an external action.

#### Scenario: Another task finishes during review

1. **WHEN** one task waits for review and another eligible task is queued
2. **THEN** the second task finishes while the first retains its pending review and saved progress.
3. **WHEN** the same valid review decision is delivered twice
4. **THEN** one continuation and one recorded effect outcome result.

#### Scenario: Approval expires while paused

1. **WHEN** a task is paused, a review decision arrives, and that review expires before explicit resume
2. **THEN** the task remains paused until resumed and requires fresh review before any send.

### Requirement: Distinguish task restart from lease recovery

Lease expiry SHALL recover an eligible task automatically. Optional task expiry SHALL end any nonterminal execution, including paused or waiting states, and publish an expired outcome. The owner SHALL be able to restart a failed or expired task explicitly under the same task ID, creating a new execution and fresh recovery budget while preserving attempts, outcomes, accepted checkpoint context, and operation receipts. Restart of an expired task SHALL require a new valid expiry or explicit removal of expiry. Successful and cancelled executions SHALL NOT be resumed by that endpoint.

#### Scenario: Restart an expired task

1. **WHEN** a task expires while paused and the owner requests restart with its expiry cleared
2. **THEN** a new execution becomes eligible under the same task ID with the previous expired outcome retained.
3. **WHEN** the restart response is lost and its request key is replayed
4. **THEN** no additional execution is created.

### Requirement: Persist and enforce reminder scheduling

A scheduled task SHALL save an absolute due instant and an unambiguous event-zone context before acknowledging creation. It SHALL remain idle before due and execute once after downtime when overdue, unless task expiry has passed. Its saved due time SHALL remain fixed when the source event changes in this version. Retry backoff SHALL NOT overwrite the promised due time. Reminder success SHALL produce a persisted result for its destination chat.

#### Scenario: Workshop reminder survives restart

1. **WHEN** a reminder is saved for 30 minutes before a 10:00 workshop and the worker restarts before 09:30
2. **THEN** the response and second-session task list contain the same task ID and 09:30 in the event's zone, with zero claims before that instant and one successful execution when due.

#### Scenario: Catch up or expire

1. **WHEN** workers return after a reminder's due time while its task remains unexpired
2. **THEN** it becomes eligible for one catch-up execution.
3. **WHEN** the same test instead returns after its task expiry
4. **THEN** it records an expired outcome without running the reminder handler and exposes explicit restart.

#### Scenario: Event changes after scheduling

1. **WHEN** the calendar event time changes after its reminder was committed
2. **THEN** the reminder retains its original saved due time and task identity.

### Requirement: Inspect durable task progress and outcomes

Task detail SHALL expose current execution state, due time, progress, attempts, outstanding input or review, safe failure information, and stable outcome/file references. Task lists SHALL support owner-scoped pagination and conversation/state filters. Tasks and Activity SHALL derive their display from persisted records and link to the same outcome shown in chat. Oversized or invalid task/checkpoint payloads SHALL be rejected before mutations.

#### Scenario: Reopen task detail

1. **WHEN** the API restarts after one failed execution and a successful explicit restart
2. **THEN** task detail retains both executions and their attempts, identifies the current successful outcome, and exposes its stable result references.
