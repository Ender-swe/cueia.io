# Task execution and recovery design

## Context

This change implements [issue #10](https://github.com/Ender-swe/cueia.io/issues/10) after bootstrap. PostgreSQL already holds workspace ownership, sessions, file metadata, and versioned claim records. `internal/claim/claim.go` only changes an eligible record to claimed. `cmd/worker/main.go` polls through a single connection; it does not execute handlers or recover work. The API and workspace service already use `pgxpool`.

`internal/document/service.go` registers PDFs, reads immutable versions, and issues previews. Reuse that service for documents and its underlying workspace storage for bytes. It is not an execution checkpoint store: tasks must also save their plan, completed steps, answers, and operation identities in PostgreSQL.

The dashboard has Main chat, local side-chat IDs, a draft-only composer, and empty persisted task sections. There is no durable conversation registry or task-message feed. Delivering results to their originating chat therefore includes a small conversation registry and a durable task-message read path. Full model chat and Rich Threads remain with `build-agent-runtime` and `map-conversations`.

## Goals / Non-Goals

**Goals:**

- Preserve task identity, document versions, accepted answers, and execution history across worker/API restarts and host replacement.
- Permit one current worker lease, reject stale progress, and reconcile effects before repeating work.
- Bound safe transient retries and unexpected worker-loss recovery to two minutes of active recovery time, then explain the failure in the task's chat.
- Support explicit pause, resume, stop, and restart after failure or expiry.
- Persist and replay exactly one terminal chat message per task execution, including when the app is closed.
- Introduce `/v1/api` without breaking the bootstrap `/api` routes or weakening their authorization.

**Non-Goals:**

- Live model execution, general chat transcripts, composer queues, Rich Threads synchronization, or a second approval ledger.
- Recurring reminder semantics, automatic rescheduling after event edits, push/email notification delivery, and watch-specific observation rules.
- Serializing an in-memory model process or guaranteeing that an arbitrary external provider executes an action exactly once.

## Decisions

### 1. Keep execution in Go and PostgreSQL

Extend the existing worker as a separate process with a `pgxpool.Pool`, bounded handler concurrency, and independent lease renewal. Use short transactions for claims and transitions; never hold a task row lock during model, filesystem, or provider I/O. Start with four execution slots per process, a one-second poll, a 30-second lease, and renewal every 10 seconds. Put these tunable defaults in task configuration, validating positive values and renewal shorter than the lease. A waiting or paused task uses no slot. Only claim when a slot is available.

Use fresh PostgreSQL time sampled after acquiring the relevant locks for due checks, deadlines, and lease ownership. Do not authorize a transition using a timestamp captured before a lock wait or at transaction start. Tests inject one shared controlled clock into the database-time abstraction used by all worker processes; test clock control is unavailable in production. A worker-local timer only wakes polling or renewal, and cannot authorize a claim on its own.

During graceful shutdown, stop claiming, cancel local handler contexts, persist an interrupted attempt with finish reason `graceful_shutdown` and checkpoint when still authorized, and atomically release the lease into queued state. This release neither starts a recovery budget nor suspends or resets an existing active deadline. If that transaction cannot commit, normal unexpected-lease-loss rules apply. After a hard stop, the persisted lease expires naturally. A failed renewal cancels local work immediately; durable writes must still verify ownership independently. Catch panics at each handler boundary, release local resources, and save a safe `handler_panic` failure and delivery intent if the execution is still eligible. Do not automatically retry a recovered panic or overwrite an earlier terminal outcome; retain all operation evidence for reconciliation. A process-level crash that cannot be caught is handled through lease expiry.

**Alternative considered:** a separate queue/broker. PostgreSQL already owns task state and atomic claims, so another queue would introduce a second recovery boundary without being needed for this milestone. The old claim fixture remains available for bootstrap regression tests; real tasks use task-run rows.

### 2. Separate task identity, executions, attempts, and operations

- A **task** has a permanent workspace, request, destination chat, and ID.
- An **execution** (`run_number`) starts at 1. Explicit restart after `failed` or `expired` increments it under the same task ID. Each execution can produce one terminal outcome and chat message.
- An **attempt** records one worker acquisition, its original start time, progress, checkpoint snapshot, and finish reason. Lease recovery and automatic retry create attempts within the current execution.
- An **operation** identifies a logical step and normalized input fingerprint across attempts and executions. A manual restart must not forget a successful or uncertain external effect.

The current execution uses these states:

| State/event | Transition and durable work |
| --- | --- |
| Create immediate task | Save `queued`, run 1, destination, and replayable response together. |
| Create future task | Save `scheduled`, due instant and zone, run 1, destination, and response together. |
| Queued or due scheduled task | Atomically claim, create attempt, set `running` and lease token/expiry. |
| Running needs input/review | Save checkpoint and request/review identity, finish attempt as waiting, clear lease, enter `waiting_input` or `waiting_review`. |
| Valid input/review decision | Consume it once and queue continuation; handler interprets denial/expiry without dispatching an action. |
| Running needs a later observation | Save checkpoint and next due time, finish attempt, clear lease, enter `scheduled`. Recurrence policy belongs to its workflow. |
| Safe transient failure | Save failure/checkpoint, finish attempt, enter `queued` with `retry_not_before`; retain the recovery budget. |
| Unexpected lease loss | Finish the expired attempt as interrupted, start recovery at its saved lease expiry if necessary, and recover only within the retained budget after operation reconciliation. |
| Recovered handler panic | Save `handler_panic`, finish the attempt, invalidate the lease, and fail with outcome/delivery unless a terminal outcome or elapsed deadline takes precedence. |
| Eligible active task is paused | Record prior state, freeze recovery budget, invalidate lease, finish any active attempt, enter `paused`. |
| Explicit resume | Restore pending input/review or scheduled state as appropriate; otherwise queue work. Reissue unanswered input under a new request ID. |
| Owner stops task | Enter `cancelled`, invalidate lease and pending input/reviews, and save a cancellation outcome. |
| Success, definitive failure, retry exhaustion, or task expiry | Enter `succeeded`, `failed`, or `expired`, clear lease, and save outcome plus chat delivery intent atomically. |
| Explicit restart of failed/expired task | Create the next execution with a fresh retry budget, the last compatible checkpoint, and prior operation records. |

Pause/cancel/answer/finish/restart races lock the task and its current run in a consistent order so resolving `current_run` cannot race a restart. Every new owner control includes `expectedRunNumber` and `expectedVersion`. Compare both under those locks; either mismatch returns 409 without applying the requested change. Increment the run's control version for lifecycle transitions, accepted answers/decisions, and changes to actionable input/review requests, but not for routine heartbeats or progress-only checkpoint writes. Worker writes remain independently fenced by execution and lease. If completion commits first, a cancellation carrying its earlier version returns 409 with the authorized current state; if cancellation commits first, stale completion is rejected. A cancellation submitted against an already terminal execution with the matching pair returns that state without another outcome. `succeeded` and `cancelled` executions cannot be resumed or restarted through the restart endpoint; a new user request can create a new task. Repeating a committed control request returns its saved response before current-version checks.

**Alternative considered:** version alone. Since each execution has its own version, matching only that number permits a delayed request from an earlier execution to mutate a later one. The execution/version pair preserves the existing schema without a second task-wide version counter.

Optional task `expires_at` is checked inside mutation transactions and by a small expiration sweep for all nonterminal states, including paused and waiting tasks. Task expiry is distinct from lease expiry and review expiry. Restarting an expired task requires a new future expiry or explicitly clearing it; a fresh review is required for any expired permission. Terminal messages from earlier executions remain readable after restart.

### 3. Reuse document identities; version execution checkpoints separately

A checkpoint has `checkpoint_version: 1`, a task type, a handler version, a current step, completed steps, remaining work, accepted input request IDs, file IDs, evidence references, and operation keys. Use bounded JSON metadata; store document bytes through the existing document/workspace services. Validate all referenced files against the task's permanent workspace before saving or reading them. Read registered PDFs by `fileId`, never by a mutable path or an expired preview URL.

Accepted-input references are service-managed checkpoint fields. Answer acceptance locks the current execution and pending request, validates the expected execution/version, schema, ownership, and deadlines, then atomically writes the immutable answered row, adds its ID to the checkpoint, queues continuation, increments the control version, and saves the mutation response. A rollback saves none of those changes. Worker checkpoint writes must preserve committed accepted-input references rather than replacing them from an older snapshot. On explicit restart, copy these references into the new compatible checkpoint and resolve the original answered rows by workspace and task, even when their `run_number` belongs to an earlier execution. Never reopen or move those answered requests into the new run.

Handlers register supported task types and checkpoint/handler versions. They receive the saved checkpoint, resolved workspace, and lease-bound execution methods. They return progress, a wait request, a due time, or an outcome. They call the shared document service directly, without loopback HTTP or reusing an expired user session. Existing user sign-out does not cancel delegated work.

An older compatible checkpoint resumes through its registered decoder. Unsupported checkpoint or handler versions fail visibly with `checkpoint_version_unsupported` before handler effects. A future release must supply a reader or explicit migration; changing `/v1/api` alone does not migrate stored state. Manual restart cannot bypass this check.

Persist the checkpoint and its state transition together. If an operation finishes after the preceding checkpoint, recovery looks up its stable key and adopts its stored result. The file service already deduplicates identical published bytes/path; the task operation also records the resulting `fileId`. Nondeterministic document transformations need their workflow's own stable publication identity before they are declared retry-safe.

**Alternative considered:** using the PDF registration handler to store all checkpoints. It only accepts PDF files and cannot atomically change task eligibility. Retaining file IDs in database checkpoints reuses its existing responsibility without mixing document content with scheduler state.

### 4. Fence writes and serialize authorization for effects

Claim in one database transaction: lock an eligible current run, account for any unexpected expired lease, recheck version/status/due/expiry/recovery/backoff gates, finish the old attempt as interrupted, issue a unique lease token, and insert the new attempt. If a deadline has elapsed, commit the terminal outcome and delivery intent without inserting another attempt. PostgreSQL conditional updates remain the source of truth even if two processes select the same candidate. Every checkpoint, heartbeat, finish, and operation-start transaction checks the current task execution, running state, token, unexpired lease, task expiry, and recovery deadline.

Use one shared deadline guard for those transactions and for owner controls, answer acceptance, review continuation, and entry into paused/input/review/scheduled states. After locking, sample fresh database time and account for a previously undiscovered unexpected lease loss before deciding whether work can continue. At `now >= expires_at`, terminate as `expired`; otherwise at `now >= retry_deadline`, terminate as `failed` with `recovery_exhausted`. Task expiry wins when both apply. Clear the lease, finish any open attempt, and commit outcome plus delivery together. Do not roll back that terminal transition when returning a conflict for the requested mutation. An already terminal outcome remains immutable, and a valid explicit restart creates a new run subject to its replacement expiry. Sweeps discover idle expired work; they are not the enforcement boundary. A pause or wait cannot freeze a budget that has already expired. Trusted receipt reconciliation is allowed after termination solely to retain evidence.

Effects use a stable operation key plus a fingerprint of the task, step, and normalized inputs. A key with different inputs is a conflict. A confirmed result is reused. An uncertain result is reconciled before any new dispatch; it is never automatically resent.

The operation state `authorized` means dispatch may already have occurred. If its worker disappears before saving a receipt, the next worker reconciles it even if the saved state never reached `uncertain`. Confirmed success stores/reuses the receipt. Authoritative proof that the old dispatch did not execute and cannot later execute permits fresh authorization under the current lease, deadlines, and review. A missing receipt or inconclusive lookup is not such proof: save `uncertain`, fail an eligible active execution with an explanatory outcome, and do not resend. The adapter's proof must also exclude dispatch by the old authorization; a stale worker's request cannot remain eligible after that proof. Manual restart follows the same rules. Keep this behavior in the controlled adapter contract before real connectors arrive.

Define a review adapter with prepare/read-decision, authorize-dispatch, read-receipt, and invalidate-pending operations. Decisions include workspace, task/run, review generation, review ID, payload fingerprint, connection generation, expiry, and a stable decision ID. The connector owns real review storage and receipts. Controlled durable fixtures implement this contract for task acceptance until `build-connectors` lands.

Before calling prepare, commit a stable preparation intent in `task_operations` with the intended run and review generation. Use `(workspace_id, task_id, operation_key, review_generation)` as its idempotency identity; the same identity and payload returns the same review after a lost response, while a payload mismatch conflicts. A fresh review for expiry or changed authorization uses a new generation committed once under the task lock. Keep old reviews and decisions in the ledger; the task row holds only the current binding and consumed decision ID. Recover an interrupted prepare using its saved identity, then atomically save the checkpoint, review binding, and waiting transition while still eligible. If eligibility was lost, retain the binding for invalidation rather than queuing work.

Run a bounded review reconciliation pass on the worker's polling cadence, independently of handler execution slots. It reads ledger decisions and expiry for saved waiting bindings; callbacks merely request an earlier pass. Read the adapter outside task locks, then recheck the current run, review generation/ID, state, and deadlines inside the continuation transaction. Save the consumed decision ID and queue continuation together, restoring any suspended recovery deadline. Two passes, a lost notification, an early decision before waiting commits, or a crash between ledger read and local commit must still produce one continuation. Paused tasks cannot be queued; explicit resume consults the ledger and revalidates the saved review. Expiry without a callback reaches the handler for fresh review or a visible terminal response, never permission to dispatch.

**Alternative considered:** relying on review callbacks alone. Durable ledger reconciliation also covers decisions saved before waiting state and lost notifications, without copying the approval ledger into the task service.

The dispatch authorization transaction serializes with pause/cancel, checks the active lease and review, and durably records the one operation allowed to execute. A worker must check again at the adapter boundary before calling the provider. There is still an unavoidable gap between database authorization and a remote provider accepting a request: cancellation stops pending work, but cannot promise retraction of already-authorized/in-flight requests. Reconciliation records a late receipt without changing a cancelled task back to succeeded.

Invalidating task eligibility immediately prevents further authorization. Closing a review in a later service is retried using the stable operation record; a stale outstanding review can never override the task eligibility check. Missing connector capabilities fail closed. A provider disconnect after dispatch records uncertainty and ends automatic execution with an explanatory failure outcome.

### 5. Bound automatic recovery to two minutes

Transient errors classified by a handler as safe to retry and unexpected lease loss enter recovery. Invalid input, unsupported versions, denied permission, recovered handler panics, and uncertain external effects do not enter an automatic retry loop. Examples of retryable failures include unavailable workspace storage or a read failing before any effect. Worker loss never establishes that a prior effect was safe to repeat; reconcile operations first.

On the first retryable failure, persist `recovery_started_at` and a deadline 120 seconds later. For the first unexpected lease loss, use that lease's saved `lease_expires_at` as the start, not the time a replacement worker discovers it. Discovery after the resulting deadline fails the run without invoking a handler. Subsequent failures and lease losses retain the same budget. Retry delays start at one second and double up to 15 seconds, bounded by remaining time. Once recovery starts, its budget covers active work and retry backoff until completion; transaction guards, handler contexts, and a deadline sweep enforce exhaustion even if a handler stalls. Initial healthy execution has no two-minute task-duration cap. A null `recovery_started_at` distinguishes a budget never started from one temporarily suspended.

On explicit pause, input/review wait, or a future scheduled wait, first enforce deadlines, then atomically save the remaining milliseconds and clear the active deadline while retaining `recovery_started_at`. Restore the deadline when resume, answer acceptance, or review consumption queues continuation, not when a worker eventually claims it. If resume remains in a genuine waiting state, keep the budget suspended. For a future scheduled wait, exclusion ends at the saved due instant; derive the restored deadline as that instant plus the saved remainder even if eligibility is discovered later. Offline time during active recovery and queued retry backoff still counts. A state transition cannot reset the budget to 120 seconds. An explicit new execution after failure or expiry starts with a fresh, not-yet-active budget.

At exhaustion, invalidate the lease and persist `failed`, error code `recovery_exhausted`, a safe message such as “The task could not finish after two minutes of recovery. You can start it again,” and the chat delivery intent. Message delivery retries are independent: unavailable chat publication must not change a successful task into a failed task.

**Alternative considered:** starting the budget only after a handler returns an error or a replacement worker returns online. That permits repeated crashes or long outages to extend automatic recovery indefinitely. Anchoring unexpected loss to lease expiry bounds recovery, at the cost of requiring Start again after a sufficiently long outage.

### 6. Make owner controls and API responses replayable

Require an `Idempotency-Key` on task creation and control mutations. Scope it by workspace and operation route/resource, normalized without the API prefix. Hash the canonical validated payload; include `expectedRunNumber`, `expectedVersion`, and input request ID where applicable. In one transaction, reserve the unique key, perform the mutation, and save the HTTP status and response body. A concurrent duplicate waits for that transaction and reads the saved response. A different payload with the same key returns 409. Rolled-back requests have no saved success.

Authenticate and authorize every replay before returning stored data. Resolve a matching committed response before evaluating current execution/version or state preconditions, so an old successful response remains replayable without mutating a newer execution. Do not store credentials in these response records. Retain successful task response records for the lifetime of the task; transient infrastructure errors before commit can be retried using the same key. Input rejected because a task is paused/stopped stores no answer payload in the task or response body. Its safe error response can still be replayed.

The UI submits Stop task for a particular task ID. It does not infer the task from the most recent message or cancel all tasks in a chat. With several tasks, show the selected task's name beside its control. Preserve the unrelated draft. Pause drops unresolved request IDs; late answers to them are rejected even after resume. Previously accepted answers and completed steps remain. Resume creates a new request ID for still-missing input.

Keep the original key and exact payload, including execution/version, while retrying an ambiguous network response. Do not refresh those preconditions and silently retarget the pending control at a newer execution. A 409 prompts a state refresh and an explicit new user action; a replayed old response is associated with its returned execution and cannot mark the current execution stopped.

A review decision arriving while paused can remain in its owning ledger, but cannot queue the task. On resume, revalidate it against expiry, connection, and payload; an invalid review requires a new one. This differs from a newly submitted ordinary input answer, which is discarded while paused.

### 7. Persist destination and publish results through a local outbox

Create one durable Main conversation per workspace using a partial unique index. An authenticated client can create a minimal side conversation and list its owned conversations; full history operations follow in `map-conversations`. Source mail threads and calendar events remain separate identities, with optional source bindings validated against the workspace. Existing local UI side-chat IDs must be registered before delegation. An explicit unknown or foreign conversation ID is rejected, never silently redirected to Main.

On task creation, resolve and store `conversation_id`; missing origin resolves Main in the same transaction. Keep this destination immutable for the task. A source file or mail thread does not substitute for a conversation ID.

Finish a run by atomically committing the terminal state, immutable outcome, and a delivery intent carrying a preallocated message ID. A dispatcher inserts the task message and marks delivery published in one PostgreSQL transaction. The unique outcome/message relationship makes retry after a lost acknowledgement harmless. No live browser connection is required. Delivery attempts use bounded backoff and remain pending through restart; this queue has no two-minute execution budget.

```text
Handler finishes
      |
      v
PostgreSQL transaction
  run outcome + result/file references + pending chat delivery
      |
      v
Dispatcher transaction
  insert stable message ID + mark delivery published
      |
      v
Originating chat (or saved Main destination)
  reopen -> read persisted message -> open same task/result
```

The initial dashboard fetches task messages when opening a conversation and polls while it is visible, also refreshing on focus. Merge by message ID and order by `(created_at, id)`. Pagination uses that same pair. Task state and conversation messages are read separately: a completed task can show “Posting result…” while delivery is pending. Tasks and Activity point to the outcome/message, not a second result copy.

When Rich Threads arrives, adopt these conversation IDs or maintain an explicit permanent mapping. Preserve the task message ID as the deduplication key when projecting into external history. The outbox remains authoritative until that projection acknowledges publication. Archiving a conversation later must not reroute its tasks to Main or delete their results.

**Alternative considered:** sending only through a live chat stream. That would lose completions while the app is closed and cannot resolve a lost acknowledgement. Local persisted task messages satisfy this milestone without waiting for the full conversation runtime.

### 8. Scheduling uses a saved instant and explicit expiry

Store an absolute due instant plus its IANA time-zone label. For the workshop reminder, the caller resolves the event's 10:00 start and subtracts 30 minutes before submitting 09:30 with an unambiguous offset/zone. Reject ambiguous local times and inconsistent zone/offset combinations. General natural-language interpretation belongs to the runtime.

Do not claim before `due_at`. As agreed during exploration, an overdue one-shot reminder runs once when workers return, unless its optional task expiry has passed. Keep its saved due time fixed when the source event changes. Completion is a persisted reminder message in the destination chat, with the same task ID; no separate email/push channel is implied. A restarted expired reminder becomes immediately eligible unless the owner supplies a new future due time, and must supply a valid new expiry or clear it.

Use a separate `retry_not_before` for backoff so retries do not overwrite the promised reminder time. “Today's activity” uses the selected display zone for day boundaries; stored timestamps stay absolute. Later watch handlers may return another due instant, but recurring reminder rules and event-change subscriptions are outside this change.

### 9. API and package boundaries

Mount the feature mux at canonical `/v1/api/`, stripping `/v1/api`, and retain `/api/` as an alias to the same handlers. Register feature routes relative to the prefix, as today. Update session-exchange protection, generated preview URLs, OpenAPI, dashboard fetch configuration, dev proxy, and deployment health checks together. Both prefixes receive the same origin checks, rate limits, authentication, ownership checks, and idempotency scope. Already-issued `/api/previews/...` links remain valid until their original expiry.

For the configured authorized origin, include `Idempotency-Key` with `Authorization` and `Content-Type` in CORS preflight responses on both mounts. Keep the existing origin allowlist. Exercise actual cross-origin browser requests because a same-origin proxy or direct HTTP test will not expose a blocked custom header.

| Method and relative route | Purpose |
| --- | --- |
| `POST /tasks` | Create immediate/scheduled task; return committed ID, destination, state, run number, version, and due time. |
| `GET /tasks` | Cursor-paginated owned tasks, filterable by conversation/state and activity time range. |
| `GET /tasks/{taskId}` | Current run, attempts, progress, pending input/review, outcomes and delivery status. |
| `POST /tasks/{taskId}/answers` | Answer a current request with its request ID, `expectedRunNumber`, and `expectedVersion`. |
| `POST /tasks/{taskId}/pause`, `/resume`, `/cancel`, `/restart` | Owner controls checked against `expectedRunNumber` and `expectedVersion`, with replayable responses. |
| `POST /conversations` | Create a minimal owned side conversation using a replayable request key. |
| `GET /conversations` | Ensure/list Main and minimal owned side conversations with stable IDs. |
| `GET /conversations/{conversationId}/task-messages` | Cursor-paginated persisted task messages and outcome references. |

Use 400 for invalid payloads or missing execution/version preconditions, 401 for invalid sessions, 404 for missing/foreign IDs, 409 for stale executions/versions, elapsed execution deadlines, closed input requests or idempotency conflicts, and 503 for unavailable dependencies. Include current authorized state and execution/version in control conflicts so clients can refresh without guessing. Limit request/checkpoint JSON to 256 KiB and list pages to 100 records; document exact schemas and limits in OpenAPI. Internal review decision delivery is a service interface, not an unauthenticated public endpoint.

Private task code belongs in `internal/task/`: `task.go`, `handler.go`, `service.go`, `memory.go`, and `config.go`, plus worker/handler-contract files when needed. Minimal conversation identity and task-message reads belong in `internal/conversation/` with the same feature layout. Put routes in `cmd/api/routes/tasks.go` and `conversations.go`. All Go tests go in the closest `test/` directory. Keep task outcome/outbox writes transactionally owned by the task service; expose a narrow transactional append operation for the conversation message table.

### 10. Database schema

These are new tables; existing workspaces and files retain their IDs. Every `workspace_id` below references `workspaces.id`. Composite constraints enforce workspace ownership for relationships; JSON file/evidence references also undergo service-level ownership validation. All listed JSON values are bounded objects/arrays with explicit versioned schemas, not executable serialized state.

Conversation identity:

```text
conversations:
  id                  UUID          PK, NOT NULL
  workspace_id        UUID          NOT NULL, FK -> workspaces.id
  kind                TEXT          NOT NULL, CHECK IN (main, side)
  title               TEXT          NOT NULL
  source_kind         TEXT          NULL, CHECK IN (mail_thread, calendar_event)
  source_id           UUID          NULL
  created_at          TIMESTAMPTZ   NOT NULL
  updated_at          TIMESTAMPTZ   NOT NULL
  UNIQUE (workspace_id, id)
  UNIQUE (workspace_id) WHERE kind = 'main'
  CHECK source_kind and source_id are both NULL or both non-NULL
```

Task identity and executions:

```text
tasks:
  id                  UUID          PK, NOT NULL
  workspace_id        UUID          NOT NULL, FK -> workspaces.id
  conversation_id     UUID          NOT NULL
  task_type           TEXT          NOT NULL
  request             JSONB         NOT NULL
  initial_plan        JSONB         NOT NULL
  current_run         INTEGER       NOT NULL, CHECK > 0
  created_at          TIMESTAMPTZ   NOT NULL
  updated_at          TIMESTAMPTZ   NOT NULL
  UNIQUE (workspace_id, id)
  FK (workspace_id, conversation_id) -> conversations(workspace_id, id)
  FK (workspace_id, id, current_run) -> task_runs(workspace_id, task_id, run_number), DEFERRABLE INITIALLY DEFERRED

task_runs:
  workspace_id        UUID          NOT NULL, FK -> workspaces.id
  task_id             UUID          NOT NULL
  run_number          INTEGER       NOT NULL, CHECK > 0
  state               TEXT          NOT NULL
  version             BIGINT        NOT NULL, CHECK >= 0
  resume_state        TEXT          NULL
  handler_version     INTEGER       NOT NULL, CHECK > 0
  checkpoint_version  INTEGER       NOT NULL, CHECK > 0
  checkpoint          JSONB         NOT NULL
  progress            JSONB         NOT NULL
  due_at              TIMESTAMPTZ   NULL
  time_zone           TEXT          NULL
  expires_at          TIMESTAMPTZ   NULL
  recovery_started_at TIMESTAMPTZ   NULL
  retry_remaining_ms  INTEGER       NOT NULL, CHECK BETWEEN 0 AND 120000
  retry_deadline      TIMESTAMPTZ   NULL
  retry_not_before    TIMESTAMPTZ   NULL
  retry_count         INTEGER       NOT NULL, CHECK >= 0
  lease_token         UUID          NULL
  lease_expires_at    TIMESTAMPTZ   NULL
  created_at          TIMESTAMPTZ   NOT NULL
  updated_at          TIMESTAMPTZ   NOT NULL
  finished_at         TIMESTAMPTZ   NULL
  PK (workspace_id, task_id, run_number)
  FK (workspace_id, task_id) -> tasks(workspace_id, id)
  CHECK state IN (queued, scheduled, running, waiting_input, waiting_review, paused, succeeded, failed, cancelled, expired)
  CHECK resume_state IS NOT NULL only when state = 'paused'
  CHECK lease_token and lease_expires_at are non-NULL exactly when state = 'running'
  CHECK state = 'scheduled' requires due_at and time_zone
  CHECK retry_deadline IS NULL or recovery_started_at IS NOT NULL

task_attempts:
  id                  UUID          PK, NOT NULL
  workspace_id        UUID          NOT NULL, FK -> workspaces.id
  task_id             UUID          NOT NULL
  run_number          INTEGER       NOT NULL
  lease_token         UUID          UNIQUE, NOT NULL
  started_at          TIMESTAMPTZ   NOT NULL
  finished_at         TIMESTAMPTZ   NULL
  finish_reason       TEXT          NULL
  checkpoint_version  INTEGER       NOT NULL
  checkpoint          JSONB         NOT NULL
  progress            JSONB         NOT NULL
  error               JSONB         NULL
  FK (workspace_id, task_id, run_number) -> task_runs(workspace_id, task_id, run_number)
  UNIQUE (workspace_id, task_id, run_number) WHERE finished_at IS NULL
```

Inputs and operation reconciliation:

```text
task_inputs:
  id                  UUID          PK, NOT NULL
  workspace_id        UUID          NOT NULL, FK -> workspaces.id
  task_id             UUID          NOT NULL
  run_number          INTEGER       NOT NULL
  state               TEXT          NOT NULL, CHECK IN (pending, answered, closed)
  prompt              JSONB         NOT NULL
  answer_schema       JSONB         NOT NULL
  answer              JSONB         NULL
  created_at          TIMESTAMPTZ   NOT NULL
  updated_at          TIMESTAMPTZ   NOT NULL
  answered_at         TIMESTAMPTZ   NULL
  FK (workspace_id, task_id, run_number) -> task_runs(workspace_id, task_id, run_number)
  CHECK answer and answered_at are non-NULL exactly when state = 'answered'

task_operations:
  workspace_id        UUID          NOT NULL, FK -> workspaces.id
  task_id             UUID          NOT NULL
  operation_key       TEXT          NOT NULL
  input_fingerprint   BYTEA         NOT NULL
  kind                TEXT          NOT NULL
  state               TEXT          NOT NULL
  dispatch_run        INTEGER       NULL
  dispatch_lease      UUID          NULL
  review_run          INTEGER       NULL
  review_generation   INTEGER       NOT NULL, CHECK >= 0
  review_id           TEXT          NULL
  review_decision_id  TEXT          NULL
  result              JSONB         NULL
  receipt             JSONB         NULL
  last_error          JSONB         NULL
  created_at          TIMESTAMPTZ   NOT NULL
  updated_at          TIMESTAMPTZ   NOT NULL
  PK (workspace_id, task_id, operation_key)
  FK (workspace_id, task_id) -> tasks(workspace_id, id)
  FK (workspace_id, task_id, review_run) -> task_runs(workspace_id, task_id, run_number)
  CHECK state IN (prepared, waiting_review, authorized, succeeded, failed, uncertain, invalidated)
  CHECK review_run IS NOT NULL exactly when review_generation > 0
  CHECK review_id IS NULL or review_run IS NOT NULL
```

Outcomes and chat publication:

```text
task_outcomes:
  id                  UUID          PK, NOT NULL
  workspace_id        UUID          NOT NULL, FK -> workspaces.id
  task_id             UUID          NOT NULL
  run_number          INTEGER       NOT NULL
  status              TEXT          NOT NULL, CHECK IN (succeeded, failed, cancelled, expired)
  message             TEXT          NOT NULL
  result              JSONB         NOT NULL
  error_code          TEXT          NULL
  created_at          TIMESTAMPTZ   NOT NULL
  UNIQUE (workspace_id, id)
  UNIQUE (workspace_id, task_id, run_number)
  FK (workspace_id, task_id, run_number) -> task_runs(workspace_id, task_id, run_number)

task_deliveries:
  id                  UUID          PK, NOT NULL, also the stable chat message ID
  workspace_id        UUID          NOT NULL, FK -> workspaces.id
  outcome_id          UUID          NOT NULL
  conversation_id     UUID          NOT NULL
  state               TEXT          NOT NULL, CHECK IN (pending, published)
  attempt_count       INTEGER       NOT NULL, CHECK >= 0
  next_attempt_at     TIMESTAMPTZ   NOT NULL
  last_error          JSONB         NULL
  created_at          TIMESTAMPTZ   NOT NULL
  published_at        TIMESTAMPTZ   NULL
  UNIQUE (workspace_id, outcome_id)
  UNIQUE (workspace_id, id, conversation_id, outcome_id)
  FK (workspace_id, outcome_id) -> task_outcomes(workspace_id, id)
  FK (workspace_id, conversation_id) -> conversations(workspace_id, id)

conversation_task_messages:
  id                  UUID          PK, NOT NULL
  workspace_id        UUID          NOT NULL, FK -> workspaces.id
  conversation_id     UUID          NOT NULL
  outcome_id          UUID          NOT NULL
  created_at          TIMESTAMPTZ   NOT NULL
  UNIQUE (workspace_id, outcome_id)
  FK (workspace_id, id, conversation_id, outcome_id) -> task_deliveries(workspace_id, id, conversation_id, outcome_id)
```

Replayable mutation responses:

```text
mutation_responses:
  workspace_id        UUID          NOT NULL, FK -> workspaces.id
  operation_scope     TEXT          NOT NULL
  request_key         TEXT          NOT NULL
  request_fingerprint BYTEA         NOT NULL
  http_status         INTEGER       NOT NULL
  response            JSONB         NOT NULL
  created_at          TIMESTAMPTZ   NOT NULL
  PK (workspace_id, operation_scope, request_key)
```

Insert task and first run in one transaction using the deferred current-run foreign key. A task outcome and delivery destination are derived under the task lock; callers cannot supply a different workspace or chat. Operation results and outcome JSON contain stable file/evidence IDs, never attachment bytes or temporary preview URLs. Input schemas must be declarative and bounded; validation does not execute caller code.

Initialize `recovery_started_at` to NULL, `retry_remaining_ms` to 120000, and `retry_count` to zero for each new execution. Set recovery start once, including when discovering an expired lease, and retain it through suspended waits. `task_inputs` rows become immutable once answered; checkpoint references to them must resolve to answered rows in the same workspace and task, without requiring the new execution's run number. Initialize operation review generation to zero; the first preparation reserves generation 1 and its intended run before adapter I/O. Persist later generations only when a fresh review is required. Consuming a review decision updates its operation binding and task continuation in the same transaction; no separate ledger table is introduced here.

Candidate indexes, in addition to primary/unique constraints:

- `task_runs(state, due_at, retry_not_before)` on queued/scheduled rows: eligible work selection.
- `task_runs(lease_expires_at)` on running rows: interrupted worker recovery.
- `task_runs(expires_at)` on nonterminal rows: task expiration sweep.
- `task_runs(retry_deadline)` where non-NULL: enforce the two-minute recovery deadline.
- `task_runs(updated_at, task_id)` on waiting-review rows: bounded periodic decision/expiry reconciliation independent of callbacks.
- `tasks(workspace_id, conversation_id, created_at, id)`: conversation-scoped task lists.
- `task_outcomes(workspace_id, created_at, id)`: activity and terminal history pagination.
- `task_attempts(workspace_id, task_id, run_number, started_at, id)`: ordered attempt history.
- `task_inputs(workspace_id, task_id, run_number)` on pending rows: outstanding prompts.
- `task_deliveries(next_attempt_at, id)` on pending rows: delivery retries.
- `conversation_task_messages(workspace_id, conversation_id, created_at, id)`: replay and pagination.

### 11. User interface

Follow [DESIGN.md](../../../DESIGN.md), reuse the existing DraftComposer, Button, turn and attachment-preview primitives, and use the uncodixfy skill when implementing frontend code. Add task controls within the composer area without changing the surrounding pane layout. The selected task is explicit, keyboard reachable, and has an accessible Stop task label; show request progress and errors without declaring it stopped before the server commits. Task detail supplies Pause, Resume, and eligible Start again actions.

```text
+----------------+-------------------------------------+-----------------------+
| Chats          | Workshop documents                  | Tasks                 |
+----------------+-------------------------------------+-----------------------+
| Main chat      | Cueia                               | Today's activity      |
| > Workshop     | Registration form is ready.         | Registration form     |
|                | [Open registration.pdf] [View task] | Completed             |
|                |                                     |                       |
|                | Preparing workshop summary...       | Workshop summary      |
|                |                                     | Running               |
|                +-------------------------------------+                       |
|                | [+] Add a message...                |                       |
|                | Workshop summary       [Stop task]  |                       |
+----------------+-------------------------------------+-----------------------+
```

The completion message belongs to Workshop in this example. Main receives only tasks created without another destination. Messages remain available across sessions and obtain fresh preview links when files are opened. At narrow widths keep the stop control usable with the existing collapsible panes and preserve the draft when opening Tasks or switching conversations. The unauthenticated visual sample remains a fixture; authenticated acceptance exercises real persistence.

## Risks / Trade-offs

- A stale process can outlive its lease -> fence every durable write and effect authorization, cancel local execution on renewal failure, and reconcile remote receipts.
- A remote provider accepts a request during cancellation -> show the actual receipt/uncertainty; do not claim cancellation reverses accepted effects.
- A retryable handler stalls -> apply the persisted recovery deadline to its context and let an independent sweep close the run; reject its late completion.
- A control arrives after explicit restart -> check the execution/version pair and never retarget a network retry at the new execution.
- An outage lasts beyond the first expired lease's recovery window -> fail visibly without invoking another handler and retain Start again; graceful release does not start a budget.
- A dispatch receipt or review notification is missing -> reconcile authorized operations conservatively and poll the owning review ledger; never infer permission to resend from absence of evidence.
- Conversation work overlaps a later proposal -> introduce only stable destinations and task messages, and require later Rich Threads integration to adopt their identities.
- The local document store cannot make arbitrary external writes idempotent -> accept only registered handler types and require explicit effect/reconciliation contracts.
- A migration or rollback leaves queued tasks with unsupported code -> stop claims during deployment, retain checkpoints/outcomes, and fail incompatibility visibly instead of restarting work from scratch.
- Delivery backlog grows while publication fails -> retain pending intents, back off, expose pending delivery and diagnostics, and monitor oldest pending age without discarding results.

## Migration Plan

1. Add an additive versioned SQL migration for the tables, constraints, and indexes above. Keep bootstrap tables and test fixtures intact. Validate fresh install and upgrade with existing workspaces and documents.
2. Add canonical `/v1/api` mounting and compatible `/api` handling, including session protection, `Idempotency-Key` CORS support, and preview URLs. Update execution/version preconditions in the contract, generated types, dashboard/proxies, and deployment checks together.
3. Deploy the API and worker behind the existing migration startup gate. Initialize Main conversations lazily and atomically; register side conversations before task creation. Initialize the worker's workspace provider using the same configuration as the API.
4. Enable registered controlled handlers/reminders and dashboard task/result reads. Live model and connector handlers integrate later through the documented interfaces.
5. On rollback, stop task claims and delivery dispatch before running older binaries. Retain the additive schema, messages, and operations. Older bootstrap workers only see their claim fixture table. Do not delete checkpoints or automatically replay tasks. A subsequent compatible deployment resumes them.

## Open Questions

- Follow-up scheduling work must define recurring reminders and whether future event edits can reschedule an existing reminder. This version uses the agreed one-shot catch-up behavior and a fixed saved due time.
- Live connector receipt lookup and Rich Threads projection require acceptance when those services land. They do not block this change's PostgreSQL-backed messages and controlled review/handler tests.

## Validation Strategy

Use the numbered scenarios in the [durable task spec](specs/durable-tasks/spec.md), [chat-result spec](specs/task-chat-results/spec.md), and [access delta](specs/workspace-access/spec.md). Run race tests against Docker PostgreSQL with separate Go processes and explicit barriers, not timing sleeps. Stop at claim, answer commit, checkpoint, review preparation/decision, effect authorization, provider acceptance, outcome commit, and message publication boundaries. Use a shared controlled clock for due times, leases, retry exhaustion, and expiry. Hold sweeps and locks independently to verify post-lock deadline checks; delay an old control until a later execution has the same version. Kill workers repeatedly before any reported failure, distinguish graceful release from unexpected loss, and exercise recovered panics. Verify accepted answers survive before a new handler checkpoint, review decisions survive early/lost notification, and an authorized operation without a receipt cannot resend without authoritative proof.

Record original Cueia fixture and expected-result provenance. Exercise original PDF bytes across owner/path changes and host replacement, a lost mutation response, a stale answer after pause/resume, duplicate review decisions, an uncertain send, a stalled retry, and a lost publication acknowledgement. Browser acceptance covers Main and side-chat replay, Stop task isolation across execution changes, fresh preview links, narrow-screen draft preservation, and real cross-origin preflight/retry across both API prefixes with unchanged keys and payloads. Regenerate/check API types and run the affected Go and browser suites plus root build/type-check/test tasks during implementation. This artifact-only change is validated with `openspec validate model-task-recovery --strict`.
