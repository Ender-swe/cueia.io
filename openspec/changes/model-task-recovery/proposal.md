# Cueia task execution and recovery

## Why

An owner can leave a job running, respond when it needs input, and recover its result after a worker stops. The task ID in Activity must remain usable through every transition.

## What Changes

### Task record and transitions

Create each task in the workspace resolved from the owner's session. Save its ID, request, type, source thread or file, initial plan, and queued state before acknowledging immediate work. For a future reminder, commit the scheduled state and due time together and return the saved ID and time to chat. The Tasks panel reads Scheduled tasks and Today's activity from these persisted records. [Bootstrap chat workspace](../bootstrap/design.md#chat-files-and-tasks)

Persist attempts with their original start time, progress, evidence IDs, input answers, checkpoint, and outcome. Expose create, inspect, answer, pause, resume, cancel, and eligible retry operations. Task detail must explain what work is due, what input or review is outstanding, and where to open its result.

| Current state | Event | Result |
| --- | --- | --- |
| Queued or due scheduled | Worker wins claim | Running with an attempt and expiring lease |
| Running | Needs information or approval | Save checkpoint and input request or review ID, release lease, and wait |
| Waiting for input or review | Valid answer or ledger decision arrives | Resume from saved context and settle or continue according to the decision |
| Running | Next observation due later | Save progress, due time, and scheduled state |
| Eligible active task | Owner pauses or cancels | Invalidate the lease. Pause retains progress for resume. Cancellation ends eligibility and closes pending reviews. |
| Running | Handler finishes or exhausts retries | Save evidence and a succeeded or failed outcome. Publish terminal status with its result or error. |

### Worker behavior

Run the scheduler in-process or from `backend/cmd/worker`. Extend bootstrap's version-checked PostgreSQL claim through `pgx` to test status and due time and acquire a unique lease token and expiry in the winning update. Renew the lease during long work. A worker losing the lease must discard stale progress. Check ownership of the lease at each durable write and before starting an effect. For an already-dispatched external write, reconcile the stored review and receipt. Cancellation cannot retract an effect already accepted by a provider. [Bootstrap claim](../bootstrap/design.md#make-the-claim-decision-in-postgresql)

Store reusable file, review, and result IDs under a stable operation identity. Resolve file inputs and outputs through the task's permanent workspace ID and the [bootstrap file service](../bootstrap/design.md#resolve-workspace-ownership-at-the-storage-boundary). A recovering worker can run on a different host and reopen the same registered file on the owner's disk. Recover an expired attempt from its last durable checkpoint and reconcile effects recorded after that checkpoint. Limit concurrent jobs while allowing unrelated jobs to progress during slow work or review. A paused task becomes eligible only on an explicit resume. A future scheduled task stays idle until due.

### Acceptance with a controlled clock and worker barriers

1. Release two separate Go workers against one eligible task in Docker PostgreSQL. Record handler entry and durable outcomes. Expect one handler and one outcome. In another run, cancel after claim and before effect dispatch. Release the stale worker and expect zero effects and a cancelled task.
2. Checkpoint a PDF input by its registered `fileId` and terminate the worker. Edit and register that path again. Give a second owner different bytes at the same path. Reopen the database, advance beyond lease expiry, and resume on a separate host. Verify the worker reads the first version's exact bytes by its checkpointed ID, finishes the original task, and retains the interrupted attempt. A late update carrying the old lease must have no effect.
3. Leave one task waiting for review and queue another. The second task finishes while the first retains its review. Deliver the same decision twice and verify one transition and one recorded outcome. Restart while paused and verify that only an explicit resume makes the task eligible.
4. Schedule a reminder for 30 minutes before a 10:00 workshop. Verify the creation response and the list read in a second session contain the same task ID and 09:30 due time in the event's zone. Restart the worker before 09:30. Verify zero claims before the deadline and one successful claim when the reminder comes due.

## Capabilities

### New Capabilities

- `durable-tasks` covers task ownership, scheduling, checkpoints, cancellation, and recovery after a worker stops.

### Modified Capabilities

## Impact

The API, worker, Tasks panel, and Activity view consume the same persisted task state. Bootstrap owns PDF registration and immutable reads. Document processing decides when a retry can reuse a filled copy. Connectors own external action receipts, and watches own observation delivery. Each workflow supplies its own recovery fixtures.
