# Cueia task execution and recovery

## Why

An owner should be able to leave a delegated job running, stop or pause it, and recover its progress after a worker stops. Its result or failure must return to the chat it belongs to, with the same task ID available after reopening the app.

## What Changes

- Persist owner-bound tasks, execution attempts, versioned checkpoints, accepted answers, operation identities, and outcomes before acknowledging changes. Reuse the existing document service for registered PDF versions and reads; keep execution metadata in PostgreSQL.
- Extend the bootstrap version-checked claim into an expiring, renewable worker lease. Recover interrupted work on another host, reject stale writes, and keep unrelated jobs moving while a task waits for input or review.
- Automatically retry safe transient failures with backoff within a persisted two-minute recovery budget. Exclude scheduled, paused, and input/review waiting time. Exhaustion produces a saved failure message; an explicit restart after failure or task expiry creates another execution under the same task ID.
- Add create, list, inspect, answer, pause, resume, cancel, and restart operations. Reject late answers while paused or cancelled, retain already accepted progress, and close pending reviews on cancellation. Lease expiry recovers automatically; review expiry never grants permission to act.
- Add a visible **Stop task** button in the associated message input. It cancels the selected durable task; stopping an ordinary chat turn retains its separate meaning. Tasks and Activity show persisted state and link to the result.
- Persist a task's destination conversation at creation. Publish its result or failure exactly once per execution to that conversation, using Main chat only when no originating chat was supplied. Commit the outcome and pending delivery together, recover delivery after restart, and replay messages when the chat reopens.
- Establish the minimum durable conversation identity and task-message read path needed for this delivery. Later conversation history adopts these IDs and message identities rather than creating another Main chat or duplicating results.
- Save mutation responses with workspace-scoped request keys so a lost response can be replayed without repeating the mutation. Introduce canonical `/v1/api` endpoints while retaining the existing `/api` routes as compatibility aliases with identical authorization.
- Support persisted one-shot due times and optional task expiry. The initial scheduling assumption is one catch-up run after downtime unless expired, with a fixed saved due time. Event-driven rescheduling and recurring reminder policy remain a follow-up discussion.

### Acceptance with a controlled clock and worker barriers

1. Release two Go workers against one eligible task in Docker PostgreSQL. Expect one claim winner, handler entry, and outcome. Cancel after claim but before effect authorization; releasing the stale worker produces zero effects.
2. Checkpoint a registered PDF, stop the worker, edit and re-register its source path, and give another owner different bytes at that path. Resume on another host after lease expiry. Read the first version's exact bytes, retain the interrupted attempt, and reject late writes from the old lease.
3. Leave one task waiting for review while another finishes. Apply a decision twice and observe one transition. Pause a task, reject a late answer without saving it, restart the service, and verify that only explicit resume makes it eligible.
4. Fail a safe operation repeatedly, restart during backoff, and advance the clock through two minutes. Observe a saved failure and one message in its chat. An explicit restart preserves the task ID and history. An uncertain external send is never retried automatically.
5. Save a 09:30 reminder for a 10:00 workshop. Lose the creation response, repeat the same request key, and read the same task ID and due time from another session. Observe no claim before due; exercise catch-up and expiry independently.
6. Finish tasks from Main chat and a side chat while the client is closed. Interrupt between outcome commit and message publication, then retry publication after a lost acknowledgement. Reopen both chats and find one correctly routed message per execution with stable result/file IDs; another owner sees neither.
7. Use the composer Stop task button while unrelated chat work and another task are active. Only the selected task stops. Verify the button and result messages at desktop and narrow widths, preserving the draft.

## Capabilities

### New Capabilities

- `durable-tasks`: task ownership, leases, scheduling, checkpoints, retries, controls, and restart recovery.
- `task-chat-results`: durable conversation destinations, task-result delivery and replay, and task controls in the message input.

### Modified Capabilities

- `workspace-access`: add versioned API access with compatible legacy routes and preserved authentication and preview semantics.

## Impact

Implements [issue #10](https://github.com/Ender-swe/cueia.io/issues/10), the second change in [PROPOSAL-ORDER.md](../../../PROPOSAL-ORDER.md). Affects PostgreSQL migrations, the Go API and worker, OpenAPI/generated types, the dashboard composer, conversation feed, and Tasks panel.

Bootstrap supplies sessions and immutable workspace files. `build-agent-runtime` supplies model execution later; controlled handlers exercise this change independently. `build-connectors` owns real review decisions and external receipts; this change defines and tests the integration contract. `map-conversations` adopts the minimal conversation registry and task messages when adding Rich Threads and full history. Document processing owns filled-copy reuse, and watches own recurrence and observation delivery. No new broker or external model service is required for acceptance.
