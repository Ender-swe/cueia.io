# Workspace access

## Purpose

Give each signed-in person access to their private Cueia workspace and let them share a narrowly scoped document preview across API restarts.

## ADDED Requirements

### Requirement: API startup and provider availability

Sample and live APIs SHALL validate required Clerk settings and the server-side `CPK_INTELLIGENCE_API_KEY` before listening. Sample mode SHALL expose the service on loopback only. The API SHALL report its health and the actual availability of optional providers without disclosing credentials.

#### Scenario: A required value is missing

1. **WHEN** either mode starts with required settings missing
2. **THEN** startup identifies the missing settings by name and exits before opening an API listener.

#### Scenario: A configured sample deployment starts with Azure credentials unset

1. **WHEN** a developer starts a migrated sample deployment with the required API and Clerk settings, the internal workspace provider configured, and Azure settings unset
2. **THEN** the health check succeeds on loopback and the sample API has no externally published non-loopback listener.

#### Scenario: Optional provider states are distinguishable

1. **WHEN** an optional provider is unconfigured, configured and usable, or configured but known to be unavailable
2. **THEN** the reported state is respectively disabled, configured, or unavailable, and the response contains no provider secret.

### Requirement: Sign-in opens the owner's workspace

Both modes SHALL permit open signup through Clerk using an email address and password, Google, or GitHub. One “Open sample workspace” action SHALL use the current Clerk session or start sign-in and continue into that person's sample workspace.

#### Scenario: First visit and return visit

1. **WHEN** a signed-out visitor chooses “Open sample workspace”
2. **THEN** they can sign up with an email address and password, Google, or GitHub and open their own workspace without a second launch action.
3. **WHEN** the visitor chooses the action again with a valid Clerk session
4. **THEN** they reopen the same workspace without another sign-in prompt.

### Requirement: Exchange a verified Clerk session

`POST /api/session` SHALL verify the Clerk token and use its verified subject (`sub`) to identify the user. Each successful exchange SHALL issue a distinct expiring opaque Cueia bearer token. The persisted Cueia session credential SHALL consist of a digest with an expiry, and only the client receiving the exchange response SHALL receive the bearer value.

#### Scenario: Two exchanges for the same user

1. **WHEN** two exchanges present valid Clerk tokens for Alice's subject
2. **THEN** each receives its own Cueia bearer token and both can retrieve Alice's workspace with `GET /api/workspace`.

#### Scenario: Unauthenticated exchange in either mode

1. **WHEN** a caller sends an empty sign-in request or an invalid Clerk token in sample or live mode within the rate limit
2. **THEN** the exchange returns 401 and issues no Cueia session.

### Requirement: Authorize every private operation against its owner

The API SHALL require an active Cueia session for private routes and derive its workspace and machine from the session's user. Filesystem browsing, search, file registration, reads, mutations, and signed-link issuance SHALL be restricted to that owner's workspace. Paths and symlinks SHALL resolve within its workspace filesystem.

#### Scenario: Two owners request workspace records

1. **WHEN** Alice and Bob each use their session to request a collection or the workspace snapshot
2. **THEN** each response contains only records belonging to the requesting owner.

#### Scenario: The same path exists on two machines

1. **WHEN** Alice and Bob each create `reports/summary.csv` with different contents and access that path through their own sessions
2. **THEN** each can list and read only the file on their own VM.
3. **WHEN** Bob supplies Alice's file ID or tries a path or symlink outside his workspace root
4. **THEN** the API denies access and exposes no file from outside Bob's workspace.

#### Scenario: Bob presents

1. **WHEN** Bob requests a direct read, a change, or a signed preview link for Alice's attachment using his session
2. **THEN** the API denies each operation, returns neither Alice's bytes nor a link, and preserves her record.
3. **WHEN** Bob requests his own attachment
4. **THEN** the API returns his file.

### Requirement: Independent session lifetime

`DELETE /api/session` SHALL revoke the presenting Cueia session. A missing, expired, or revoked Cueia bearer SHALL receive 401 on private routes. Revoking one session SHALL preserve the authority of other active sessions, including those for the same user.

#### Scenario: Alice signs out after a restart

1. **WHEN** Alice revokes the first of two sessions after restarting the API
2. **THEN** a snapshot request with that bearer returns 401, while Alice's second bearer and Bob's bearer each open their respective workspaces.

#### Scenario: Session reaches expiry

1. **WHEN** the session expiry is reached and the client requests its workspace
2. **THEN** the API returns 401.

### Requirement: Web origin and unauthenticated request limits

The API SHALL return 403 for a disallowed web origin even with a valid session. It SHALL rate-limit unauthenticated attempts by connection within bounded windows and return 429 for throttled attempts.

#### Scenario: Disallowed origin

1. **WHEN** a request from a disallowed web origin presents a valid Cueia bearer
2. **THEN** the API returns 403 before serving private data or performing a mutation.

#### Scenario: Unauthenticated attempts exceed a window

1. **WHEN** unauthenticated attempts from a connection exceed the limit in a window
2. **THEN** excess attempts receive 429.
3. **WHEN** that window expires and the connection makes a new empty sign-in request
4. **THEN** normal authentication handling resumes and the request receives 401.

### Requirement: Resource-specific signed previews

A signed link SHALL authorize only its bound purpose, resource ID, and permitted operation until expiry. Each use SHALL check those bounds and the signature. A valid link SHALL remain usable through an API restart within its lifetime.

#### Scenario: Preview a single attachment

1. **WHEN** an owner obtains and opens a signed preview link for their PDF
2. **THEN** it returns that PDF's bytes.
3. **WHEN** the link is altered, used for another resource or purpose, or presented to list the workspace or modify a file
4. **THEN** access is denied and the file is unchanged.

#### Scenario: Preview a PDF discovered in Files

1. **WHEN** Alice selects an ordinary PDF on her VM for preview and obtains a signed link, then edits the source file
2. **THEN** that link still refers to the captured bytes until it expires, and opening it grants access only to that document version.

#### Scenario: Restart and expire

1. **WHEN** the API restarts on the same persistent storage and an unexpired preview link is opened
2. **THEN** the same attachment opens.
3. **WHEN** the fixture clock advances beyond link expiry
4. **THEN** the old link is rejected and a fresh link issued to the authorized owner opens the same bytes.

### Requirement: Documented API responses

The workspace API SHALL publish an OpenAPI contract for its HTTP routes, JSON inputs, responses, and error formats. Handlers SHALL validate inputs and return a structured error code and safe message for invalid input, missing resources, and unavailable dependencies.

#### Scenario: API failures follow the contract

1. **WHEN** a client sends an invalid payload, requests a missing resource, or calls an operation whose required dependency is unavailable
2. **THEN** each response matches its documented error status and schema, with a code and message safe to show a client.
