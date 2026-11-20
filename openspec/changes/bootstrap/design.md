# Bootstrap design

## Context

Cueia has planning documents and no application or database to migrate. The first deployment must work from a clean Cueia checkout. Two processes can race on first sign-in, and two different stores must agree before the owner can read a PDF. The foundation milestone and its four acceptance cases are in the [bootstrap proposal](proposal.md#acceptance-on-a-fresh-installation).

## Goals / Non-Goals

One verified Clerk subject maps to one durable workspace per deployment with its own cloud VM and persistent disk. Separate Cueia sessions can share that workspace and expire independently. The same Go storage code and SQL migrations serve the API, worker, and integration tests. The worker boundary in this change ends at a versioned claim. Scheduling, leases, retries, and task execution belong to [task recovery](../model-task-recovery/proposal.md).

## Decisions

### Repository layout

The planned Cueia workspace groups the TypeScript applications and packages alongside the Go module. Build and deployment commands run from the repository root.

```text
apps/
  dashboard/          # Vite React app with React Router v7 and the Spectrum provider
    package.json
    app/pages/
      sign-in/        # Clerk sign-in and return
      workspace/      # Main chat, Files sidebar, Tasks panel, and workspace preferences
  browser-worker/     # Bun process that runs Playwright
    package.json
packages/
  components/         # Reusable Adobe React Spectrum 2 UI, including the preview dialog
    package.json
  api-types/          # TypeScript types generated from backend/api/openapi.yaml
    package.json
backend/
  cmd/
    api/main.go       # Register routes and middleware, start and shut down HTTP
    worker/main.go    # Connect the worker to the shared claim operation
    migrate/main.go   # Run pending SQL migrations and exit
  internal/
    config/           # Parse and validate environment settings
    database/         # Open the pgx pool and run versioned migrations
    middleware/       # Authenticate private requests, check origins, rate-limit attempts
    session/          # Exchange Clerk credentials, issue and revoke Cueia sessions
    workspace/        # Provision the owner's VM, browse files, read snapshots, update preferences
    document/         # Publish attachment bytes, serve reads, sign and verify previews
    claim/            # Claim an eligible record at its expected version
  api/openapi.yaml    # HTTP contract used to generate TypeScript client types
  migrations/         # Versioned SQL files for both sample and live deployments
  go.mod
  go.sum
package.json          # Bun version, workspace membership, and root commands
bun.lock              # Dependency lockfile for all TypeScript workspaces
turbo.json            # Build, type-check, and test tasks
Dockerfile            # Multi-stage build for the three Go commands
compose.yaml          # PostgreSQL, migrations, API, worker, and workspace provider configuration
openspec/
  changes/            # Active proposals and archived changes
  specs/              # Accepted specifications
```

### Organize Go code under `backend/`

Keep one pinned Go module with three executable entry points. Each command loads the configuration it needs and wires the packages it uses.

Keep HTTP handlers, business logic, SQL queries, and domain types together in the feature that owns them. For example, `workspace/handler.go` handles snapshot and preference requests, `service.go` coordinates provisioning, `repository.go` holds workspace queries, and `model.go` holds workspace types. Use those files as the feature needs them. The API entry point registers the handlers on a `net/http` mux and owns the listener lifecycle.

### Build from the root and share the HTTP contract

Register `apps/*` and `packages/*` in the root Bun workspace, with one manifest per TypeScript workspace. The Vite dashboard in `apps/dashboard` depends on `packages/components` and `packages/api-types`. The Playwright process runs under Bun from `apps/browser-worker`. The Go worker that claims database records has its entry point in `backend/cmd/worker`.

Pin Bun in the root `package.json` and commit one `bun.lock`. Run build, type check, and test through root Turbo tasks in workspace dependency order. Run Go commands from `backend/` and declare Go sources, module files, migrations, and the API contract as inputs where applicable.

The [bootstrap API routes](proposal.md#bootstrap-api-routes) cover health and provider status, sessions, snapshots and collections, preferences, filesystem access, and document previews. Record the agreed methods, paths, requests, responses, and errors in `backend/api/openapi.yaml`. Generate the dashboard’s request, response, and error types into `packages/api-types`. The root checks verify that the generated package matches the contract. Go handlers validate requests at runtime.

### Build the dashboard with Adobe React Spectrum 2

Mount one `Provider` from `@react-spectrum/s2` inside the dashboard’s React Router v7 root. Connect it to the router’s navigation and link resolution so Spectrum links follow dashboard routes. Pages and shared components inherit the same theme. [Adobe React Spectrum setup](https://react-spectrum.adobe.com/getting-started)

Keep the sign-in and workspace pages and their page-specific controls under `apps/dashboard/app/pages/`. Put reusable Cueia UI in `packages/components`. For example, the workspace page requests a signed URL for the selected attachment and passes it to the shared Spectrum preview dialog. This keeps the session and API call in the dashboard and lets the dialog render a PDF wherever a page supplies an authorized URL.

The React Router v7 sample entry route starts Clerk sign-in when necessary and returns to the same route to finish the exchange and fetch the snapshot. Configure open signup in Clerk for each mode with email and password authentication and the Google and GitHub social providers. Only the publishable Clerk configuration enters the web build. Keep the Cueia bearer in client memory and exchange again using the existing Clerk session on a return visit. Explicit sign-out revokes the presenting Cueia bearer before clearing client state.

### Chat, Files, and Tasks

The desktop workspace uses a dark Spectrum theme, a slim navigation rail, and a central conversation with left-aligned agent replies, right-aligned user messages, and a composer at the bottom. The sketch shows the product once an agent can schedule a reminder for a 10:00 workshop.

```text
+-------+-----------------+----------------------------------------+-------------------------+
|       | [Search]        | Main chat                              | Cueia               [x] |
| Chat  | [Main chat]     |                                        | Connected               |
| Find  | Side chats  [+] |                                    You |                         |
| Files |                 | Remind me 30 minutes before Saturday's | Scheduled tasks (1)     |
| Ideas |                 |                              workshop. | Workshop reminder       |
| Tasks |                 |                                        | Sat 09:30               |
|       |                 | Cueia                                  | See all scheduled tasks |
|       |                 | Reminder set for Saturday at 09:30.    |                         |
|       |                 |                                        | Today                   |
|       |                 |                                        | Scheduled reminder      |
| Menu  |                 | [ +   Message...                 Mic ] |                         |
+-------+-----------------+----------------------------------------+-------------------------+
```

- **Chat.** Open Main chat on entry. In sample mode, show the workshop mail with its sender and the event with its time from the signed-in owner's snapshot. The owner can type a local draft. Mark sending unavailable until the agent runtime is connected. The client and conversation changes supply live turns, composer actions, and side-thread creation and history.
- **Files.** Opening Files replaces the thread list with the current workspace directory. Refresh reads the directory again, and filename search includes nested user and agent files. Show folders and all file types even when no preview is available. For a supported PDF, resolve its workspace-relative entry to a stable file ID and pass its signed URL to the preview dialog. Closing the dialog returns focus to the entry. Chat restores the thread list and preserves the draft and open Tasks panel.
- **Tasks.** The Tasks button opens a separate right panel. Show the agent's identity and connection state above Scheduled tasks and Today's activity. Closing this panel preserves the left sidebar and the conversation. Bootstrap displays the unavailable agent state and empty task sections. A seeded calendar event alone does not schedule a reminder. Once the runtime and scheduler are connected, the chat confirmation and the scheduled entry refer to the same saved task ID and due time. Activity comes from the owner's recorded actions.
- **Workspace controls.** Menu provides preferences and sign-out. Keep the conversation and composer in view as the sidebars open. On narrow screens, use dismissible panels that return the owner to the same chat. Call and Invite controls are excluded from the workspace.

### Keep deployment state and secrets separate by mode

Use separate Compose projects, PostgreSQL volumes, and deployment identities for sample and live. Scope cloud VM and disk allocations to the deployment and workspace. Both deployments run the same schema. Each person's VM mounts only that workspace's persistent disk for workspace files. Configure the owner and agent working directories on that disk so a file created from a terminal or a tool appears under the browsable workspace root. The API and worker reach it over a private authenticated provider connection. Keep provider credentials server-side. Publish the sample API on loopback and bind a directly run sample API to loopback.

Validate the Clerk verification configuration and `CPK_INTELLIGENCE_API_KEY` before opening the API listener. Report missing names together without logging values. Derive each optional provider's status from its configuration and known availability. A failed availability check yields unavailable even when credentials are present. Only the status reaches the client.

### Provision once and issue a separate bearer per exchange

Verify the Clerk token for the configured instance, including signature, issuer, expiry, and allowed audience or authorized party as applicable. In a read-committed transaction, take a transaction-scoped advisory lock on the verified subject before looking it up. On first sign-in, commit the user, workspace and pending machine reservation together. Concurrent callers reuse that reservation. Use the deployment and permanent workspace ID as the provider's idempotency key for creating or finding the VM and disk. If creation times out or the API stops after allocation, reconcile that identity with the provider before retrying. Cloud calls run outside the database transaction.

In sample mode, publish the fixture PDF once the private filesystem is reachable. In one transaction, lock the reservation, recheck its state, commit any required seed records once and mark the workspace ready. Live mode starts with an empty workspace filesystem and no sample records. Issue a distinct session to each exchange only after readiness commits. Returning sessions reuse the existing binding. Keep a failed or interrupted reservation retryable without exposing a partial workspace. On VM replacement, verify the deployment and owner binding before reattaching the original disk.

Enforce `users.clerk_user_id UNIQUE NOT NULL` and `users.workspace_id UNIQUE NOT NULL REFERENCES workspaces(id)`. Key sessions by a digest of a cryptographically random opaque bearer and store the user reference, expiry, and revocation state. Check this row on each private request. Deleting a session's authority then takes effect across API processes as soon as revocation commits. A self-contained Cueia token would need an additional revocation lookup to provide that behavior.

### Bootstrap database tables

Sample and live deployments use the same PostgreSQL schema. `PK` marks a primary key and `FK` a foreign key. Provider and Clerk identifiers are opaque text. All timestamps use `TIMESTAMPTZ` (timestamp with time zone). The planned migration types include `BIGINT` for sizes and versions, `BYTEA` for binary data, and `JSONB` for settings.

#### Users and sessions

```text
users:
  id                          UUID          PK
  clerk_user_id               TEXT          UNIQUE, NOT NULL
  workspace_id                UUID          UNIQUE, NOT NULL, FK -> workspaces.id
  created_at                  TIMESTAMPTZ

sessions:
  token_digest                BYTEA         PK
  user_id                     UUID          FK -> users.id
  created_at                  TIMESTAMPTZ
  expires_at                  TIMESTAMPTZ
  revoked_at                  TIMESTAMPTZ   nullable
```

#### Workspaces

```text
workspaces:
  id                          UUID          PK
  created_at                  TIMESTAMPTZ

workspace_machines:
  workspace_id                UUID          PK, FK -> workspaces.id
  provider                    TEXT
  region                      TEXT
  vm_resource_id              TEXT          nullable while provisioning
  workspace_disk_resource_id  TEXT          nullable while pending or if provider omits it
  provisioning_state          TEXT
  updated_at                  TIMESTAMPTZ   NOT NULL
  last_error                  JSONB         nullable
  power_state                 TEXT          nullable until observed
  last_heartbeat_at           TIMESTAMPTZ   nullable

  UNIQUE (provider, vm_resource_id)
  UNIQUE (provider, workspace_disk_resource_id)

workspace_preferences:
  workspace_id                UUID          PK, FK -> workspaces.id
  settings                    JSONB         validated JSON object
  updated_at                  TIMESTAMPTZ
```

The workspace ID survives VM replacement. Persist the filesystem identity needed to reopen the workspace even when the provider exposes no separate disk ID.

#### Mail and file storage

```text
mail_threads:
  id                          UUID          PK
  workspace_id                UUID          FK -> workspaces.id
  subject                     TEXT

  UNIQUE (workspace_id, id)

mail_messages:
  id                          UUID          PK
  workspace_id                UUID          FK -> workspaces.id
  thread_id                   UUID
  sender_address              TEXT
  body_text                   TEXT
  sent_at                     TIMESTAMPTZ

  UNIQUE (workspace_id, id)
  FK (workspace_id, thread_id) -> mail_threads(workspace_id, id)

files:
  id                          UUID          PK
  workspace_id                UUID          FK -> workspaces.id
  filename                    TEXT
  media_type                  TEXT
  byte_size                   BIGINT        >= 0
  sha256_digest               BYTEA         exactly 32 bytes
  storage_key                 TEXT          workspace-relative path

  UNIQUE (workspace_id, id)
  UNIQUE (workspace_id, storage_key)

mail_attachments:
  workspace_id                UUID          FK -> workspaces.id
  message_id                  UUID
  file_id                     UUID

  PK (workspace_id, message_id, file_id)
  FK (workspace_id, message_id) -> mail_messages(workspace_id, id)
  FK (workspace_id, file_id)    -> files(workspace_id, id)
```

The composite foreign keys keep a message, its thread, and its attached files in the same workspace. Register managed documents and files needing durable references. Ordinary directory entries require no row.

#### Calendar and worker claims

```text
calendar_events:
  id                          UUID          PK
  workspace_id                UUID          FK -> workspaces.id
  title                       TEXT
  starts_at                   TIMESTAMPTZ
  ends_at                     TIMESTAMPTZ
  time_zone                   TEXT

claim_records:
  id                          UUID          PK
  workspace_id                UUID          FK -> workspaces.id
  status                      TEXT          'eligible' or 'claimed'
  version                     BIGINT        >= 0
```

The seeded calendar event ends after it starts.

#### Deployment

```text
deployment_signing_keys:
  singleton                   BOOLEAN       PK, must be TRUE (proposed name and type)
  key_material                BYTEA
  created_at                  TIMESTAMPTZ

schema_migrations:
  version                     BIGINT        PK
  applied_at                  TIMESTAMPTZ
```

#### Candidate indexes

| Table | Candidate index | Purpose |
| --- | --- | --- |
| `sessions` | `(user_id)` | Find or revoke a user’s sessions |
| `sessions` | `(expires_at)` | Purge expired sessions when volume warrants it |
| `mail_messages` | `(workspace_id, thread_id, sent_at, id)` | Read and paginate messages in a thread |
| `mail_attachments` | `(workspace_id, file_id)` | Find attachment references to a file |
| `calendar_events` | `(workspace_id, starts_at)` | List workspace events by start time |

### Resolve workspace ownership at the storage boundary

Resolve the user and workspace from the active bearer. Pass the resolved workspace ID to every collection, detail, mutation, and link-issuance query. Qualify record reads and writes by both workspace and record ID. Treat a foreign ID like an absent record. Carry workspace IDs through references between threads and attachments so those relationships also stay within one owner. Assemble each snapshot from persisted records in a consistent read transaction. Updating a preference commits to the same rows the other session reads. For directory listings, searches, reads, and registration, the server resolves the machine from that workspace. A stopped or unreachable machine returns a recoverable unavailable response. Resolve requested paths beneath the workspace root on the VM at the time of access, including symlinks. Keep service-private staging and operating-system paths outside that root.

Check a supplied web origin against the allowlist before dispatch. Apply the unauthenticated limiter to failed or missing authentication, using the connection's peer address, bounded counters, and expiring windows. Ignore client-supplied forwarding headers when identifying the peer. An allowed-origin request within budget gets 401 for a missing, expired, or revoked credential. A disallowed origin gets 403 and a throttled caller gets 429. Map invalid inputs, missing or foreign records, and dependency failures to documented codes and safe messages at the HTTP boundary.

### Bind preview authority to a single operation

After an owner-scoped lookup, sign a payload containing the workspace, file ID, purpose, read operation, and short expiry. The preview handler checks the signature, route resource, operation, purpose, and time on every use, then resolves that file's metadata. Only the preview handler accepts this credential. Inject a clock into issuance and verification so expiry can be tested without waiting.

Initialize a random signing key in a deployment-private singleton PostgreSQL row using a unique key and insert-on-conflict. Read the committed winner before serving requests. A per-process key would invalidate links on restart and make concurrent starters disagree. Keep raw key material, session tokens, and signed URLs out of logs.

### Publish bytes before committing their database references

Publish the sample PDF on the owner's persistent filesystem and link its file ID to the seeded mail message. The mail attachment and its managed entry in Files both resolve to that committed `fileId`. Ordinary files appear directly from the filesystem without import.

#### Register a PDF from Files

1. Resolve the source beneath the owner's workspace root and normalize its workspace-relative path. Copy the PDF into private staging on the same disk. Compute its size and SHA-256 digest from the bytes in that copy.
2. Derive the workspace-relative `storage_key` under the service-managed document directory from the normalized source path and content digest. For example, use separate path components for the SHA-256 of the normalized path and the SHA-256 of the captured bytes. Flush the staged file and publish it atomically at that key without replacing an existing version. If another publisher already created the copy, verify its size and digest. Flush the containing directory before committing metadata.
3. In a PostgreSQL read-committed transaction, insert the file metadata using `INSERT ... ON CONFLICT (workspace_id, storage_key) DO NOTHING RETURNING id`. If the insert returns no row, issue a separate `SELECT id FROM files WHERE workspace_id = $1 AND storage_key = $2` in the same transaction. That statement can see the winning registration after it commits. Return the resulting `fileId` after commit.

Registering changed bytes at the same source path creates a distinct key and file ID. Retain each immutable copy. For direct document reads and signed previews, resolve the authorized workspace and `fileId` to the committed `files.storage_key` and serve those stored bytes. An unexpired preview of the first version continues to open that version after the edit. The file service lists only committed, verified entries in the managed directory. Ordinary user directories are listed directly.

Coordinate managed publication and recovery per workspace with PostgreSQL advisory locks. Publishers hold a shared lock from staging through commit. Recovery takes an exclusive lock, clears abandoned entries only in Cueia's private staging area, and withholds incomplete managed documents. A retry can adopt a complete managed object under the same publication key after verifying its bytes. Reconcile on startup or reconnection before exposing that workspace's managed documents. Never traverse the user's directories to delete files missing from PostgreSQL. Original sample bytes and independently authored expected results live in the repository with their provenance.

### Make the claim decision in PostgreSQL

Store an eligible record's ID, status, and version. Both Go processes call the same `pgx` update conditioned on ID, eligible status, and expected version. The update records the claim and increments the version. One returned row means success, zero rows means the caller lost the claim. A process-local mutex cannot coordinate separate workers. The acceptance harness holds two processes at a barrier after both have read the version, then checks their results and the final row.

## Risks / Trade-offs

- **OAuth can make an acceptance run environment-dependent** → Use a configured Clerk test instance for the Google and GitHub journeys. Exercise time, rejected tokens, and error handling with controlled fixtures in separate tests.
- **An interrupted allocation or write can outlive the caller.** Exercise a lost provider response and both sides of file rename. Verify one machine, one seed set, and preservation of a file created directly by the owner.
- **A previous successful migration container can mask a later failure** → The deployment start command creates a fresh one-shot migration container each time. Exercise a failure with the same service dependencies as a normal start.

## Migration Plan

1. Build the three commands from the `backend/` module in the root multi-stage `Dockerfile`. Package the SQL files from `backend/migrations/` with the migration target. Root `compose.yaml` provides PostgreSQL with a health check and persistent volumes per deployment.
2. For a fresh start or redeploy, keep both application services stopped until the current startup's migration job succeeds. Require database health for the migration service and successful migration completion for both application services. Keep the migration invocation outside Turbo caching. Record applied SQL versions in PostgreSQL and commit each version with its schema changes.
3. If a migration fails, keep the API and worker stopped, correct the failed migration or deployment configuration, then rerun against the retained database. For rollback across an incompatible schema change, restore a matching database backup and the corresponding workspace disk snapshots before starting the previous image.
