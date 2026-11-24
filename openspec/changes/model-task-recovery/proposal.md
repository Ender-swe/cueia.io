# Cueia task execution and recovery

## Why

An owner should be able to leave a delegated job running, stop or pause it, and recover its progress after a worker stops. Its result or failure must return to the chat it belongs to, with the same task ID available after reopening the app.

## What Changes

- Persist owner-bound tasks, execution attempts, versioned checkpoints, accepted answers, operation identities, and outcomes before acknowledging changes. Reuse the existing document service for registered PDF versions and reads; keep execution metadata in PostgreSQL.
- Extend the bootstrap version-checked claim into an expiring, renewable worker lease. Recover interrupted work on another host, reject stale writes, and keep unrelated jobs moving while a task waits for input or review.
- Automatically retry safe transient failures and recover unexpected lease loss within a persisted two-minute recovery budget. Start worker-loss recovery at the expired lease's saved expiry, count downtime, and never reset the budget on another interruption. Exclude scheduled, paused, and input/review waiting time; graceful shutdown does not start a new budget. Exhaustion and recovered handler panics produce saved failures; an explicit restart creates another execution under the same task ID.
- Add create, list, inspect, answer, pause, resume, cancel, and restart operations. Bind controls to the expected execution number and version so delayed requests cannot affect a restarted task. Commit accepted answers, checkpoint references, and continuation together; retain them through recovery and explicit restart. Reject late answers while paused or cancelled and close pending reviews on cancellation.
- Enforce task expiry and recovery deadlines inside mutation and effect-authorization transactions using fresh database time, independently of sweeps. Expiry takes precedence when both deadlines have elapsed; late receipts remain evidence without changing terminal outcomes.
- Reconcile abandoned dispatch authorizations before repeating effects. Missing receipts mean uncertainty unless the adapter proves the old dispatch did not and cannot execute. Prepare reviews idempotently and reconcile saved decisions and expiry even when notifications arrive early or are lost; keep review authority in the connector ledger.
- Add a visible **Stop task** button in the associated message input. It cancels the selected durable task; stopping an ordinary chat turn retains its separate meaning. Tasks and Activity show persisted state and link to the result.
- Persist a task's destination conversation at creation. Publish its result or failure exactly once per execution to that conversation, using Main chat only when no originating chat was supplied. Commit the outcome and pending delivery together, recover delivery after restart, and replay messages when the chat reopens.
- Establish the minimum durable conversation identity and task-message read path needed for this delivery. Later conversation history adopts these IDs and message identities rather than creating another Main chat or duplicating results.
- Save mutation responses with workspace-scoped request keys so a lost response can be replayed without repeating the mutation. Introduce canonical `/v1/api` endpoints while retaining the existing `/api` routes as compatibility aliases with identical authorization. Allow `Idempotency-Key` through CORS for the authorized origin and retain the same key and payload on browser retries.
- Support persisted one-shot due times and optional task expiry. An overdue reminder runs once after downtime unless expired, with a fixed saved due time. Event-driven rescheduling and recurring reminder policy remain a follow-up discussion.

### Acceptance with a controlled clock and worker barriers

1. Release two Go workers against one eligible task in Docker PostgreSQL. Expect one claim winner, handler entry, and outcome. Cancel after claim but before effect authorization; releasing the stale worker produces zero effects.
2. Checkpoint a registered PDF, stop the worker, edit and re-register its source path, and give another owner different bytes at that path. Resume on another host after lease expiry within the remaining recovery budget. Read the first version's exact bytes, retain the interrupted attempt, and reject late writes from the old lease.
3. Leave one task waiting for review while another finishes. Deliver approval early, lose its notification, and apply the saved decision twice; observe one continuation. Retry review preparation after response loss and observe one review. Pause a task, reject a late answer without saving it, restart the service, and verify that only explicit resume makes it eligible. Accept an answer and stop before another handler checkpoint; recovery and explicit restart retain it.
4. Fail a safe operation repeatedly, restart during backoff, and advance the clock through two minutes. Repeat with workers dying before reporting an error; observe one saved failure and message without a fresh budget. Verify graceful shutdown does not start a budget and a recovered handler panic fails visibly. Kill a separate worker after dispatch authorization and after provider acceptance, before any receipt is saved; recovery reconciles both and never resends on missing evidence.
5. Save a 09:30 reminder for a 10:00 workshop. Lose the creation response, repeat the same request key, and read the same task ID and due time from another session. Observe no claim before due; exercise catch-up and expiry independently.
6. Finish tasks from Main chat and a side chat while the client is closed. Interrupt between outcome commit and message publication, then retry publication after a lost acknowledgement. Reopen both chats and find one correctly routed message per execution with stable result/file IDs; another owner sees neither.
7. Use the composer Stop task button while unrelated chat work and another task are active. Only the selected task stops. Verify the button and result messages at desktop and narrow widths, preserving the draft.
8. Hold an execution-1 control request until execution 2 has the same numeric version; its first arrival returns 409 without affecting execution 2. Hold deadline sweeps, advance past task expiry or recovery exhaustion, and attempt completion, effect authorization, and entry into waiting; no expired work proceeds. Record a late receipt without replacing the terminal outcome.
9. From the authorized cross-origin browser, create a task, lose its response, and retry with the same key and payload through the other API prefix. Expect successful preflight and the same task ID. A disallowed origin remains rejected.

## Capabilities

### New Capabilities

- `durable-tasks`: task ownership, leases, scheduling, checkpoints, retries, controls, and restart recovery.
- `task-chat-results`: durable conversation destinations, task-result delivery and replay, and task controls in the message input.

### Modified Capabilities

- `workspace-access`: add versioned API access with compatible legacy routes and preserved authentication and preview semantics.

## Impact

Implements [issue #10](https://github.com/Ender-swe/cueia.io/issues/10), the second change in [PROPOSAL-ORDER.md](../../../PROPOSAL-ORDER.md). Affects PostgreSQL migrations, the Go API and worker, OpenAPI/generated types, the dashboard composer, conversation feed, and Tasks panel.

Bootstrap supplies sessions and immutable workspace files. `build-agent-runtime` supplies model execution later; controlled handlers exercise this change independently. `build-connectors` owns real review decisions and external receipts; this change defines and tests the integration contract. `map-conversations` adopts the minimal conversation registry and task messages when adding Rich Threads and full history. Document processing owns filled-copy reuse, and watches own recurrence and observation delivery. No new broker or external model service is required for acceptance.
