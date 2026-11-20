# Cueia workspace foundation

## Why

An owner can sign in, work with fictional sample data, and reopen the same workspace after the API restarts. A developer starting with the Cueia repository alone must be able to install, build, run, and verify that journey.

## What Changes

### Repository and startup

Build the TypeScript workspace at the repository root. Pin Bun in `package.json`, track one `bun.lock`, and install with `bun install --frozen-lockfile`. Coordinate builds, type checks, and tests through root `turbo.json`. Run the Hono API, task worker, browser worker, and migration command under Bun. Share domain types and Zod validation across the API, Expo client, and workers. Author the application code, artwork, test data, and expected results for Cueia. Declare third-party libraries and service protocols as dependencies.

Sample mode binds to loopback and seeds an empty data directory once. Include a fictional owner, mailbox thread with a fillable PDF, calendar event, and workspace preferences. Use stable IDs so restarting or signing in again opens the same records. Live mode requires an operator-supplied access secret and keeps live accounts separate from sample records. Both API modes require the server-side `CPK_INTELLIGENCE_API_KEY`. Validate required values before listening and identify missing settings by name. Report optional providers as disabled, configured, or unavailable according to their actual state.

### Access and storage

- Provide `POST /api/session` for sign-in, `GET /api/workspace` for the authenticated snapshot, and `DELETE /api/session` for sign-out. The loopback sample accepts an empty sign-in object for its fictional owner. Live sign-in verifies the configured access secret. Return an opaque bearer token on success and store its digest and expiry. Reject a disallowed web origin with 403 even when the caller has a valid session.
- Exchange the owner’s access credential for an expiring session. Authenticate private requests, derive the owner from that session, and scope collection queries, individual reads, and mutations to that owner. Sign-out revokes the presenting session. Expired and revoked sessions return 401. Allow only configured browser origins. Rate-limit unauthenticated attempts by connection within bounded windows and report throttling with 429.
- Issue short-lived signed links for a specific resource and permitted operation, such as viewing one PDF. Check purpose, resource ID, and expiry on every use. Persist signing material across restarts and initialize it atomically on first start. A preview link must confer no authority to list the workspace or modify a file. Return a structured error code and safe message for invalid inputs, missing resources, and unavailable dependencies.
- Persist records and owner-scoped snapshots with Drizzle ORM, one PostgreSQL schema, and versioned SQL migrations. Use embedded PGlite for a single-process sample installation. Use shared PostgreSQL for separate API and worker processes. Store document bytes in a private data directory accessible to the processes that consume them. Persist the stable file identity and metadata in the database. Publish complete files and reconcile interrupted writes before exposing a document.
- Build API, worker, and migration services from a root multi-stage `Dockerfile`. Root `compose.yaml` waits for PostgreSQL health, runs one uncached migration job, then starts the API and worker on migration success. The embedded API applies pending migrations before serving requests.

### Acceptance on a fresh installation

1. Install from the lockfile, build from the root, configure a sample API, and check its health. Sign in twice. Read the same seeded thread and attachment through both sessions, change a preference, and record the IDs and file bytes.
2. Restart with the same data directory. Verify the changed preference, record counts, IDs, and bytes. Sign out of one session and read with each token. Only the still-active session succeeds. Advance a fixture clock beyond link expiry and verify a fresh authorized link opens the same file.
3. Start two Bun processes on the same Docker PostgreSQL database. Release them together to claim one eligible record using its current version. Require one successful update. Start Compose against a deliberately failing fixture migration and verify neither application process accepts work.

## Capabilities

### New Capabilities

- `workspace-access` covers startup modes, owner sessions, API access, and signed resources.
- `workspace-storage` covers durable owner data and storage shared by API and worker.

### Modified Capabilities

## Impact

All feature proposals depend on this workspace, authentication, and storage contract. Keep production data and service credentials on the server. Record provenance for independently created fixtures and assets and for redistributed dependencies.

## Build order and ownership

| Stage | Proposals | Working milestone |
| --- | --- | --- |
| 1 | Foundation | Sign in, restart, recover sample data, and revoke a session. |
| 2 | [Tasks](../model-task-recovery/proposal.md), [runtime](../build-agent-runtime/proposal.md), [clients](../build-clients/proposal.md) | Submit a job, close the client, and return to its saved result. The client shell can start alongside the worker. |
| 3 | [Threads](../map-conversations/proposal.md), [connectors](../build-connectors/proposal.md), [browser](../build-browser/proposal.md), [computer](../build-computer/proposal.md) | Replay a conversation, inspect its evidence, and review an action. |
| 4 | [Personal workspace](../build-personal-workspace/proposal.md), [documents](../trace-document-review/proposal.md), [watches](../characterize-watches/proposal.md) | Reopen a saved result, send a reviewed form, and follow a watch alert. |
| 5 | [Release](../ship/proposal.md) | Build from a clean Cueia checkout and record results by platform and provider. |

Bank feeds, device push, and OCR are future extensions. Show the actual availability of each configured integration throughout the five stages.
