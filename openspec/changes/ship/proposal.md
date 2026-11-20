# Cueia release acceptance

## Why

A developer starting with only a fresh Cueia checkout needs to build the product, exercise its workflows, and identify precisely which platforms and live services were verified.

## What Changes

### Distribution and deployment

Provide root install, build, migration, start, test, and platform export commands. Pin Bun, install with `bun install --frozen-lockfile`, and run lint plus Turbo type checks, tests, and builds for the API, workers, and clients. Include operator, contributor, security, and license documentation, configuration examples with placeholders, and provenance for distributed assets, dependencies, and fixtures. Build tests and expected results from Cueia’s own requirements. Use only declared build inputs.

Ship the web bundle, iOS/Android export commands, Render blueprint, and root multi-stage Dockerfile and Compose configuration. A configured `docker compose up` starts healthy PostgreSQL, applies migrations once, and gates API and task-worker startup on success. Package the optional Chromium service and Linux computer image, keep worker controls private, persist database and file/profile volumes across restarts, and document shared storage for split workers. Report health and capability status when optional providers are absent. Include clean shutdown, restart, and recovery instructions.

Provide a local demo launcher with isolated fictional data, a scripted model, and fixture transports for network services. Keep telemetry disabled and production accounts inaccessible in the automated fixtures. Exercise a fictional mail request that leads to two sourced browser choices, and a separate fillable workshop form that leads to one reviewed reply. Document the required server-side Intelligence project configuration for a normal app launch and identify when tests use a service double.

### Telemetry and operational evidence

Use the CopilotKit SDK transport with Cueia app attribution and server-side project binding through `CPK_TELEMETRY_ID`. Keep the Intelligence API key separate and private. Default the sampling rate to 1 and honor an explicit `COPILOTKIT_TELEMETRY_SAMPLE_RATE`. Honor `COPILOTKIT_TELEMETRY_DISABLED` and `DO_NOT_TRACK` when set to 1 or true. Exercise runtime initialization and request events against an intercepted transport and verify the attribution, project binding, sampling, both opt-outs, and absence of the API key. Suppress transport in ordinary fixtures. Log operation, status, time, and correlation ID with credentials and private workspace content redacted.

### Acceptance inventory

For each public route, client screen, task type, adapter/mode, and build entry point, name its owning Cueia proposal, setup, authored fixture, steps, and observable expectation. Record the actual result, environment, and evidence from the run. Include failure cases. Compare generated IDs by their relationships and timestamps by ordering or fixture clock. Keep byte equality, event counts, and permissions exact.

| Run | Required evidence |
| --- | --- |
| Fresh root install and build | Frozen install succeeds, lint/type checks pass, server and workers build, and web, iOS, and Android bundles export. |
| Embedded and split deployments | PGlite sample data survives restart. Two Bun workers contend for one due PostgreSQL task and only one invokes its handler. A failed migration prevents both services from starting. |
| Browser and local computer | Real Chromium and the browser container retain the fixture profile after restart and import one PDF. A disposable Docker computer retains file bytes and records an interrupted command. |
| Document and personal workflows | Start with zero outgoing messages. Approve a filled PDF once, restart, and observe one reply and one receipt. Reopen a finance summary, linked milestone, and watch alert using the expected records from their proposals. |
| Client sessions | Record interactive web, iOS, and Android journeys with platform, device or simulator, and observed results. Report bundle exports independently of installed-device checks. |
| Configured external services | Record Google authorization and reviewed delivery, Intelligence replay, each selected model/gateway, live Jev, Parallel search, and hosted desktop with the environment, observed result, or specific pending prerequisite. Identify adapter-only coverage explicitly. |

Run the local acceptance suite in a fresh Cueia-only build context. Record missing dependencies, failed assertions, unexercised modes, and the fixture needed to reproduce each discrepancy. A scripted response supplies evidence for the scripted mode. Mark live-provider or device checks pending until exercised in their stated environment.

## Capabilities

### New Capabilities

- `deployment-modes` covers reproducible local, split-worker, and hosted operation.
- `runtime-telemetry` covers attribution, sampling, privacy settings, and opt-out behavior.

### Modified Capabilities

## Impact

Release acceptance consumes all eleven feature proposals. The release report links each measurement to its Cueia contract and fixture, so a later developer can repeat the run from this repository.
