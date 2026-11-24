## ADDED Requirements

### Requirement: Persist an owned task before acknowledging creation

The system SHALL derive task ownership from the authenticated workspace and persist the task ID, type, request, initial plan, destination conversation, initial execution, state, and any due time or expiry before returning success. This release SHALL accept one-shot reminders and read-only tasks over registered PDFs. Unsupported task types, including requests to send mail or change an external resource, SHALL be rejected before creating work. Source documents and evidence SHALL be validated against that workspace. Task list and detail operations SHALL return persisted state and reject access by another owner. A delegated task SHALL continue independently of the creating session's lifetime.

#### Scenario: Commit before confirmation

1. **WHEN** task creation is held before database commit
2. **THEN** no success response or scheduling confirmation is emitted.
3. **WHEN** the transaction commits and the owner opens a second session
4. **THEN** the response and task list identify the same task, destination, and due time.

#### Scenario: Sign-out and owner isolation

1. **WHEN** Alice creates a task, signs out of its creating session, and Bob requests its detail, mutation, or file references
2. **THEN** Alice's saved task remains eligible, Bob receives no private task data and cannot change it, and Alice's other active session can inspect it.

#### Scenario: Unsupported task type

1. **WHEN** an owner requests a task whose type would send a provider message
2. **THEN** creation returns a validation error, leaves the task list unchanged, and issues zero provider calls.

### Requirement: Replay a committed mutation after response loss

Task creation and owner control mutations SHALL require a workspace-scoped idempotency key and persist the response in the same transaction as the mutation. An authorized retry with the same scope, key, and validated payload SHALL return the saved status/body without repeating the mutation, even when the execution or version has since changed. A different payload using that key SHALL return a conflict. Scope SHALL be identical across supported API prefixes. Replays SHALL recheck authentication and ownership before returning stored data, and SHALL resolve committed responses before evaluating current execution preconditions.

#### Scenario: Response is lost after commit

1. **WHEN** task creation commits but its response is lost, and two requests repeat its key and payload
2. **THEN** both return the original saved task ID and response, with one task and one initial execution.

#### Scenario: Conflicting payload and revoked session

1. **WHEN** a caller reuses a committed key with a different due time
2. **THEN** the API returns 409 without creating or changing a task.
3. **WHEN** the original request is replayed with a revoked session
4. **THEN** it returns 401 without revealing the saved response.

### Requirement: Bind owner controls to the intended execution

Pause, resume, cancel, restart, and answer mutations SHALL require `expectedRunNumber` and `expectedVersion`. Except for replay of a committed response, both values SHALL match the current execution atomically or the request SHALL return 409 without applying its requested change. Lifecycle transitions, accepted answers, and changes to actionable input requests SHALL increment the execution's control version; routine heartbeat renewal and progress-only checkpoint updates SHALL NOT. Task reads and mutation responses SHALL expose the execution number and control version.

#### Scenario: Delayed control reaches a restarted task

1. **WHEN** a cancellation for execution 1 is held before its first server submission, execution 1 fails, and execution 2 reaches the same numeric version
2. **THEN** releasing the old cancellation returns 409 and leaves execution 2 unchanged, even with a previously unused idempotency key.

#### Scenario: Heartbeats do not invalidate a control

1. **WHEN** the owner reads the current execution and version, and the worker renews its lease and saves progress without a lifecycle transition
2. **THEN** a control submitted with that pair remains valid.

#### Scenario: Replay after execution changes

1. **WHEN** an authorized owner replays a committed control's original key and payload after the task has entered another execution
2. **THEN** the original response is returned without applying the control to the new execution.

### Requirement: Acquire and maintain a single current lease

The worker SHALL atomically check the current execution, state, version, due time, backoff, and task expiry when claiming work. A winning claim SHALL create one attempt with its original start time and a unique renewable lease. Every heartbeat, checkpoint, progress update, wait, and completion SHALL require the current running execution and unexpired lease. A worker that loses the lease SHALL cancel its handler context. A handler still running under the old lease SHALL be unable to change durable state.

#### Scenario: Competing processes

1. **WHEN** two separate Go workers against Docker PostgreSQL are released at a barrier against the same eligible task
2. **THEN** one enters its handler under a valid lease, one attempt is created, and one durable outcome is recorded.

#### Scenario: Renewal and stale completion

1. **WHEN** a long operation renews its lease before expiry
2. **THEN** a competing worker cannot reclaim it.
3. **WHEN** renewal later fails and another worker acquires the expired lease
4. **THEN** the old worker cannot save progress, enter a wait, or complete the task.

### Requirement: Enforce optional task expiry

The system SHALL check an execution's optional expiry using database time sampled after acquiring its locks. At or past expiry, a nonterminal execution SHALL become expired with one outcome, including when paused or waiting for input. This check SHALL apply before accepting an answer, applying a control, claiming, renewing, checkpointing, or completing work. An elapsed task expiry SHALL take precedence over an attempt timeout. Idle tasks SHALL also be checked by the worker.

#### Scenario: A completion crosses task expiry

1. **WHEN** expiry polling is held and a completion waits for a lock until the task expiry instant
2. **THEN** the execution records an expired outcome, rejects the late completion, and remains expired when polling resumes.

#### Scenario: Expiry while waiting

1. **WHEN** a paused or waiting task expires while its owner is offline
2. **THEN** the worker records one expired outcome and closes outstanding input requests.

### Requirement: Recover versioned checkpoints and immutable document inputs

The system SHALL save versioned execution checkpoints containing completed and remaining work, accepted input references, evidence, and registered file IDs. Document reads SHALL reuse the existing document/workspace services and the task's permanent workspace. Recovery SHALL preserve the task ID, retain the interrupted attempt and its original start time, and create a new attempt from the last compatible checkpoint. An unsupported handler or checkpoint version SHALL fail visibly before invoking the handler.

#### Scenario: Resume the original file on another host

1. **WHEN** a worker checkpoints a PDF by `fileId` and stops, the owner edits and registers that path again, and another owner stores different bytes at the same path
2. **THEN** a replacement worker on another host uses the checkpointed file ID to read the first version's exact bytes from workspace storage. It finishes under the original task ID and retains the interrupted attempt.
3. **WHEN** the old worker submits a late checkpoint
4. **THEN** it changes neither progress nor the recovered outcome.

#### Scenario: Incompatible saved state

1. **WHEN** a claimed task contains a checkpoint or handler version unsupported by the installed worker
2. **THEN** the task records one incompatibility failure readable in its chat. The worker preserves the saved state and makes zero handler calls.

### Requirement: Bound automatic retries per execution

The system SHALL allow two automatic retries for safe work in each execution. Closing an attempt on a transient error, timeout, or unexpected lease loss SHALL increment a persisted failure count exactly once. The first two failures SHALL queue work with increasing, persisted backoff. The third SHALL commit one failed outcome. Each attempt, including the first, SHALL time out two minutes after claim even if heartbeats continue. Deadline checks SHALL reject late writes and cap lease renewal at that attempt's deadline.

Waiting for input, pausing, and a successful graceful release of a valid lease SHALL preserve the count. Downtime SHALL consume at most the attempt that was running when its worker stopped. A permanent error or a caught handler panic SHALL save a safe terminal failure immediately when the execution is still eligible.

#### Scenario: Recovery after a long outage

1. **WHEN** a worker dies after its first checkpoint and workers return an hour later to an unexpired task
2. **THEN** one interrupted attempt is recorded, elapsed backoff permits a new claim, and the new worker resumes the checkpoint.

#### Scenario: Failures survive process restarts

1. **WHEN** the first attempt fails transiently, the API and worker restart during backoff, and the next worker dies before returning a result
2. **THEN** the persisted count and next eligible time still govern the retry, and discovering the same expired lease twice counts it once.
3. **WHEN** the third attempt also fails
4. **THEN** the task has one failure message in its chat and further polling starts no fourth attempt.

#### Scenario: A handler keeps heartbeating while blocked

1. **WHEN** a stalled attempt renews its lease up to the two-minute deadline and then tries to save a result
2. **THEN** the result is rejected, the worker records one timed-out attempt, and the same retry limit applies.

#### Scenario: Waiting and graceful shutdown preserve remaining retries

1. **WHEN** a task with one prior failure pauses, resumes, waits for an answer, and later releases a valid lease during graceful shutdown
2. **THEN** it still has one recorded failure and resumes from its last committed checkpoint.

#### Scenario: Handler panic

1. **WHEN** an eligible handler panics while another task is queued
2. **THEN** it records one safe `handler_panic` outcome, the worker releases its slot, and the other task can finish.

### Requirement: Commit accepted answers with recoverable continuation

A handler needing input SHALL atomically save its checkpoint and pending request, release its lease and slot, and enter a waiting state. The owner’s answer SHALL be validated against ownership, execution, version, current request, schema, and task expiry. The answered record, its reference in the checkpoint, the queued continuation, and the replayable response SHALL commit together. No success SHALL be acknowledged before commit. Recovery and explicit restart SHALL retain those references and read the original answered records, including records belonging to earlier executions of the same task. A worker SHALL NOT overwrite committed accepted-input references with an older checkpoint or reopen an answered request.

#### Scenario: Answer survives before another handler checkpoint

1. **WHEN** an answer commits and the API and worker stop before the handler saves another checkpoint
2. **THEN** recovery reads the same accepted answer and continues without asking again or accepting it twice.
3. **WHEN** that execution instead expires before its next handler invocation and the owner explicitly restarts it with a valid expiry
4. **THEN** the new execution retains the original answer reference while the earlier request remains answered and immutable.

#### Scenario: Answer transaction rolls back

1. **WHEN** answer acceptance is interrupted before commit
2. **THEN** neither an accepted answer nor a queued continuation nor a success response is committed, and replay with the same key can complete the mutation once.

#### Scenario: A wait releases capacity

1. **WHEN** one task waits for input and another task becomes eligible with a single configured worker slot
2. **THEN** the second task finishes while the first retains its pending request and checkpoint.

### Requirement: Pause and stop invalidate pending execution and answers

The owner SHALL be able to pause or cancel an eligible active task. Pause SHALL retain committed progress and accepted answers, release the execution slot, invalidate its lease and unresolved input request IDs, and require explicit resume. New answers submitted while paused or cancelled, or to closed request IDs, SHALL be rejected without storing the answer or queuing work. Resume SHALL reissue still-needed input under a new request ID. Cancellation SHALL end eligibility for that execution and close unresolved input requests. Races with completion SHALL resolve atomically to one terminal outcome.

#### Scenario: Late answer after pause and resume

1. **WHEN** the owner pauses a task awaiting input and an answer to its old request arrives
2. **THEN** the answer is rejected without storage and the task remains paused, including after restart.
3. **WHEN** the owner resumes and then answers the old request ID again
4. **THEN** the old answer is still rejected; only the new current request can be answered.

#### Scenario: Stop races completion

1. **WHEN** cancellation and completion race on a running task
2. **THEN** exactly one terminal transition wins and produces an outcome; a stale worker cannot resurrect cancelled work, and a cancellation carrying the pre-completion version returns 409 if completion wins.

### Requirement: Distinguish task restart from lease recovery

Lease loss SHALL recover an eligible task within the current execution's retry limit. An explicit restart of a failed or expired task SHALL create a new execution under the same task ID with zero failed attempts. Prior attempts and outcomes SHALL remain readable. The new execution SHALL retain the last compatible checkpoint and accepted answers. Restart of an expired task SHALL require a new valid expiry or explicit removal of expiry. Successful and cancelled executions SHALL NOT be resumed by that endpoint.

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

Task detail SHALL expose current execution state, due time, progress, attempts, outstanding input, safe failure information, and stable outcome/file references. Task lists SHALL support owner-scoped pagination and conversation/state filters. Tasks and Activity SHALL derive their display from persisted records and link to the same outcome shown in chat. Oversized or invalid task/checkpoint payloads SHALL be rejected before mutations.

#### Scenario: Reopen task detail

1. **WHEN** the API restarts after one failed execution and a successful explicit restart
2. **THEN** task detail retains both executions and their attempts, identifies the current successful outcome, and exposes its stable result references.
