# Cueia task execution and recovery

## Why

An owner can leave a job running, respond when it needs input, and recover its result after a worker stops. The task ID in Activity must remain usable through every transition.

## What Changes

### Task record and transitions

On creation, assign an owner-scoped task ID and save the request, task type, source thread or file, initial plan, and queued state before acknowledging the client. Persist attempts with their original start time, progress, evidence IDs, input answers, checkpoint, and outcome. Expose create, inspect, answer, pause, resume, cancel, and eligible retry operations. Task detail must explain what work is due, what input or review is outstanding, and where to open its result.

| Current state | Event | Result |
| --- | --- | --- |
| Queued or due scheduled | Worker wins claim | Running with an attempt and expiring lease |
| Running | Needs information or approval | Save checkpoint and input request or review ID, release lease, and wait |
| Waiting for input or review | Valid answer or ledger decision arrives | Resume from saved context and settle or continue according to the decision |
| Running | Next observation due later | Save progress, due time, and scheduled state |
| Eligible active task | Owner pauses or cancels | Invalidate the lease. Pause retains progress for resume. Cancellation ends eligibility and closes pending reviews. |
| Running | Handler finishes or exhausts retries | Save evidence and a succeeded or failed outcome. Publish terminal status with its result or error. |

### Worker behavior

Provide in-process and standalone Bun entry points using the same scheduler. Select eligible work by status and due time. Atomically claim the task with a unique lease token and expiry, and renew during long work. A worker losing the lease must discard stale progress. Check ownership of the lease at each durable write and before starting an effect. For an already-dispatched external write, reconcile the stored review and receipt. Cancellation cannot retract an effect already accepted by a provider.

Store reusable file, review, and result IDs under a stable operation identity. Recover an expired attempt from its last durable checkpoint and reconcile effects recorded after that checkpoint. Limit concurrent jobs while allowing unrelated jobs to progress during slow work or review. A paused task becomes eligible only on an explicit resume. A future scheduled task stays idle until due.

### Acceptance with a controlled clock and worker barriers

1. Release two separate Bun workers against one eligible task in Docker PostgreSQL. Record handler entry and durable outcomes. Expect one handler and one outcome. In another run, cancel after claim and before effect dispatch. Release the stale worker and expect zero effects and a cancelled task.
2. Save a file and its ID, terminate the worker, reopen the database, advance beyond the lease expiry, and start another worker. It must reopen the file, finish the original task, and retain the interrupted attempt in history. A late update carrying the old lease must have no effect.
3. Leave one task waiting for review and queue another. The second task finishes while the first retains its review. Deliver the same decision twice and verify one transition and one recorded outcome. Restart while paused and verify that only an explicit resume makes the task eligible.

## Capabilities

### New Capabilities

- `durable-tasks` covers task ownership, scheduling, checkpoints, cancellation, and recovery after a worker stops.

### Modified Capabilities

## Impact

The API, worker, and Activity view consume the same persisted state. Document processing owns file-reuse rules, connectors own external action receipts, and watches own observation delivery. Each workflow supplies its own recovery fixtures.
