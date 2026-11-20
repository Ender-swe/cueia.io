# Bootstrap tasks

## 1. Build and contract

- [ ] 1.1 Create the pinned Go module and API, worker, and migration entry points. Commit `go.mod` and `go.sum`. Verify a clean module download, build, and `go test ./...` with the pinned toolchain.
- [ ] 1.2 Set up the root TypeScript workspace with Bun pinned in `package.json`, one tracked `bun.lock`, and root Turbo build, type-check, and test tasks covering Go and TypeScript. Document the root commands and run them after `bun install --frozen-lockfile` in a clean checkout.
- [ ] 1.3 Write the OpenAPI contract for the bootstrap routes and error responses. Generate the React client's TypeScript types from it. Verify generation is reproducible and that changing a response without regenerating its types fails the contract check.

## 2. Shared database and deployment

- [ ] 2.1 Implement the versioned SQL migration command and shared `pgx` storage configuration. Create workspaces, users, sessions, sample record tables, file metadata, signing material, and the minimal versioned claim record. Run migrations twice on Docker PostgreSQL. Verify each version applies once and that duplicate or null Clerk IDs, duplicate or null user workspace references, and dangling workspace references are rejected.
- [ ] 2.2 Build the three Go service targets from a root multi-stage `Dockerfile`. Add root `compose.yaml` and document how to start each mode. Inspect both resolved deployments for separate persistent volumes between modes, a shared database and private document directory within each, and loopback-only host publication in sample mode.
- [ ] 2.3 Gate migration on database health and gate API and worker startup on that migration's success. Run two successive deployment starts on retained volumes. Verify each executes one fresh uncached migration job and each job finishes before either application service starts.
- [ ] 2.4 Implement the conditional claim update shared by Go processes. Against PostgreSQL, check success for an eligible row at the expected version, zero updates for an ineligible or stale row, and exactly one version increment on success.

## 3. Provisioning and document durability

- [ ] 3.1 Author the original sample mail thread, fillable PDF, calendar event, default preferences, and independent expected results. Open the PDF and verify its form fields. Record provenance for Cueia code, fixtures, artwork used, and expectations, and list third-party libraries, redistributed dependencies, and service protocols with their sources.
- [ ] 3.2 Implement private file staging, atomic publication, and metadata persistence. Interrupt writes before rename and between rename and database commit. Run recovery and retry against the same directory. Verify that readers never receive a partial attachment, abandoned files are reconciled, and the recovered bytes match the fixture.
- [ ] 3.3 Provision the user, workspace, and complete sample record set in one transaction keyed by verified Clerk subject. Race two first visits and inject a rollback before commit. Query the database after retry for one user, one workspace, one complete seed set, and no orphan workspace. Provision the same subject in live mode and verify the sample thread, attachment, and event are absent.
- [ ] 3.4 Implement owner-scoped snapshot, collection, preference, and attachment storage operations. Test with two owners. Verify one owner's preference update is immediately visible to their other session's workspace lookup, each list contains only that owner's records, and a foreign record ID cannot read or change another owner's data.

## 4. HTTP access and signing

- [ ] 4.1 Validate required Clerk settings and server-side `CPK_INTELLIGENCE_API_KEY` before listening in both modes. Implement health and optional provider status. Test missing values by name, sample loopback binding, and disabled, configured, and unavailable provider responses with a controlled provider. Check responses and diagnostics for secret values.
- [ ] 4.2 Implement `POST /api/session` with Clerk verification and opaque bearer issuance after commit. Test empty and invalid tokens for 401 in both modes. Exchange twice for one verified `sub` and check for distinct working bearers, digest-only credential storage, and persisted expiries.
- [ ] 4.3 Authenticate private `net/http` handlers, including `GET /api/workspace` and `DELETE /api/session`. Wire collection, preference, and attachment handlers to the session-derived workspace. Test missing, expired, and revoked bearers for 401, independent revocation, and owner isolation through the HTTP routes.
- [ ] 4.4 Enforce origin checks, bounded unauthenticated rate limits by connection peer, and input validation. Test a disallowed origin with a valid bearer for 403 and a burst of unauthenticated requests for 429 followed by normal handling in the next window. Compare invalid-input, missing-resource, and unavailable-dependency responses with the OpenAPI error codes and safe messages.
- [ ] 4.5 Atomically initialize the deployment signing key and implement owner-authorized preview issuance and redemption. Start two API instances together on an empty key store and redeem a link across instances. Test wrong resource, purpose, operation, signature, and expiry. Verify a preview alone cannot list a workspace or change a file.

## 5. Sample client and sign-in

- [ ] 5.1 Build the Vite React client with React Router v7 and the generated types. Implement “Open sample workspace”, the Clerk return path, snapshot display, preference edit, attachment preview, and session sign-out. Exercise both launch paths with the Playwright browser worker running under Bun. Verify that one action reaches the workspace after sign-in and that an active Clerk session skips the prompt.
- [ ] 5.2 Document sample and live Clerk setup, server configuration, and the sample launch command. On a fresh Clerk test configuration, verify open signup offers only Google and GitHub and the built client contains no server credentials.

## 6. Fresh-install acceptance

Use one fresh sample deployment for cases 1–3. Retain both owners' IDs, counts, preferences, attachment metadata and bytes, three Cueia sessions, and an unexpired preview link across the restart. Keep token values out of test output. These cases follow the [proposal's acceptance sequence](proposal.md#acceptance-on-a-fresh-installation).

- [ ] 6.1 **Case 1 — Alice's first visit.** From the Cueia checkout alone, install from the frozen lockfile, build from the root, configure Clerk and the sample API, start Compose, and check health and an empty exchange (401). Sign Alice in through Google. Release two session exchanges together before either can provision her. Confirm one workspace and one seed set with matching thread and attachment IDs. Change a preference in session A, read it in B, and record Alice's snapshot and file bytes.
- [ ] 6.2 **Case 2 — Bob's workspace.** Sign Bob in through GitHub. Check for a distinct workspace, thread and attachment IDs, and the default preference. With Bob's bearer, request a direct read and a preview link using Alice's attachment ID. Verify both fail, then open Bob's own attachment and record his snapshot and bytes.
- [ ] 6.3 **Case 3 — Recovery and revocation.** Issue Alice a preview link, restart PostgreSQL and the API on the same database and document volumes, and compare both owners' counts, IDs, preferences, metadata, and exact bytes. Verify the pre-restart link still opens. Revoke Alice's session A and check its next snapshot returns 401 while B and Bob's session each open the correct workspace. Advance the fixture clock past link expiry while B is valid. Check the old link fails and a fresh link obtained with B returns the original PDF bytes.
- [ ] 6.4 **Case 4 — Competing processes.** Start two separate Go processes on the same Docker PostgreSQL database. Wait until both have read the same eligible row and version, release them together, and assert one successful update, one loser, and one persisted version increment.
- [ ] 6.5 **Case 4 — Failed deployment.** Start Compose against an isolated database with a deliberately failing fixture migration. Verify a nonzero migration exit, no API listener, and no running worker or claimed work. Exercise the startup gate in both sample and live mode.
