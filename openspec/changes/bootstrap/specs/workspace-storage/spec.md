# Workspace storage

## Purpose

Preserve private workspace records and document bytes across restarts, and give the API and worker a consistent view of each deployment's data.

## ADDED Requirements

### Requirement: One workspace per verified user

The system SHALL create the user and their private workspace atomically. Every user SHALL have one unique Clerk user ID and one unique workspace reference. Concurrent first sign-ins for the same subject SHALL resolve to one committed workspace.

#### Scenario: First sign-ins arrive together

1. **WHEN** two first sign-ins for Alice's Clerk subject arrive concurrently
2. **THEN** both succeed with the same workspace ID, and storage contains one user for that subject and one associated workspace.

#### Scenario: Provisioning is interrupted

1. **WHEN** initial workspace creation fails before commit and the owner signs in again
2. **THEN** the retry opens one complete workspace, with no extra workspace or partially provisioned user left by the failed attempt.

### Requirement: Private sample records and separate live data

Each new sample workspace SHALL receive one fictional mailbox thread, its fillable PDF attachment, one calendar event, and default preferences. Each copy SHALL have its own record identities. Later sign-ins SHALL retain that copy and its edits. Live workspaces SHALL be provisioned separately from sample data.

#### Scenario: Alice and Bob open separate samples

1. **WHEN** Alice opens two sessions on her first visit and changes a preference in one
2. **THEN** both sessions see the same single set of sample records, the same thread and attachment IDs, and the updated preference.
3. **WHEN** Bob first signs in
4. **THEN** his workspace, thread, attachment, and event have their own IDs, his attachment opens, and his preferences have the original defaults.

#### Scenario: Same person enters live mode

1. **WHEN** a person who has used sample mode opens a new live workspace
2. **THEN** the live workspace contains neither the sample fixtures nor changes made in that person's sample workspace.

### Requirement: Durable owner snapshots and document identity

Within a deployment the API and worker SHALL use the same durable database and address documents by stable IDs. Restarting PostgreSQL and the API with the same database and document volumes SHALL preserve owner-scoped snapshots, document metadata, and the exact bytes of each attachment.

#### Scenario: Reopen both owners' workspaces

1. **WHEN** the test records each owner's counts, IDs, preferences, attachment metadata and bytes, restarts PostgreSQL and the API, and reads both snapshots again
2. **THEN** those values match each owner's recorded values and the original attachment IDs still open the same bytes.

### Requirement: Publish complete documents

The system SHALL keep document storage private and expose a document only when its complete bytes and durable metadata agree. It SHALL reconcile interrupted writes before making the affected document available to readers.

#### Scenario: Interruption during document publication

1. **WHEN** a process stops after writing only part of a file or after publishing its bytes but before committing its record
2. **THEN** neither snapshots nor file reads expose an incomplete attachment.
3. **WHEN** recovery and a retry finish
4. **THEN** the owner sees one complete attachment under its committed ID and can read all its expected bytes.

### Requirement: Shared signing material

First startup SHALL initialize one durable set of signing material atomically per deployment. API processes sharing that storage SHALL agree on the material across concurrent startup and restart.

#### Scenario: Two API instances start on empty storage

1. **WHEN** two instances start together and the first issues a signed link
2. **THEN** the second can validate that link, including after a restart within the link's lifetime.

### Requirement: One winner for a versioned claim

The shared store SHALL accept a claim only if the record is eligible and its current version matches the expected version. Competing processes claiming the same eligible version SHALL produce exactly one successful update.

#### Scenario: Release two claimants together

1. **WHEN** two separate Go processes connected to the same Docker PostgreSQL database read the same eligible record and version and are released together to claim it
2. **THEN** one reports success, the other reports no update, and the persisted record shows a single claim and one version increment.

### Requirement: Migration gates service startup

Sample and live deployments SHALL use the same versioned SQL migrations and schema. On deployment startup, Compose SHALL wait for PostgreSQL health, run one uncached migration job, and start the API and worker only after that job succeeds.

#### Scenario: Fresh deployment and subsequent startup

1. **WHEN** a developer starts a fresh sample or live deployment and later starts it again on the same database
2. **THEN** each startup executes the migration job against the healthy database, applies only outstanding versions, and admits API and worker traffic or work after success.

#### Scenario: Migration fails

1. **WHEN** Compose starts with a deliberately failing fixture migration
2. **THEN** the migration job exits unsuccessfully, the API accepts no requests, and the worker accepts no work.
