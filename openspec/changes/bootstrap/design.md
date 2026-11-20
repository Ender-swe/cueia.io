# Bootstrap design

## Context

Cueia has planning documents and no application or database to migrate. The first deployment must work from a clean Cueia checkout. Two processes can race on first sign-in, and two different stores must agree before the owner can read a PDF. The foundation milestone and its four acceptance cases are in the [bootstrap proposal](proposal.md#acceptance-on-a-fresh-installation).

## Goals / Non-Goals

One verified Clerk subject maps to a durable workspace. Separate Cueia sessions can share that workspace and expire independently. The same Go storage code and SQL migrations serve the API, worker, and integration tests. The worker boundary in this change ends at a versioned claim. Scheduling, leases, retries, and task execution belong to [task recovery](../model-task-recovery/proposal.md).

## Decisions

### Build from the root and share the HTTP contract

Use one pinned Go module with `cmd/api`, `cmd/worker`, and `cmd/migrate`, and shared internal packages for configuration and PostgreSQL access through `pgx`. Track `go.mod` and `go.sum`. Use a root Bun workspace for the Vite React client, the generated TypeScript contract, and the Bun-run Playwright browser worker. Pin Bun in `package.json` and commit one `bun.lock`. Expose build, type check, and test through root Turbo tasks, declaring Go sources, modules, migrations, and the API contract as inputs where applicable.

Version OpenAPI alongside the Go `net/http` handlers. Document health and provider status, session exchange and revocation, workspace snapshot, preference updates, attachment reads, and preview issuance and use. Generate client types from it and check for drift in the root checks. Handlers validate requests at runtime. Separate handwritten client types would allow a response change to compile on just one side of the contract.

The React Router v7 sample entry route starts Clerk sign-in when necessary and returns to the same route to finish the exchange and fetch the snapshot. Configure open signup with only Google and GitHub in Clerk for each mode. Only the publishable Clerk configuration enters the web build. Keep the Cueia bearer in client memory and exchange again using the existing Clerk session on a return visit. Explicit sign-out revokes the presenting Cueia bearer before clearing client state.

### Keep deployment state and secrets separate by mode

Use distinct Compose projects with distinct database and private document volumes for sample and live. Both projects run the same images and schema. Separate volumes allow the same Clerk subject to sign in to both without sharing records, signing keys, or fixture bytes. A mode flag on individual queries would leave isolation dependent on every caller. Publish the sample API only on host loopback under Compose and bind a directly run sample API to loopback.

Validate the Clerk verification configuration and `CPK_INTELLIGENCE_API_KEY` before opening the API listener. Report missing names together without logging values. Derive each optional provider's status from its configuration and known availability. A failed availability check yields unavailable even when credentials are present. Only the status reaches the client.

### Provision once and issue a separate bearer per exchange

Verify the Clerk token for the configured instance, including signature, issuer, expiry, and allowed audience or authorized party as applicable. Use the verified `sub` as the identity key. In a PostgreSQL transaction, take a transaction-scoped advisory lock keyed by that subject **before** looking up the user. Use read-committed isolation so a waiting exchange sees the winner's commit on its subsequent lookup. For a new subject, insert the workspace and user, publish sample files and insert all fixture rows in sample mode, then insert the session. For a returning subject, reuse the workspace and insert only a new session. Return a bearer only after commit. A check followed by an unlocked insert could give concurrent callers different workspaces or cause one exchange to fail.

Enforce `users.clerk_user_id UNIQUE NOT NULL` and `users.workspace_id UNIQUE NOT NULL REFERENCES workspaces(id)`. Key sessions by a digest of a cryptographically random opaque bearer and store the user reference, expiry, and revocation state. Check this row on each private request. Deleting a session's authority then takes effect across API processes as soon as revocation commits. A self-contained Cueia token would need an additional revocation lookup to provide that behavior.

### Resolve workspace ownership at the storage boundary

Resolve the user and workspace from the active bearer. Pass the resolved workspace ID to every collection, detail, mutation, and link-issuance query. Qualify record reads and writes by both workspace and record ID. Treat a foreign ID like an absent record. Carry workspace IDs through references between threads and attachments so those relationships also stay within one owner. Assemble each snapshot from persisted records in a consistent read transaction. Updating a preference commits to the same rows the other session reads.

Check a supplied web origin against the allowlist before dispatch. Apply the unauthenticated limiter to failed or missing authentication, using the connection's peer address, bounded counters, and expiring windows. Ignore client-supplied forwarding headers when identifying the peer. An allowed-origin request within budget gets 401 for a missing, expired, or revoked credential. A disallowed origin gets 403 and a throttled caller gets 429. Map invalid inputs, missing or foreign records, and dependency failures to documented codes and safe messages at the HTTP boundary.

### Bind preview authority to a single operation

After an owner-scoped lookup, sign a payload containing the workspace, file ID, purpose, read operation, and short expiry. The preview handler checks the signature, route resource, operation, purpose, and time on every use, then resolves that file's metadata. Only the preview handler accepts this credential. Inject a clock into issuance and verification so expiry can be tested without waiting.

Initialize a random signing key in a deployment-private singleton PostgreSQL row using a unique key and insert-on-conflict. Read the committed winner before serving requests. A per-process key would invalidate links on restart and make concurrent starters disagree. Keep raw key material, session tokens, and signed URLs out of logs.

### Publish bytes before committing their database references

Give each workspace its own thread, attachment, event, and preference rows. Allocate fresh IDs during provisioning and retain the association between the thread and the fillable PDF. Store file ID, workspace, private storage key, size, media type, and content digest in PostgreSQL. Give API and worker access to the same private document volume. Neither the host web server nor the client serves paths from that volume directly.

Write new bytes to a temporary file on the destination filesystem, verify size and digest, flush the file, rename it atomically under an immutable key, and flush the directory before committing the referencing rows. A seed transaction publishes all its rows together. A crash before commit can leave an unreferenced file, while a metadata-first commit could expose a missing PDF. Readers select only committed metadata for complete files.

Coordinate publication and reconciliation with a deployment-wide PostgreSQL advisory lock. Writers hold a shared lock from staging through commit. The recovery pass holds the exclusive lock while inspecting the private directory against committed metadata, discards abandoned temporary or unreferenced files, and withholds entries with missing or mismatched bytes. Run recovery before serving documents on startup. A retry of rolled-back provisioning can then publish one complete set. Original sample bytes and independently authored expected results live in the repository with their provenance.

### Make the claim decision in PostgreSQL

Store an eligible record's ID, status, and version. Both Go processes call the same `pgx` update conditioned on ID, eligible status, and expected version. The update records the claim and increments the version. One returned row means success, zero rows means the caller lost the claim. A process-local mutex cannot coordinate separate workers. The acceptance harness holds two processes at a barrier after both have read the version, then checks their results and the final row.

## Risks / Trade-offs

- **OAuth can make an acceptance run environment-dependent** → Use a configured Clerk test instance for the Google and GitHub journeys. Exercise time, rejected tokens, and error handling with controlled fixtures in separate tests.
- **A filesystem failure can occur on either side of rename** → Interrupt publication at both boundaries against real volumes and compare recovered bytes to the original fixture.
- **A previous successful migration container can mask a later failure** → The deployment start command creates a fresh one-shot migration container each time. Exercise a failure with the same service dependencies as a normal start.

## Migration Plan

1. Build API, worker, and migration service targets from the root multi-stage `Dockerfile`. Root `compose.yaml` provides PostgreSQL with a health check and persistent volumes per deployment.
2. For a fresh start or redeploy, keep both application services stopped until the current startup's migration job succeeds. Require database health for the migration service and successful migration completion for both application services. Keep the migration invocation outside Turbo caching. Record applied SQL versions in PostgreSQL and commit each version with its schema changes.
3. If a migration fails, keep the API and worker stopped, correct the failed migration or deployment configuration, then rerun against the retained volumes. For rollback across an incompatible schema change, restore a matching database and document-volume backup before starting the previous image.
