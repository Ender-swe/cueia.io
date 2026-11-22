# Cueia workspace foundation

## Why

Cueia is a private workspace where people chat with AI agents and decide what work to automate. An owner can ask for a workshop reminder in chat, see when the agent scheduled it in Tasks, and open the registration form from Files while reading the conversation.

Bootstrap establishes the workspace for that journey. Each person gets one private workspace per deployment, with its own machine identity and persistent filesystem. The bootstrap sample deployment and acceptance suite use the internal provider, which represents each VM and retained disk locally. Azure provisioning is outside this milestone. They can sign in, inspect their own sample mail and calendar event, open the attached PDF, change a preference, and recover those records after a restart. A developer starting with the Cueia repository alone must be able to install, build, run, and verify that milestone.

## What Changes

### Repository and startup

Build the Go backend as one module in `apps/server/`. Pin the Go toolchain and track `apps/server/go.mod` and `apps/server/go.sum`. Under `apps/server/cmd/`, `api` registers the feature handlers with `net/http` and manages the HTTP listener, `worker` claims eligible work, and `migrate` applies pending SQL files from `apps/server/migrations/` and exits. Group private Go code by feature under `apps/server/internal/`, with shared packages for configuration, database access, and HTTP middleware.

Register `apps/*` and `packages/*` in the root Bun workspace. Pin Bun in `package.json`, track one `bun.lock`, and install with `bun install --frozen-lockfile`. Coordinate Go and TypeScript builds, checks, and tests through root Turbo tasks.

Build the React dashboard with Vite and React Router v7. Run the Playwright browser worker under Bun. See the [repository layout](design.md#repository-layout) for the planned directories and their responsibilities.

Mount the Adobe React Spectrum 2 `Provider` from `@react-spectrum/s2` at the dashboard root so pages and shared components use the same theme and browser navigation. Describe the HTTP routes, JSON requests and responses, and error formats in `apps/server/api/openapi.yaml`. The dashboard imports the generated types from `packages/api-types`, and the Go handlers validate requests. Author the application code, artwork, test data, and expected results for Cueia. Declare third-party libraries and service protocols as dependencies.

Both modes use Clerk with open signup using an email address and password, Google, or GitHub. Sample mode binds to loopback. A single “Open sample workspace” action uses the current Clerk session or starts Clerk sign-in. On first sign-in, each person receives a private copy of the fictional mailbox thread, fillable PDF, calendar event, and default preferences. Later sign-ins and restarts reopen those records with the same IDs. Keep live workspaces separate from sample data.

Both API modes require Clerk configuration and the server-side `CPK_INTELLIGENCE_API_KEY`. Validate required values before listening and identify missing settings by name. Report optional providers as disabled, configured, or unavailable according to their actual state.

### Chat workspace

Open **Main chat** after sign-in. Use the [dark desktop layout](design.md#chat-files-and-tasks) with a narrow navigation rail, a left sidebar for Main chat and Side chats, the conversation and composer in the center, and a closable Tasks panel on the right. **Files** switches the left sidebar to a directory browser for the owner's workspace filesystem. Opening or refreshing Files lists user- and agent-created files, including files written directly on the VM without a Cueia database record. **Tasks** shows the agent's connection state, scheduled tasks, and today's activity. The two panels open independently so the owner can find a file and check on work from the same conversation. Omit Call and Invite controls.

In bootstrap, Main chat introduces the sample workshop using the owner's mailbox thread and calendar event. Files opens the attached form. Tasks shows the empty schedule and activity sections of a new workspace. Live replies and task delegation arrive with the runtime and client in stage 2. The task scheduler owns reminder times and the client shows a confirmation only after the task is saved.

### Access and storage

- At `POST /api/session`, verify the Clerk session token and identify the user by its verified subject (`sub`). On success, return an expiring opaque Cueia bearer token and store its digest and expiry. A missing or invalid Clerk token returns 401 in both modes.
- Store the Clerk user ID in a unique, non-null `users.clerk_user_id` column. Give each user a unique, non-null `users.workspace_id` foreign key to `workspaces.id`. Create the user and workspace in one transaction. Concurrent first sign-ins for the same Clerk user must resolve to one workspace, one VM with its own persistent disk, and, in sample mode, one complete set of sample records. A retry after an interrupted provider request must recover the original allocation.
- Authenticate private requests and derive the workspace from the session’s user. Scope collection queries, individual reads, mutations, and signed-link issuance to that workspace. `GET /api/workspace` returns its snapshot. `DELETE /api/session` revokes the presenting Cueia session. Expired and revoked sessions return 401. Reject a disallowed web origin with 403 even with a valid session. Rate-limit unauthenticated attempts by connection within bounded windows and report throttling with 429.
- Issue short-lived signed links for a specific resource and permitted operation, such as viewing one PDF. Check purpose, resource ID, and expiry on every use. Persist signing material across restarts and initialize it atomically on first start. A preview link must confer no authority to list the workspace or modify a file. Return a structured error code and safe message for invalid inputs, missing resources, and unavailable dependencies.
- Store workspace records and stable document references in PostgreSQL through `pgx`. Sample and live deployments use the same migrations and separate databases and VM allocations. The API and worker resolve each owner's machine from the workspace ID. Read file bytes and list directories on that machine's workspace filesystem. Register each PDF version under a key derived from its normalized workspace-relative path and the SHA-256 digest of the captured bytes. The existing `UNIQUE (workspace_id, storage_key)` constraint gives concurrent registrations of that version one `fileId`. A change in content creates a new version. Each ID continues to open its own immutable copy. Restrict cleanup of abandoned writes to Cueia's private staging area. [Document registration design](design.md#publish-bytes-before-committing-their-database-references)
- Build API, worker, and migration services from a root multi-stage `Dockerfile`. In both sample and live modes, root `compose.yaml` provides PostgreSQL, waits for database health, runs one uncached migration job, then starts the API and worker on migration success.

### Bootstrap API routes

**Specified** means the workspace access spec names the method and path. **Proposed** means the operation is planned and its method and path are candidates for the OpenAPI contract. [Workspace access spec](specs/workspace-access/spec.md) · [OpenAPI contract task](tasks.md#1-build-and-contract)

| Route                                             | Status    | What it does                                                                                                                                                        |
| ------------------------------------------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `POST /api/session`                               | Specified | Exchanges a verified Clerk token for an expiring, opaque Cueia bearer. Two sign-ins for the same person get distinct bearers for the same workspace.                |
| `GET /api/workspace`                              | Specified | Returns the owner's snapshot. In sample mode, Main chat can show the workshop mail, attachment reference, calendar event, and saved preferences from this response. |
| `DELETE /api/session`                             | Specified | Signs out the presenting Cueia session. Its next private request gets 401. The owner's other sessions keep working.                                                 |
| `GET /api/health`                                 | Proposed  | Reports API health for the deployment check.                                                                                                                        |
| `GET /api/providers`                              | Proposed  | Reports each optional provider as disabled, configured, or unavailable.                                                                                             |
| `GET /api/mail/threads`                           | Proposed  | Lists the owner's mail threads independently of the initial snapshot.                                                                                               |
| `GET /api/calendar/events`                        | Proposed  | Lists the owner's calendar events independently of the initial snapshot.                                                                                            |
| `PATCH /api/workspace/preferences`                | Proposed  | Saves changed preferences. The owner's other session sees them on its next snapshot.                                                                                |
| `GET /api/files?directory=workshop`               | Proposed  | Lists the current contents of a directory on the owner's VM, including files created directly on disk.                                                              |
| `GET /api/files/search?name=registration`         | Proposed  | Finds matching filenames across nested directories.                                                                                                                 |
| `GET /api/files/content?path=reports/summary.csv` | Proposed  | Reads the current bytes of an ordinary file by workspace-relative path.                                                                                             |
| `POST /api/documents`                             | Proposed  | Registers a PDF by workspace-relative path and returns the `fileId` for the captured version. Registering the same path and bytes again returns the same ID.        |
| `GET /api/documents/{fileId}/content`             | Proposed  | Reads the registered version's bytes using the owner's Cueia bearer. A mail attachment uses this same file ID.                                                      |
| `POST /api/documents/{fileId}/preview-links`      | Proposed  | Checks ownership and returns a short-lived signed URL for viewing that PDF version.                                                                                 |
| `GET /api/previews/{fileId}?token=…`              | Proposed  | Serves that version's stored bytes when the preview dialog opens a valid signed URL. The link expires at its signed expiry time.                                    |

### Acceptance on a fresh installation

1. Install from the lockfile, build from the root, configure the sample API, Clerk, and the internal workspace provider, start the sample deployment with Compose, and check API health. Verify an empty sign-in request returns 401. Sign Alice in through Google and open two Cueia sessions concurrently on her first visit. Verify both sessions see one workspace and one VM with the same seeded thread and attachment IDs. From Main chat, open Tasks, find the form in Files, open its preview, and close it. Verify Alice returns to the same conversation with both panels open. Create a nested file directly on the VM, refresh Files in the other session, and find it without an import. Change a preference through one session, read it through the other, and record the IDs and file bytes.
   1. From both of Alice's sessions, concurrently register an ordinary PDF at the same workspace-relative path. Verify both receive the same `fileId` and PostgreSQL contains one `files` row for that workspace and storage key. Obtain a preview link for this version.
   2. Edit the PDF at that path and register it again. Verify it gets a new `fileId`. Open the still-valid first link and a link for the new ID. Compare each response with the exact bytes captured before and after the edit, respectively.
2. Sign Bob in through GitHub. Verify Bob gets a separate workspace, VM, and disk, his own thread and attachment IDs, and the default preference. Using Bob’s session and Alice’s attachment ID, attempt a direct read and request a signed preview link. Both requests must fail. Create the same relative filename on Bob's disk with different contents. Each owner sees only their own bytes. Bob can open his own attachment.
3. Restart PostgreSQL, the API, and the workspace VMs with their disks retained. Replace Alice's VM and reattach her disk. Verify each workspace retains its record counts, IDs, preferences, and attachment bytes, along with the directly created file. Sign out of one of Alice’s sessions and verify its token returns 401. Alice’s other session and Bob’s session still open their respective workspaces. Advance a fixture clock beyond link expiry and verify a fresh authorized link opens the same file.
4. Start two Go processes on the same Docker PostgreSQL database. Release them together to claim one eligible record using its current version. Require one successful update. Start Compose against a deliberately failing fixture migration and verify neither application process accepts work.

## Capabilities

### New Capabilities

- `workspace-access` covers startup modes, owner sessions, API access, and signed resources.
- `workspace-storage` covers provisioning one private VM filesystem per workspace, directory browsing, durable document identity, and storage shared by the API and worker.

### Modified Capabilities

## Impact

All feature proposals depend on this workspace, authentication, and storage contract. Keep production data and service credentials on the server. Record provenance for independently created fixtures and assets and for redistributed dependencies.

## Build order and ownership

| Stage | Proposals                                                                                                                                                                      | Working milestone                                                                                                                                                                   |
| ----- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1     | Foundation                                                                                                                                                                     | Sign in to one private VM-backed workspace, browse its files, preview the sample PDF, recover files after restart, and revoke a session.                                            |
| 2     | [Tasks](../model-task-recovery/proposal.md), [runtime](../build-agent-runtime/proposal.md), [clients](../build-clients/proposal.md)                                            | Ask an agent to do work in chat, track the delegated task, close the client, and return to its saved result. The client can connect the shell to these APIs as the worker is built. |
| 3     | [Threads](../map-conversations/proposal.md), [connectors](../build-connectors/proposal.md), [browser](../build-browser/proposal.md), [computer](../build-computer/proposal.md) | Replay a conversation, inspect its evidence, and review an action.                                                                                                                  |
| 4     | [Personal workspace](../build-personal-workspace/proposal.md), [documents](../trace-document-review/proposal.md), [watches](../characterize-watches/proposal.md)               | Reopen a saved result, send a reviewed form, and follow a watch alert.                                                                                                              |
| 5     | [Release](../ship/proposal.md)                                                                                                                                                 | Build from a clean Cueia checkout and record results by platform and provider.                                                                                                      |

Bank feeds, device push, and OCR are future extensions. Show the actual availability of each configured integration throughout the five stages.
