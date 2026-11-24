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

Pause, resume, cancel, restart, and answer mutations SHALL require `expectedRunNumber` and `expectedVersion`. Except for replay of a committed response, both values SHALL match the current execution atomically or the request SHALL return 409 without applying its requested change. Lifecycle transitions, accepted inputs/decisions, and changes to actionable input/review requests SHALL increment the execution's control version; routine heartbeat renewal and progress-only checkpoint updates SHALL NOT. Task reads and mutation responses SHALL expose the execution number and control version.

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

The worker SHALL atomically check current execution, version, state, due time, retry eligibility, and expiry when claiming work. A winning claim SHALL create an attempt with an original start time and a unique expiring lease. Workers SHALL renew leases during long work. Every progress write, checkpoint, completion, and effect authorization SHALL verify the current running execution and unexpired lease. Losing ownership SHALL stop local execution and prevent stale durable changes.

#### Scenario: Competing processes

1. **WHEN** two separate Go workers against Docker PostgreSQL are released at a barrier against the same eligible task
2. **THEN** one enters its handler under a valid lease, one attempt is created, and one durable outcome is recorded.

#### Scenario: Renewal and stale completion

1. **WHEN** a long operation renews its lease before expiry
2. **THEN** a competing worker cannot reclaim it.
3. **WHEN** renewal later fails and another worker acquires the expired lease
4. **THEN** the old worker cannot save progress, authorize an effect, or complete the task.

### Requirement: Enforce deadlines independently of background sweeps

Every transaction that advances work, accepts input, changes eligibility, renews a lease, authorizes an effect, or completes a nonterminal execution SHALL evaluate task expiry and any active recovery deadline using fresh authoritative database time after acquiring the relevant locks. A deadline SHALL be elapsed at equality. An elapsed deadline SHALL prevent the requested work and produce the terminal outcome and delivery intent atomically, without waiting for a sweep. Task expiry SHALL take precedence over recovery exhaustion when both have elapsed. Entering pause or a waiting state SHALL NOT suspend an already exhausted budget. Trusted late receipts SHALL remain recordable as evidence without reopening the execution or changing its terminal outcome.

#### Scenario: Expiry with the sweeper stopped

1. **WHEN** a running task's lease is still valid, deadline sweeps are held, and the controlled clock reaches its task expiry
2. **THEN** attempts to authorize an effect or complete the task produce an expired outcome with no new effect or success, regardless of which request arrives first.

#### Scenario: Recovery deadline cannot be escaped by waiting

1. **WHEN** recovery has exhausted its budget while sweeps are held and a handler attempts to enter input/review wait or the owner attempts pause
2. **THEN** the execution becomes failed with delivery intent rather than freezing an exhausted budget.
3. **WHEN** task expiry has also elapsed in the same setup
4. **THEN** the terminal outcome is expired instead of failed.

#### Scenario: Lock wait crosses a deadline

1. **WHEN** a completion transaction begins before expiry but waits for the run lock until the deadline has elapsed
2. **THEN** its post-lock time check rejects success and records the expired outcome.

#### Scenario: Receipt arrives after a terminal outcome

1. **WHEN** a provider receipt arrives after cancellation, task expiry, or recovery exhaustion
2. **THEN** it is retained under its operation identity without changing the terminal outcome or creating another terminal message.

### Requirement: Recover versioned checkpoints and immutable document inputs

The system SHALL save versioned execution checkpoints containing completed and remaining work, accepted input references, stable operation keys, evidence, and registered file IDs. Document reads SHALL reuse the existing document/workspace services and the task's permanent workspace. Recovery SHALL preserve the task ID, retain the interrupted attempt and its original start time, and create a new attempt from the last compatible checkpoint. An unsupported handler or checkpoint version SHALL fail visibly before effects.

#### Scenario: Resume the original file on another host

1. **WHEN** a worker checkpoints a PDF by `fileId` and stops, the owner edits and registers that path again, and another owner stores different bytes at the same path
2. **THEN** recovery on a separate host after lease expiry, within the remaining recovery budget, reads the first registered version's exact bytes and retains the interrupted attempt.
3. **WHEN** the old worker submits a late checkpoint
4. **THEN** it changes neither progress nor the recovered outcome.

#### Scenario: Incompatible saved state

1. **WHEN** a claimed task contains a checkpoint or handler version unsupported by the installed worker
2. **THEN** the task records a safe incompatibility failure and chat delivery intent, with zero handler effects and no silent restart from empty state.

### Requirement: Reconcile operations before repeating effects

Each effectful step SHALL have a stable operation identity and normalized input fingerprint. Recovery and manual restart SHALL reuse confirmed results and reconcile receipts created after the last checkpoint. A mismatched fingerprint SHALL be rejected. A dispatch with an uncertain result SHALL NOT be automatically repeated. Review decisions SHALL be consumed idempotently and revalidated against workspace, task eligibility, payload, connection generation, and expiry before dispatch. Cancellation SHALL invalidate pending reviews and prevent new authorization; it SHALL NOT claim to retract an already-dispatched provider effect.

An abandoned `authorized` operation SHALL be treated as potentially executed even when no receipt or uncertainty was saved. Reconciliation SHALL reuse confirmed success, permit fresh authorization only with authoritative proof that the old dispatch did not and cannot execute, or persist uncertainty and fail the active execution without resending. Absence of a receipt SHALL NOT count as that proof. Fresh authorization SHALL recheck the current execution, deadlines, lease, and review. Explicit restart SHALL obey the same reconciliation rules.

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

#### Scenario: Worker dies around dispatch without saving a receipt

1. **WHEN** a separate worker process is killed after authorization but before the provider call, and separately after provider acceptance but before receipt storage
2. **THEN** recovery reconciles the saved authorization in both cases and issues no resend based only on a missing receipt.
3. **WHEN** lookup confirms success or remains inconclusive
4. **THEN** recovery respectively reuses the confirmed result or saves an uncertain failure, including after explicit restart.

#### Scenario: Adapter proves dispatch cannot execute

1. **WHEN** reconciliation obtains authoritative evidence that the old dispatch never executed and cannot execute later
2. **THEN** a new authorization is allowed only while the current execution, lease, deadlines, and reviewed payload remain valid.

### Requirement: Bound automatic recovery and fail handler panics visibly

Workers SHALL automatically retry transient failures only when repetition is safe. The first retryable failure or unexpected lease expiry SHALL start a persisted two-minute recovery budget, using backoff and a persisted next-attempt time. For unexpected lease expiry, the budget SHALL start at the saved lease expiry, not at later discovery. Subsequent failures and lease losses SHALL retain the budget. Active recovery and retry backoff, including worker downtime, SHALL consume it. Scheduled waits, explicit pauses, and input/review waits SHALL preserve its remaining duration without consuming or resetting it; returning to queued work SHALL restore its deadline immediately rather than waiting for a claim. Graceful shutdown SHALL neither start a new budget nor suspend or reset an existing active budget. Exhaustion SHALL invalidate the lease, save a failure message, and enqueue chat delivery without another handler invocation. Initial healthy execution SHALL NOT be limited to two minutes solely by this recovery policy.

A recovered handler panic SHALL produce a safe `handler_panic` failure and delivery intent when the execution is still eligible, without automatic retry. An earlier terminal outcome or elapsed deadline SHALL remain authoritative. Potential effects SHALL retain their reconciliation records; neither panic handling nor later explicit restart SHALL interpret absent receipts as proof of no effect.

For recovery suspended by a future scheduled wait, exclusion SHALL end at the saved due instant. Its effective deadline SHALL be that instant plus the saved remaining budget even if workers discover eligibility later. A first execution whose recovery budget has never started SHALL retain the ordinary one-shot catch-up policy.

#### Scenario: Restart during retry backoff

1. **WHEN** a safe operation repeatedly fails, the worker restarts during backoff, and the controlled clock reaches the original recovery deadline
2. **THEN** no additional retry begins, the task becomes failed, and its saved message explains that two minutes of recovery were exhausted and that the user can start it again.

#### Scenario: Wait preserves remaining budget

1. **WHEN** 30 seconds of recovery have elapsed and the task is paused for five minutes
2. **THEN** its remaining recovery budget stays at 90 seconds.
3. **WHEN** the owner explicitly resumes it and safe failures continue
4. **THEN** it fails after those remaining 90 seconds rather than receiving a new two minutes.

#### Scenario: A retry stalls

1. **WHEN** a retry handler remains blocked past the persisted recovery deadline
2. **THEN** the task becomes failed through deadline enforcement and a late handler result cannot overwrite that outcome.

#### Scenario: Repeated worker deaths before reporting failure

1. **WHEN** a worker dies before reporting a handler failure and replacement workers repeatedly die after their leases are acquired
2. **THEN** recovery retains the first expired lease's expiry plus two minutes as its deadline and fails without another handler invocation once that deadline is reached.

#### Scenario: Workers return after the recovery window

1. **WHEN** all workers remain offline past the first unexpectedly expired lease's expiry plus two minutes
2. **THEN** discovery records a recovery-exhausted failure with one delivery intent instead of starting another attempt, and exposes explicit restart.

#### Scenario: Graceful shutdown and resumed queue time

1. **WHEN** a healthy initial handler releases its lease through graceful shutdown
2. **THEN** the release does not start a recovery budget.
3. **WHEN** the same shutdown occurs during active recovery, or an input answer resumes a suspended recovery budget while no execution slot is free
4. **THEN** queued time consumes the existing remaining budget and cannot grant a fresh two minutes.

#### Scenario: Handler panic

1. **WHEN** an eligible controlled handler panics and the worker catches it at the handler boundary
2. **THEN** the task records `handler_panic` and one delivery intent, retains operation evidence, and is not automatically retried while unrelated tasks can continue.

#### Scenario: Scheduled recovery becomes due during downtime

1. **WHEN** a recovering task schedules a later continuation with 90 seconds remaining and workers return more than 90 seconds after that due instant
2. **THEN** it records recovery exhaustion without another handler invocation, while a separate unexpired one-shot reminder whose recovery budget never started remains eligible for catch-up.

### Requirement: Commit accepted answers with recoverable continuation

Answer acceptance SHALL atomically validate ownership, current execution/version, pending request, answer schema, and deadlines; save the immutable answered-input record; update the checkpoint's accepted-input references; make continuation eligible; and save the replayable response. No success SHALL be acknowledged before commit. Recovery and explicit restart SHALL retain those references and read the original answered records, including records belonging to earlier executions of the same task. A worker SHALL NOT overwrite committed accepted-input references with an older checkpoint or reopen an answered request.

#### Scenario: Answer survives before another handler checkpoint

1. **WHEN** an answer commits and the API and worker stop before the handler saves another checkpoint
2. **THEN** recovery reads the same accepted answer and continues without asking again or accepting it twice.
3. **WHEN** that execution instead expires before its next handler invocation and the owner explicitly restarts it with a valid expiry
4. **THEN** the new execution retains the original answer reference while the earlier request remains answered and immutable.

#### Scenario: Answer transaction rolls back

1. **WHEN** answer acceptance is interrupted before commit
2. **THEN** neither an accepted answer nor a queued continuation nor a success response is committed, and replay with the same key can complete the mutation once.

### Requirement: Pause and stop invalidate pending execution and answers

The owner SHALL be able to pause or cancel an eligible active task. Pause SHALL retain committed progress and accepted answers, release the execution slot, invalidate its lease and unresolved input request IDs, and require explicit resume. New answers submitted while paused or cancelled, or to closed request IDs, SHALL be rejected without storing the answer or queuing work. Resume SHALL reissue still-needed input under a new request ID. Cancellation SHALL end eligibility for that execution and close pending reviews. Races with completion SHALL resolve atomically to one terminal outcome.

#### Scenario: Late answer after pause and resume

1. **WHEN** the owner pauses a task awaiting input and an answer to its old request arrives
2. **THEN** the answer is rejected without storage and the task remains paused, including after restart.
3. **WHEN** the owner resumes and then answers the old request ID again
4. **THEN** the old answer is still rejected; only the new current request can be answered.

#### Scenario: Stop races completion

1. **WHEN** cancellation and completion race on a running task
2. **THEN** exactly one terminal transition wins and produces an outcome; a stale worker cannot resurrect cancelled work, and a cancellation carrying the pre-completion version returns 409 if completion wins.

### Requirement: Wait without blocking unrelated tasks

A task needing input or review SHALL atomically save its checkpoint and pending request/review before releasing its lease and execution slot. Valid answers and decisions SHALL advance it once. A paused task SHALL NOT resume automatically when a review decision arrives; explicit resume SHALL revalidate the saved decision. Denial, expiry, or invalid connection identity SHALL NOT permit an external action.

The task service SHALL persist a stable review-preparation intent before calling the adapter. Repeated preparation for the same operation, review generation, and payload SHALL resolve to the same ledger review. A fresh review SHALL use a new persisted generation. A periodic reconciler SHALL discover saved decisions and review expiry for waiting tasks independently of notifications. It SHALL validate the current execution and review generation, then atomically record decision consumption and queue continuation once. Notifications SHALL only accelerate reconciliation; the connector ledger SHALL remain the authority for reviews.

#### Scenario: Another task finishes during review

1. **WHEN** one task waits for review and another eligible task is queued
2. **THEN** the second task finishes while the first retains its pending review and saved progress.
3. **WHEN** the same valid review decision is delivered twice
4. **THEN** one continuation and one recorded effect outcome result.

#### Scenario: Approval expires while paused

1. **WHEN** a task is paused, a review decision arrives, and that review expires before explicit resume
2. **THEN** the task remains paused until resumed and requires fresh review before any send.

#### Scenario: Review preparation response is lost

1. **WHEN** the adapter saves a review but the worker stops before recording the returned review ID
2. **THEN** recovery repeats the saved preparation identity and obtains the same review rather than creating another pending action.

#### Scenario: Early approval and lost notification

1. **WHEN** approval is saved before the task commits waiting state, or its notification is lost after the ledger commits
2. **THEN** periodic reconciliation discovers the saved decision and commits one continuation without another user approval.
3. **WHEN** a reconciler stops before committing continuation and another pass reads the same decision
4. **THEN** decision consumption and continuation still commit exactly once.

#### Scenario: Review expires without a notification

1. **WHEN** a waiting review expires and no callback is delivered
2. **THEN** reconciliation makes the expiry available to the handler without authorizing a send; any required fresh review uses a new generation.
3. **WHEN** the task is paused in the same setup
4. **THEN** it remains paused and only explicit resume can revalidate or replace its review.

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
