# Cueia release acceptance

## Why

A developer starting with only a fresh Cueia checkout needs to build the product, exercise its workflows, and identify precisely which platforms and live services were verified.

## What Changes

### Distribution and deployment

Provide root install, build, migration, start, and test commands. Pin Go and Bun. Resolve Go dependencies from `go.mod` and `go.sum` and install TypeScript dependencies with `bun install --frozen-lockfile`. Run lint and root Turbo tasks for Go checks, tests, and builds of the API, task worker, and migration command, and for TypeScript checks, tests, and builds of the web client and Bun services. Include operator, contributor, security, and license documentation, configuration examples with placeholders, and provenance for distributed assets, dependencies, and fixtures. Build tests and expected results from Cueia’s own requirements. Use only declared build inputs.

Ship the Vite production build, Render blueprint, and root multi-stage Dockerfile and Compose configuration. Configure web hosting to serve the app entry point on direct visits to client routes so React Router v7 can open a saved thread. Preserve API responses and static asset handling. Defer iOS and Android packaging to a later release.

A configured `docker compose up` starts healthy PostgreSQL, applies migrations once, and gates API and task-worker startup on success. Package the private Bun service for CopilotKit, the optional Chromium service and Linux computer image, keep worker controls private, persist database and file/profile volumes across restarts, and document shared storage for split workers. Report health and capability status when optional providers are absent. Include clean shutdown, restart, and recovery instructions.

Provide a local demo launcher with isolated fictional data, a scripted model, and fixture transports for network services. Keep telemetry disabled and production accounts inaccessible in the automated fixtures. Exercise a fictional mail request that leads to two sourced browser choices, and a separate fillable workshop form that leads to one reviewed reply. Document the required server-side Intelligence project configuration for a normal app launch and identify when tests use a service double.

### Telemetry and operational evidence

Use the CopilotKit SDK transport with Cueia app attribution and server-side project binding through `CPK_TELEMETRY_ID`. Keep the Intelligence API key separate and private. Default the sampling rate to 1 and honor an explicit `COPILOTKIT_TELEMETRY_SAMPLE_RATE`. Honor `COPILOTKIT_TELEMETRY_DISABLED` and `DO_NOT_TRACK` when set to 1 or true. Exercise runtime initialization and request events against an intercepted transport and verify the attribution, project binding, sampling, both opt-outs, and absence of the API key. Suppress transport in ordinary fixtures. Log operation, status, time, and correlation ID with credentials and private workspace content redacted.

### Acceptance inventory

For each public route, client screen, task type, adapter/mode, and build entry point, name its owning Cueia proposal, setup, authored fixture, steps, and observable expectation. Record the actual result, environment, and evidence from the run. Include failure cases. Compare generated IDs by their relationships and timestamps by ordering or fixture clock. Keep byte equality, event counts, and permissions exact.

| Run | Required evidence |
| --- | --- |
| Fresh root install and build | Go dependencies verify and the frozen Bun install succeeds. Generate client types from OpenAPI, type-check the client against them, and run lint and tests. Build the Go binaries and Bun services, and verify Vite produces a deployable web bundle. |
| Sample and split-worker deployments | Sample records and attachment bytes survive a Compose restart with persistent PostgreSQL and document volumes. Two Go workers contend for one due PostgreSQL task and only one invokes its handler. A failed migration prevents both services from starting. |
| Browser and local computer | Real Chromium and the browser container retain the fixture profile after restart and import one PDF. A disposable Docker computer retains file bytes and records an interrupted command. |
| Document and personal workflows | Start with zero outgoing messages. Approve a filled PDF once, restart, and observe one reply and one receipt. Reopen a finance summary, linked milestone, and watch alert using the expected records from their proposals. |
| Client sessions | Serve the production web build and run the client journey in desktop and mobile browsers. With a valid session, open a saved thread URL directly and reload it. Verify the same thread opens both times. Record the browser, device or viewport, and observed results. |
| Configured external services | Record Google authorization and reviewed delivery, Intelligence replay, each selected model/gateway, live Jev, Parallel search, and hosted desktop with the environment, observed result, or specific pending prerequisite. Identify adapter-only coverage explicitly. |

Run the local acceptance suite in a fresh Cueia-only build context. Record missing dependencies, failed assertions, unexercised modes, and the fixture needed to reproduce each discrepancy. A scripted response supplies evidence for the scripted mode. Mark live-provider or device checks pending until exercised in their stated environment.

## Capabilities

### New Capabilities

- `deployment-modes` covers reproducible local, split-worker, and hosted operation.
- `runtime-telemetry` covers attribution, sampling, privacy settings, and opt-out behavior.

### Modified Capabilities

## Impact

Release acceptance consumes all eleven feature proposals. The release report links each measurement to its Cueia contract and fixture, so a later developer can repeat the run from this repository.
