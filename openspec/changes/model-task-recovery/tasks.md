# Model task recovery implementation tasks

Implements [issue #10](https://github.com/Ender-swe/cueia.io/issues/10). Use the session feature layout and put Go tests in the closest `test/` directory. [Design and schema](design.md) · [Durable tasks](specs/durable-tasks/spec.md) · [Chat results](specs/task-chat-results/spec.md) · [Versioned access](specs/workspace-access/spec.md)

## 1. Persist tasks, destinations, and outcomes

- [x] 1.1 Add the seven tables and ownership constraints in the design. Test a fresh install, an upgrade from bootstrap, one Main conversation per workspace, one open attempt per run, and one outcome per execution against PostgreSQL.
- [ ] 1.2 Implement workspace-scoped mutation replay. Commit the validated request fingerprint and response with the mutation. Test concurrent duplicates, rollback, key reuse with a different payload, and authorization on replay.
- [ ] 1.3 Resolve owned Main and side conversations and create the task, initial run, immutable destination, and optional due time/expiry together. Verify that only supported task types and owned file/source IDs are accepted and that a lost creation response returns the original task ID.
- [ ] 1.4 Implement the shared terminal transition in `internal/task/`. Commit the final state and immutable outcome together. Exercise rollback and lost acknowledgement, then read the same message ID with all workers offline.

## 2. Run and recover reminders and PDF reads

- [ ] 2.1 Extend the Go worker with a pool, bounded execution slots, conditional claims, and lease renewal. Fence every worker write by run, token, and fresh database time. Extend bootstrap's separate-process barrier test to observe one handler entry and reject a stale checkpoint.
- [ ] 2.2 Run the one-shot reminder against its saved instant and zone. Verify no early claim, one catch-up after downtime, and a fixed due time after a source-event edit. Enforce optional task expiry after locking and during idle polling, including paused and waiting tasks.
- [ ] 2.3 Wire the worker to the same workspace provider as the API. Implement the read-only PDF handler and versioned checkpoint. Verify original bytes by registered `fileId` after a path edit and reject unsupported handler/checkpoint versions before invoking a handler.
- [ ] 2.4 Implement the persisted failure count, backoff, and two-minute timeout per attempt. Clamp renewals to the attempt deadline. With a controlled clock, verify stalled handlers, repeated lease loss, reuse of remaining retries after a long outage, and exactly one failure outcome on the third failed attempt.
- [ ] 2.5 Handle graceful shutdown and handler panics. Verify a clean release of a valid lease preserves the count and checkpoint, a panic fails the eligible run, and an unrelated task can still finish.

## 3. Accept input and owner controls

- [ ] 3.1 Save a pending input request and checkpoint when releasing the lease. Atomically validate and save an answer, its checkpoint reference, the queued continuation, and the replayable response. Test single-slot progress during a wait, rollback, and restart immediately after answer commit.
- [ ] 3.2 Implement Pause, Resume, and Stop against the selected run and version. Preserve progress and the failure count, close old requests, and reissue needed input on resume. Verify pause survives restart, late answers are rejected, and a Stop/completion race yields one terminal outcome.
- [ ] 3.3 Implement Start again for failed or expired runs. Retain the task ID, history, and compatible checkpoint with accepted answers. Test expiry replacement, a fresh retry allowance, duplicate restarts, delayed controls, and replay of an old response after a new run starts.

## 4. Serve the versioned API and connect the dashboard

- [ ] 4.1 Mount the same feature mux at canonical `/v1/api` and compatible `/api`. Share session protection, rate limits, origin rules, and idempotency scope. Allow the required preflight headers, issue canonical preview links, and test existing signed links and cross-prefix exchange/revocation.
- [ ] 4.2 Register the relative task and conversation routes. Implement owned list/detail reads and cursor-paginated chat messages by joining outcomes to their task's saved destination. Verify same-timestamp pagination, Main fallback, foreign-resource denial, and one message per execution after manual restart.
- [ ] 4.3 Document the routes, controls, limits, errors, and compatibility aliases in OpenAPI. Regenerate `packages/api-types` and update the dashboard API configuration, development proxy, and deployment health checks to use the canonical prefix. Verify API and worker startup still depend on successful migration.
- [ ] 4.4 Connect authenticated chat, Scheduled tasks, and Today's activity to persisted data. Register side chats before delegation, refresh outcomes on open/focus and while visible, merge by outcome ID, and open the original file through a fresh preview. Follow `DESIGN.md` and the uncodixfy skill for dashboard changes.
- [ ] 4.5 Bind the keyboard-accessible composer Stop button and task-detail controls to the selected task, run, and version. Let the owner answer a pending request. Show committed responses, preserve the draft, retry ambiguous requests with the original key/payload, and refresh conflicts without retargeting a newer execution.

## 5. Verify the user journeys

- [ ] 5.1 Run the PDF recovery scenario with separate Go workers, controlled time, and barriers against Docker PostgreSQL. Kill the first worker after checkpoint, replace the source file, recover on another host, and submit a late write under the old lease. Reopen the original result in a second session. Repeat with Stop winning before completion and verify cancellation is the only outcome.
- [ ] 5.2 Run browser acceptance for a 09:30 reminder, lost-response replay across prefixes from the authorized cross-origin client, and one saved result after reopening. Check Main/side-chat isolation, a disallowed origin, failed-then-restarted history, and narrow-screen keyboard/draft preservation. Record the provenance of original fixtures and expected results.
- [ ] 5.3 Run the affected Go and browser suites, generated-contract check, root build/type-check/test tasks, and `openspec validate model-task-recovery --strict`. Record acceptance evidence and any environment limits before marking implementation complete.
