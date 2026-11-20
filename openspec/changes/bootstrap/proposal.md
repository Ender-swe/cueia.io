# Cueia workspace foundation

## Why

Each person can sign in to a private workspace, work with fictional data in sample mode, and reopen the same records after the API restarts. A developer starting with the Cueia repository alone must be able to install, build, run, and verify that journey.

## What Changes

### Repository and startup

Build the API, task worker, and migration command in a Go module. Pin the Go toolchain and track `go.mod` and `go.sum`. Serve the workspace API with Go’s `net/http`. Describe its HTTP routes, JSON requests and responses, and error formats in an OpenAPI document. Generate the React client’s TypeScript API types from that contract and validate requests in the Go handlers.

Keep a root Bun workspace for TypeScript packages. Pin Bun in `package.json`, track one `bun.lock`, and install with `bun install --frozen-lockfile`. Coordinate Go and TypeScript builds, checks, and tests through root Turbo tasks. Build the React and TypeScript web client with Vite and use React Router v7 for browser navigation. Run the Playwright browser worker under Bun. Author the application code, artwork, test data, and expected results for Cueia. Declare third-party libraries and service protocols as dependencies.

Both modes use Clerk with open signup through Google or GitHub only. Sample mode binds to loopback. A single “Open sample workspace” action uses the current Clerk session or starts Clerk sign-in. On first sign-in, each person receives a private copy of the fictional mailbox thread, fillable PDF, calendar event, and default preferences. Later sign-ins and restarts reopen those records with the same IDs. Keep live workspaces separate from sample data.

Both API modes require Clerk configuration and the server-side `CPK_INTELLIGENCE_API_KEY`. Validate required values before listening and identify missing settings by name. Report optional providers as disabled, configured, or unavailable according to their actual state.

### Access and storage

- At `POST /api/session`, verify the Clerk session token and identify the user by its verified subject (`sub`). On success, return an expiring opaque Cueia bearer token and store its digest and expiry. A missing or invalid Clerk token returns 401 in both modes.
- Store the Clerk user ID in a unique, non-null `users.clerk_user_id` column. Give each user a unique, non-null `users.workspace_id` foreign key to `workspaces.id`. Create the user and workspace in one transaction. Concurrent first sign-ins for the same Clerk user must resolve to one workspace and, in sample mode, one complete set of sample records.
- Authenticate private requests and derive the workspace from the session’s user. Scope collection queries, individual reads, mutations, and signed-link issuance to that workspace. `GET /api/workspace` returns its snapshot. `DELETE /api/session` revokes the presenting Cueia session. Expired and revoked sessions return 401. Reject a disallowed web origin with 403 even with a valid session. Rate-limit unauthenticated attempts by connection within bounded windows and report throttling with 429.
- Issue short-lived signed links for a specific resource and permitted operation, such as viewing one PDF. Check purpose, resource ID, and expiry on every use. Persist signing material across restarts and initialize it atomically on first start. A preview link must confer no authority to list the workspace or modify a file. Return a structured error code and safe message for invalid inputs, missing resources, and unavailable dependencies.
- Persist records and owner-scoped snapshots in PostgreSQL through `pgx`. Use one schema and the same versioned SQL migrations in sample and live deployments. Point the API and task worker at the same database within each deployment. Store document bytes in a private data directory accessible to the processes that consume them. Persist the stable file identity and metadata in the database. Publish complete files and reconcile interrupted writes before exposing a document.
- Build API, worker, and migration services from a root multi-stage `Dockerfile`. In both sample and live modes, root `compose.yaml` provides PostgreSQL, waits for database health, runs one uncached migration job, then starts the API and worker on migration success.

### Acceptance on a fresh installation

1. Install from the lockfile, build from the root, configure the sample API and Clerk, start the sample deployment with Compose, and check API health. Verify an empty sign-in request returns 401. Sign Alice in through Google and open two Cueia sessions concurrently on her first visit. Verify both sessions see one workspace with the same seeded thread and attachment IDs. Change a preference through one session, read it through the other, and record the IDs and file bytes.
2. Sign Bob in through GitHub. Verify he receives his own workspace with its own thread and attachment IDs and the default preference. Using Bob’s session and Alice’s attachment ID, attempt a direct read and request a signed preview link. Both requests must fail. Bob can open his own attachment.
3. Restart PostgreSQL and the API with the same database and document volumes. Verify each workspace retains its record counts, IDs, preferences, and attachment bytes. Sign out of one of Alice’s sessions and verify its token returns 401. Alice’s other session and Bob’s session still open their respective workspaces. Advance a fixture clock beyond link expiry and verify a fresh authorized link opens the same file.
4. Start two Go processes on the same Docker PostgreSQL database. Release them together to claim one eligible record using its current version. Require one successful update. Start Compose against a deliberately failing fixture migration and verify neither application process accepts work.

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
