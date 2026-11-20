# Cueia release acceptance

## Why

A developer starting with only a fresh Cueia checkout needs to build the product, exercise its workflows, and identify precisely which platforms and live services were verified.

## What Changes

### Distribution and deployment

Exercise bootstrap's root install, build, migration, start, and test commands on a clean checkout. Resolve the pinned Go module from `apps/server/go.mod` and `apps/server/go.sum`. Install the Bun workspaces with `bun install --frozen-lockfile` from the root. Run lint and root Turbo tasks for Go checks, tests, and builds of the API, task worker, and migration command, and for TypeScript checks, tests, and builds of the dashboard and Bun services. Include operator, contributor, security, and license documentation, configuration examples with placeholders, and provenance for distributed assets, dependencies, and fixtures. Build tests and expected results from Cueia's own requirements. Use only declared build inputs. [Bootstrap repository layout](../bootstrap/design.md#repository-layout)

Ship the Vite production build, Render blueprint, and root multi-stage Dockerfile and Compose configuration. Configure web hosting to serve the app entry point on direct visits to client routes so React Router v7 can open a saved thread. Preserve API responses and static asset handling. Defer iOS and Android packaging to a later release.

A configured `docker compose up` starts healthy PostgreSQL, runs an uncached migration job on each startup, applies pending versions, and gates API and task-worker startup on success. Package the private Bun service for CopilotKit, the optional Chromium service and Linux computer image, and keep worker controls private. Persist the database and browser profiles. Configure the workspace VM provider and document how split workers resolve the same owner's disk through the [bootstrap machine binding](../bootstrap/design.md#bootstrap-database-tables). Cover backup and recovery of the workspace disks together with their database bindings. Report health and capability status when optional providers are absent. Include clean shutdown, restart, and recovery instructions.

Provide a local demo launcher with isolated fictional data, a scripted model, and fixture transports for network services. Keep telemetry disabled and production accounts inaccessible in the automated fixtures. Exercise a fictional mail request that leads to two sourced browser choices, and the seeded workshop form that leads to one reviewed reply. Document Clerk settings, the server-side `CPK_INTELLIGENCE_API_KEY`, and workspace VM provider setup for sample and live launches. Verify sample binds to loopback and the same Clerk subject opens a separate live workspace with no sample records. Identify runs that use a service double.

### Telemetry and operational evidence

Use the CopilotKit SDK transport with Cueia app attribution and server-side project binding through `CPK_TELEMETRY_ID`. Keep the Intelligence API key separate and private. Default the sampling rate to 1 and honor an explicit `COPILOTKIT_TELEMETRY_SAMPLE_RATE`. Honor `COPILOTKIT_TELEMETRY_DISABLED` and `DO_NOT_TRACK` when set to 1 or true. Exercise runtime initialization and request events against an intercepted transport and verify the attribution, project binding, sampling, both opt-outs, and absence of the API key. Suppress transport in ordinary fixtures. Log operation, status, time, and correlation ID with credentials and private workspace content redacted.

### Acceptance inventory

For each public route, client screen, task type, adapter/mode, and build entry point, name its owning Cueia proposal, setup, authored fixture, steps, and observable expectation. Take HTTP methods, paths, and error responses from the release's `apps/server/api/openapi.yaml`. Bootstrap lists candidate routes alongside the three specified session and snapshot routes. Record the actual result, environment, and evidence from the run. Include failure cases. Compare generated IDs by their relationships and timestamps by ordering or fixture clock. Keep byte equality, event counts, and permissions exact. [Bootstrap API routes](../bootstrap/proposal.md#bootstrap-api-routes)

| Run | Required evidence |
| --- | --- |
| Fresh root install and build | Go dependencies verify and the frozen Bun install succeeds. Regenerate `packages/api-types` from `apps/server/api/openapi.yaml` and check for drift. Type-check the dashboard, run lint and tests, build the Go binaries and Bun services, and verify Vite produces a deployable web bundle. |
| Sample and split-worker deployments | Concurrent first sign-ins create one workspace VM and disk per person. Refresh Files to find a nested file created directly on Alice's VM. Put different bytes at the same path on Bob's disk and verify isolation. Restart Compose and replace Alice's VM with her disk reattached. Verify file bytes, stable document IDs, preferences, and independent session revocation. Two Go workers contend for one due PostgreSQL task and only one invokes its handler. A failed migration prevents both services from starting. |
| Browser and local computer | Real Chromium and the browser container retain the fixture profile after restart and import one PDF. A disposable Docker computer retains file bytes and records an interrupted command. |
| Document and personal workflows | Start with zero outgoing messages. Approve a filled PDF once, restart, and observe one reply and one receipt. Reopen a finance summary, linked milestone, and watch alert using the expected records from their proposals. |
| Client sessions | Serve the production web build and run the client journey in desktop and mobile browsers. With a valid session, open a saved thread URL directly and reload it. Verify the same thread opens both times. Record the browser, device or viewport, and observed results. |
| Configured external services | Record private cloud VM allocation, owner disk isolation and reattachment, Google authorization and reviewed delivery, Intelligence replay, each selected model/gateway, live Jev, Parallel search, and hosted desktop with the environment, observed result, or specific pending prerequisite. Identify adapter-only coverage explicitly. |

Run these checks on the sample deployment used for bootstrap's fresh-install acceptance. [Bootstrap acceptance](../bootstrap/proposal.md#acceptance-on-a-fresh-installation)

1. From two of Alice's sessions, register the same ordinary PDF concurrently. Require the same `fileId` and one metadata row for that workspace and storage key. Obtain a preview link, edit the source, and register it again. Record the second ID and each version's expected bytes.
2. Restart PostgreSQL and the API, replace Alice's VM, and reattach her disk. Open the still-valid first link and preview the second ID. Compare each response with its version's exact bytes. Verify Bob's requests for Alice's file and for a preview link are both denied. Revoke one of Alice's sessions. After the first link expires, use her other session to obtain a fresh link for the original ID and compare the bytes again.
3. In the working client, ask Main chat for a reminder 30 minutes before the sample workshop. Check that the confirmation and Scheduled tasks give the same task ID and due time. With Tasks open, type a new draft, open the form from Files, and close its preview. Verify the file regains focus and the draft and task entry survive. Close the client, let the worker finish, then reopen the saved outcome by task ID.

Run the local acceptance suite in a fresh Cueia-only build context. Record missing dependencies, failed assertions, unexercised modes, and the fixture needed to reproduce each discrepancy. A scripted response supplies evidence for the scripted mode. Mark live-provider or device checks pending until exercised in their stated environment.

## Capabilities

### New Capabilities

- `deployment-modes` covers reproducible local, split-worker, and hosted operation.
- `runtime-telemetry` covers attribution, sampling, privacy settings, and opt-out behavior.

### Modified Capabilities

## Impact

Release acceptance consumes all eleven feature proposals. The release report links each measurement to its Cueia contract and fixture, so a later developer can repeat the run from this repository.
