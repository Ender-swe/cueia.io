# Cueia task execution and recovery

## Why

An owner should be able to schedule a workshop reminder, close the app, and return to its result in the same chat. If a worker stops while reading a document, another worker should resume from the saved checkpoint while the owner retains the ability to pause or stop the task.

## What Changes

- Save the task ID, workspace, originating conversation and any one-shot due time before confirming creation. Use the owner's Main chat when no origin was supplied. An overdue reminder runs once when workers return, unless its optional task expiry has passed.
- Extend the bootstrap PostgreSQL claim with a renewable lease. Run a reminder and a read-only PDF handler. A replacement worker uses the last versioned checkpoint and the original registered `fileId`. Only the current lease holder can save handler progress or commit a successful result.
- Bound each attempt to two minutes. A transient error, timeout, or unexpected lease loss counts as one failed attempt. Allow two automatic retries with backoff per execution. A long outage leaves unused retries available. Exhaustion or a handler panic produces a visible failure.
- Support create, list, inspect, answer, pause, resume, cancel, and explicit restart. A task waiting for an answer releases its worker slot. Preserve accepted answers and earlier executions. Bind controls, including the composer's **Stop task** button, to the selected execution and version.
- Commit one immutable outcome with each terminal transition. Its ID is the chat message ID. Chat, Tasks, and Activity read that same record. An owner who was offline can open the result as soon as the database commit succeeds.
- Serve canonical `/v1/api` endpoints and keep `/api` as a compatibility alias to the same handlers. Both mounts share authentication, owner checks, and replayable mutation responses. The authorized browser origin can retry with the original `Idempotency-Key` and payload through either mount.

### Acceptance with a controlled clock and worker barriers

1. Save a 09:30 reminder through `/v1/api`, lose the creation response, and retry through `/api`. Expect the same task ID and due time. Close the client and bring two workers online after 09:30. Expect one claim and one reminder in Main.
2. In a side chat, checkpoint a PDF by `fileId` and kill the worker. Edit and re-register the source, then recover on another host after a long outage. Verify the original bytes. Inject a late write carrying the first lease. Reopen the chat in a second session and find one result with the same ID as in Tasks and Activity. Another workspace cannot read it.
3. Repeat with the replacement worker held before completion. Press **Stop task**, then release the worker. Reopen the chat and find only the cancellation for that execution. The draft and unrelated work survive.
4. Exhaust three attempts across worker restarts. Expect one failure and no fourth automatic attempt. Start again under the same task ID and finish. Both outcomes remain readable, and a delayed control for the first execution leaves the second alone.
5. Pause a task waiting for input, resume it and answer the new request. Restart before the next handler checkpoint. Verify the accepted answer survives and another task can run throughout the wait.

## Capabilities

### New Capabilities

- `durable-tasks` covers task ownership, leases, scheduling, checkpoints, bounded retries, controls, and restart recovery.
- `task-chat-results` covers durable conversation destinations, terminal outcomes readable as chat messages, and task controls in the message input.

### Modified Capabilities

- `workspace-access` adds versioned API access with compatible legacy routes and preserved authentication and preview semantics.

## Impact

Implements [issue #10](https://github.com/Ender-swe/cueia.io/issues/10), the second change in [PROPOSAL-ORDER.md](../../../PROPOSAL-ORDER.md). Affects PostgreSQL migrations, the Go API and worker, OpenAPI/generated types, the dashboard composer, conversation feed, and Tasks panel.

Bootstrap supplies sessions and immutable workspace files. `build-agent-runtime` supplies model execution. `map-conversations` adopts the conversation and outcome IDs for Rich Threads. Reviewed writes first become available when the connector ledger can recover approvals and reconcile provider receipts. Document processing owns filled-copy reuse, and watches own recurrence. [Connector review and receipts](../build-connectors/proposal.md#review-ledger-and-receipts)
