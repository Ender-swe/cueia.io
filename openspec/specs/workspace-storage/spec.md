# workspace-storage Specification

## Purpose

Preserve each person's private workspace VM filesystem, records, and document identities across restarts so the owner and their agents can reopen the same files. In bootstrap acceptance scenarios, the internal provider represents each VM with a machine identity and each persistent disk with a separate directory.

## Requirements

### Requirement: One workspace per verified user

The system SHALL create the user and their private workspace atomically. Each user SHALL have one unique Clerk subject and one unique workspace reference per deployment. Each workspace SHALL have its own machine identity and persistent filesystem. The bootstrap sample deployment and acceptance tests SHALL use the internal workspace provider. Concurrent sign-ins and retries SHALL converge on that allocation before the workspace becomes available.

#### Scenario: First sign-ins arrive together

1. **WHEN** two first sign-ins for Alice's Clerk subject arrive concurrently
2. **THEN** both succeed with separate sessions for the same workspace and the deployment contains one user, one workspace, one VM, and one attached workspace disk for Alice.

#### Scenario: Provisioning is interrupted

1. **WHEN** initial provisioning fails before the user transaction commits, or the provider creates the VM but the caller stops before recording the result, and the owner signs in again
2. **THEN** the retry opens one complete workspace on a single VM and disk, and no session observes a partial seed set.

### Requirement: Private sample records and separate live data

Each new sample workspace SHALL receive one fictional mailbox thread, its fillable PDF attachment, one calendar event, and default preferences. Each copy SHALL have its own record identities. Later sign-ins SHALL retain that copy and its edits. Live workspaces SHALL be provisioned separately from sample data.

#### Scenario: Alice and Bob open separate samples

1. **WHEN** Alice opens two sessions on her first visit and changes a preference in one
2. **THEN** both sessions see the same single set of sample records, the same thread and attachment IDs, and the updated preference.
3. **WHEN** Bob first signs in
4. **THEN** his workspace, thread, attachment, and event have their own IDs, his attachment opens, and his preferences have the original defaults.

#### Scenario: Same person enters live mode

1. **WHEN** a person who has used sample mode opens a new live workspace
2. **THEN** the live workspace has a separate VM and disk and contains neither the sample fixtures nor files or preferences created in sample mode.

### Requirement: Browse the entire workspace filesystem

The Files sidebar SHALL list and search the owner's workspace filesystem on open or refresh, including nested files written directly by the owner or an agent. Browsing SHALL work without prior import or file registration. The sidebar SHALL distinguish an unavailable filesystem from an empty directory.

#### Scenario: Direct writes appear in Files

1. **WHEN** Alice creates a nested text file and an agent writes a CSV directly on her workspace VM, then Alice opens or refreshes Files in another session
2. **THEN** she can browse to both entries and find them by filename without importing them.
3. **WHEN** her VM becomes unreachable and she refreshes Files
4. **THEN** she sees that her files are unavailable and can retry when the machine returns.

### Requirement: Durable owner snapshots and document identity

The API and worker SHALL use the same database and resolve files on the owner's persistent workspace disk. Documents referenced by mail or preview SHALL have stable file IDs. An infrastructure restart or VM replacement using the retained disk SHALL preserve workspace identity, records, registered document bytes and ordinary files.

#### Scenario: Reopen both owners' workspaces

1. **WHEN** the test records each owner's counts, IDs, preferences, ordinary files, attachment metadata and bytes, restarts PostgreSQL, the API and the VMs with their disks retained, then replaces Alice's VM with one using her original disk
2. **THEN** those values match each owner's recorded values, directly created files remain readable, and the original attachment IDs still open the same bytes.

### Requirement: Publish complete documents

The system SHALL expose a managed document only when its complete bytes and committed metadata agree. It SHALL reconcile interrupted writes and restrict cleanup of abandoned files to its private staging area. It SHALL preserve ordinary workspace files regardless of whether they have database records.

#### Scenario: Interruption during document publication

1. **WHEN** an owner has an uncatalogued file on disk and a process stops after writing part of a managed file or publishing its bytes before the database commit
2. **THEN** snapshots, Files, and document reads withhold the incomplete or uncommitted managed document, and recovery preserves the owner's uncatalogued file.
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
