## ADDED Requirements

### Requirement: Versioned API access preserves existing clients

The API SHALL expose canonical `/v1/api` routes and retain `/api` compatibility aliases for existing clients. Both prefixes SHALL use the same feature handlers, origin policy, rate limiting, session exchange and revocation semantics, workspace authorization, and mutation idempotency scope. Newly generated clients and preview links SHALL use the canonical prefix. Existing signed preview links SHALL remain valid until their original expiry. API versioning SHALL NOT substitute for versioning persisted task checkpoints.

Both prefixes SHALL allow `Idempotency-Key` alongside `Authorization` and `Content-Type` in CORS preflight responses for the existing authorized origin. This SHALL NOT broaden the set of allowed origins. Browser retries after an ambiguous network failure SHALL retain the original key and payload, including expected execution/version, rather than silently issuing a new mutation.

#### Scenario: Exchange and revoke across prefixes

1. **WHEN** Alice exchanges a valid Clerk session through `/v1/api/session`
2. **THEN** the resulting Cueia bearer can read her workspace through both `/v1/api/workspace` and `/api/workspace`.
3. **WHEN** she revokes it through `/api/session`
4. **THEN** private requests through either prefix return 401.

#### Scenario: Security is identical across mounts

1. **WHEN** a disallowed origin or foreign-resource request is sent through either prefix
2. **THEN** both enforce the existing origin and owner restrictions before returning data or changing state.
3. **WHEN** unauthenticated requests alternate prefixes to exceed the connection rate limit
4. **THEN** changing the prefix does not bypass throttling.

#### Scenario: Existing preview and mutation replay

1. **WHEN** a signed `/api/previews/...` URL issued before the versioned API rollout is opened within its lifetime
2. **THEN** it still returns only its bound document version.
3. **WHEN** a task mutation committed through `/v1/api` is repeated through `/api` with the same key and payload
4. **THEN** it returns the same saved response without another mutation.

#### Scenario: Contract and generated clients

1. **WHEN** a developer regenerates TypeScript API types from the OpenAPI contract
2. **THEN** the canonical versioned task, conversation, and existing workspace routes have documented request, response, pagination, and error schemas, and compatibility aliases are identified in the contract.

#### Scenario: Cross-origin browser retries a lost response

1. **WHEN** the authorized cross-origin browser preflights a task mutation with `Idempotency-Key`, `Authorization`, and `Content-Type` through either prefix
2. **THEN** the preflight permits those headers and the authenticated mutation can proceed.
3. **WHEN** task creation commits through one prefix, its response is lost, and the browser retries the same key and payload through the other
4. **THEN** it receives the original task ID with one saved task and execution.
5. **WHEN** a disallowed origin attempts the same preflight or mutation
6. **THEN** it is rejected without changing task state.
