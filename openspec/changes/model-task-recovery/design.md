# Task execution and recovery design

## Context

Bootstrap's Go worker polls `claim_records`. Its version-checked update and separate-process PostgreSQL test already demonstrate that one claimant wins. The worker still needs to run handlers and recover their work. The API and document service already resolve immutable PDF versions by workspace and `fileId`. [Claim implementation](../../../apps/server/internal/claim/claim.go) · [Document service](../../../apps/server/internal/document/service.go)

The dashboard has Main chat, local side-chat IDs, a draft composer, and empty task sections. The first working client needs a reminder to finish with the tab closed and a result the owner can open from the same conversation or from Tasks.

## Goals / Non-Goals

Preserve task identity, accepted answers, document versions, and outcomes through worker or API restarts. Let the owner stop one execution and let safe work resume after a long outage. Keep `/v1/api` as the canonical URL with the existing `/api` clients supported.

This change runs a one-shot reminder and a read-only PDF handler. The runtime adds model execution. The connector change owns waiting for approval, reconciling decisions, and determining whether a provider accepted a write. Rich Threads and recurring watches belong to their respective proposals. [Connector scope](../build-connectors/proposal.md#review-ledger-and-receipts)

## Decisions

### Run safe handlers in the existing worker

Use PostgreSQL and the separate Go worker. The same database can decide both who owns the attempt and whether its result commits. Keep the bootstrap claim fixture for regression tests and apply its conditional-update pattern to real task runs. Use a connection pool and short transactions. Perform file I/O outside row locks.

The worker uses the API's workspace provider configuration and the task's permanent workspace ID. A controlled PDF reader checkpoints a registered `fileId`, reads that version through the document service, and returns the ID and a digest of the bytes. If the ID is missing, it can request one from the owner. The reminder finishes by saving its due-time message. These handlers only read inputs and commit task state. Reject unregistered task types at creation. A stored handler or checkpoint version the worker cannot read produces a visible incompatibility failure before calling the handler.

### Claim, checkpoint, and retry

A task keeps its workspace, conversation, and ID for life. An execution starts at run 1. **Start again** creates a new run after failure or expiry. Each worker acquisition creates an attempt within the current run, with its own start time and lease token. Checkpoints record completed and remaining steps, progress, accepted input IDs, evidence, and immutable file IDs. Check ownership of referenced files on save and read.

Start with four handler slots, a one-second poll, a 30-second lease, and renewal every 10 seconds. Validate positive settings and a renewal interval shorter than the lease. Claim only when a slot is free. In one transaction, lock the task and current run in that order, sample fresh database time, check eligibility, create the attempt, and assign its token. A heartbeat, progress write, wait, or completion must match the running execution and its unexpired token. Cancel local work on lease loss. Use the same lock order and clock for owner controls and expiry.

Each attempt has a fixed deadline of its original start plus two minutes. Bound the handler context and every lease grant or renewal by that instant. At equality the attempt has timed out. The normal poll also closes expired leases and expires idle tasks independently of free handler slots. An elapsed task expiry closes any nonterminal run first, including paused or waiting runs. Commit the expired outcome even when the triggering operation returns a conflict.

| What ended the attempt | What the worker saves |
| --- | --- |
| Transient error, timeout, or unexpected lease loss | Close the open attempt once and increment `failed_attempts` under the run lock. Queue after one second on the first failure and two seconds on the second. The third failure records a terminal outcome. |
| Wait for input, owner pause, or graceful shutdown while the lease is valid | Save the checkpoint and finish reason, release the lease, and preserve the failure count. Graceful shutdown returns work to the queue. |
| Permanent error, unsupported checkpoint, or caught handler panic | Finish the eligible run with a safe failure message. A panic releases the local slot so another task can run. |
| Handler success, owner cancellation, or task expiry | Finish the run and commit its immutable outcome in the same transaction. Close any open attempt and unresolved input request. |

Persist `retry_not_before` separately from the promised reminder time. Measure backoff from the error time or, for a lost lease, its saved expiry. On discovering an expired lease, account for that attempt before acting on the run. An hour offline can lose one running attempt. If retries remain and backoff has elapsed, a new worker can claim immediately. Tracking failed attempts removes the need to keep a recovery clock running or suspend it during user waits.

### Schedule and control the current execution

Save a future reminder's absolute due instant and IANA time zone before confirmation. Reject ambiguous or inconsistent local times. Only claim when both the promised time and any retry backoff have elapsed. The saved time survives source-event edits. An overdue, unexpired reminder runs once when workers return.

Creation and owner controls use `Idempotency-Key`. Controls on an existing execution also carry `expectedRunNumber` and `expectedVersion`. Authenticate and check workspace ownership, then resolve an already committed request before evaluating the current run. Scope keys by workspace, operation, and target independent of the URL prefix. Save the canonical payload fingerprint, status and response with the mutation. Concurrent matching requests read the same response. A changed payload or a stale execution/version on a new request returns 409. Increment the control version on lifecycle and actionable input changes. Heartbeats and progress alone preserve it.

| Owner action or handler request | Durable transition |
| --- | --- |
| Wait for input | Commit the checkpoint and one pending request, close the attempt, release the slot, and enter `waiting_input`. |
| Answer | Validate the current request and schema, commit the immutable answer and its checkpoint reference, queue continuation, and save the response together. Worker checkpoints preserve service-managed answer references. |
| Pause / Resume | Pause invalidates the lease, closes pending input, and remembers the prior state. Explicit resume reissues any needed input with a fresh request ID, restores a future schedule, or queues work with the saved backoff. |
| Stop task | Cancel the selected execution, invalidate its lease, close pending input, and commit its outcome. A competing completion can win only if it commits first. |
| Start again | For a failed or expired run, insert the next run with zero failed attempts and the last compatible checkpoint. Keep earlier attempts, answers, and outcomes. Require a future expiry or explicit removal of an elapsed expiry. |

### Use the outcome itself as the chat message

Resolve Main once per workspace, including across concurrent first requests. Register side conversations before delegation. Persist the chosen destination on the task. An omitted origin resolves to Main. Reject an explicit unknown or foreign ID. Mail thread and calendar event IDs remain source references.

In one PostgreSQL transaction, transition the run to its terminal state and insert one row in `task_outcomes`, unique per workspace, task, and run. Its primary key is the stable message ID. Losing the acknowledgement after commit leaves the outcome readable with every worker offline. A rollback leaves neither the terminal state nor its message. A separate delivery queue would add a second commit between the task and a chat read in the same database.

Serve chat messages by joining outcomes to tasks on workspace and task ID and filtering on the saved conversation. Order and paginate by `(created_at, id)` on the outcome. Tasks and Activity use that same outcome and destination. A new execution can add another outcome under the same task ID. Keep stable result/file IDs in the row and issue fresh authorized preview links when opened. Later Rich Threads history can reference the existing outcome ID.

### API and package boundaries

Mount one feature mux at canonical `/v1/api/` and compatible `/api/`, stripping the appropriate prefix. Share authentication, session exchange protection, rate limits, workspace authorization, CORS origin checks, and mutation replay across both mounts. Allow `Idempotency-Key` alongside `Authorization` and `Content-Type` in preflight for the authorized origin. Issue new preview links under `/v1/api` and honor existing signed `/api/previews/...` links through their original expiry.

| Relative route | Purpose |
| --- | --- |
| `POST /tasks` | Commit the new task, destination, initial run, any due time, and replayable response. |
| `GET /tasks`, `GET /tasks/{taskId}` | List owned tasks by conversation, state, or activity time. Inspect runs, attempts, pending input, and outcome IDs. |
| `POST /tasks/{taskId}/answers` | Submit an answer bound to the current request, execution, and version. |
| `POST /tasks/{taskId}/pause`, `/resume`, `/cancel`, `/restart` | Apply a replayable owner control to the selected execution. |
| `POST /conversations`, `GET /conversations` | Create an owned side conversation with a replayable key, or resolve Main and list owned destinations. |
| `GET /conversations/{conversationId}/task-messages` | Read paginated outcomes for the saved conversation. The response message ID is the outcome ID. |

Return 400 for invalid input or missing control preconditions, 401 for an invalid session, 404 for missing or foreign resources, 409 for conflicts, and 503 for unavailable dependencies. Include authorized current state and execution/version on a control conflict. Limit request/checkpoint JSON to 256 KiB and pages to 100 records. Describe the canonical routes and aliases in OpenAPI and regenerate `packages/api-types`.

Keep private task and conversation code under `apps/server/internal/task/` and `internal/conversation/` following the session layout. The task service owns terminal writes. The conversation service reads those outcomes. Register the relative paths in `cmd/api/routes/tasks.go` and `conversations.go`. Go tests belong in the closest `test/` directory.

### Database schema

Apply one additive migration to bootstrap. Composite foreign keys keep relationships inside the workspace. Validate file/evidence and accepted-answer references in bounded, versioned JSON at the service boundary. Completed answers remain readable by workspace and task across executions.

#### Conversation identity

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

#### Task identity and executions

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
  failed_attempts     INTEGER       NOT NULL, DEFAULT 0, CHECK BETWEEN 0 AND 3
  retry_not_before    TIMESTAMPTZ   NULL
  lease_token         UUID          NULL
  lease_expires_at    TIMESTAMPTZ   NULL
  created_at          TIMESTAMPTZ   NOT NULL
  updated_at          TIMESTAMPTZ   NOT NULL
  finished_at         TIMESTAMPTZ   NULL
  PK (workspace_id, task_id, run_number)
  FK (workspace_id, task_id) -> tasks(workspace_id, id)
  CHECK state IN (queued, scheduled, running, waiting_input, paused, succeeded, failed, cancelled, expired)
  CHECK resume_state IS NOT NULL only when state = 'paused'
  CHECK lease_token and lease_expires_at are non-NULL exactly when state = 'running'
  CHECK state = 'scheduled' requires due_at and time_zone

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

#### Task inputs

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
  UNIQUE (workspace_id, task_id, run_number) WHERE state = 'pending'
```

#### Terminal outcomes and chat messages

```text
task_outcomes:
  id                  UUID          PK, NOT NULL, also the stable chat message ID
  workspace_id        UUID          NOT NULL, FK -> workspaces.id
  task_id             UUID          NOT NULL
  run_number          INTEGER       NOT NULL
  status              TEXT          NOT NULL, CHECK IN (succeeded, failed, cancelled, expired)
  message             TEXT          NOT NULL
  result              JSONB         NOT NULL
  error_code          TEXT          NULL
  created_at          TIMESTAMPTZ   NOT NULL
  UNIQUE (workspace_id, task_id, run_number)
  FK (workspace_id, task_id, run_number) -> task_runs(workspace_id, task_id, run_number)
```

#### Replayable mutation responses

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

Insert the task and its first run in one transaction using the deferred current-run foreign key. A run owns the latest checkpoint. An attempt retains its original start time, last saved progress, and finish reason. Keep successful mutation responses for the life of the task and exclude credentials from their contents.

#### Candidate indexes

- `task_runs(state, due_at, retry_not_before)` on queued and scheduled runs selects eligible work.
- `task_runs(lease_expires_at)` on running rows locates lost or timed-out attempts. `task_runs(expires_at)` on nonterminal rows locates expired tasks, including idle waits.
- `tasks(workspace_id, conversation_id, created_at, id)` supports task lists and the join from chat to its tasks.
- `task_outcomes(workspace_id, created_at, id)` supports ordered Activity and message reads. The per-run unique key supports lookup from task detail.
- `task_attempts(workspace_id, task_id, run_number, started_at, id)` orders the attempt history.

### Dashboard

Use the existing composer, turn, button, and attachment-preview components. Keep the selected task's name and keyboard-accessible **Stop task** control together. Preserve the draft when a control runs or the owner switches chats. On an ambiguous response, retry the exact command with its original key. On a conflict, refresh the task and require a fresh user action to control another execution. Task detail exposes input, Pause, Resume, and eligible Start again.

```text
+----------------+-------------------------------------+-----------------------+
| Chats          | Workshop documents                  | Tasks                 |
+----------------+-------------------------------------+-----------------------+
| Main chat      | Cueia                               | Today's activity      |
| > Workshop     | Read the registration form.         | Registration form     |
|                | [Open registration.pdf] [View task] | Completed             |
|                |                                     |                       |
|                | Reading workshop notes.pdf...       | Workshop summary      |
|                |                                     | Running               |
|                +-------------------------------------+                       |
|                | [+] Add a message...                |                       |
|                | Workshop summary       [Stop task]  |                       |
+----------------+-------------------------------------+-----------------------+
```

Fetch the selected conversation's outcomes when it opens, while visible, and on focus. Merge by message ID. Opening a task's result navigates to its saved conversation. Scheduled tasks and Today's activity use the same persisted IDs and the owner's display zone. Keep the unauthenticated sample as a fixture. [Dashboard design](../../../DESIGN.md)

## Risks / Trade-offs

- A worker outlives its lease. Check the token and fresh database time on every durable worker write. The first two handlers are safe to repeat.
- A long outage delays a reminder. The saved due time and optional task expiry determine whether it can still run. The retry count bounds failed attempts across outages.
- A later workflow can change a provider resource. Enable that handler only with the connector-owned review and receipt checks. The first release can establish recovery using file reads and local outcomes.

## Migration Plan

1. Test the additive migration on a fresh database and a bootstrap workspace with registered documents. Gate API and worker startup on migration success.
2. Ship the shared mux, canonical URLs, compatible aliases, generated clients, dashboard configuration, development proxy, preview links, and deployment health checks together.
3. Start the worker with the same workspace access as the API. Resolve Main lazily and register side chats before task creation. Enable the reminder and PDF handlers.
4. To roll back, stop task claims and restore the bootstrap API, worker, clients, and health checks together. Retain the additive tables and checkpoints for the next compatible deployment.

## Open Questions

The watches proposal will decide recurring schedules and how event edits affect them. The one-shot reminder uses its original saved time.

## Validation Strategy

Use controlled database time and separate worker processes against Docker PostgreSQL. Exercise competing claims, lost leases and acknowledgements, the third failed attempt, task expiry under a lock wait, and Stop racing completion. Keep workers offline when checking that a committed outcome is already readable. Test the original PDF bytes after path edits from a second host and the correct Main or side-chat result after reopening from another session. Record Cueia fixture and expected-result provenance.

Use a cross-origin browser to retry a lost create through the other prefix and to stop one selected task at a narrow width. Check the Go and browser suites, generated contract, and root build, type-check, and test tasks during implementation. [Durable task requirements](specs/durable-tasks/spec.md) · [Chat result requirements](specs/task-chat-results/spec.md) · [Versioned access requirements](specs/workspace-access/spec.md)
