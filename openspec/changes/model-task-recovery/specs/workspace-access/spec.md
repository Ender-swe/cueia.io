## ADDED Requirements

### Requirement: Versioned API access preserves existing clients

The API SHALL expose canonical `/v1/api` routes and retain `/api` compatibility aliases for existing clients. Both prefixes SHALL use the same feature handlers, origin policy, rate limiting, session exchange and revocation semantics, workspace authorization, and mutation idempotency scope. Newly generated clients and preview links SHALL use the canonical prefix. Existing signed preview links SHALL remain valid until their original expiry. API versioning SHALL NOT substitute for versioning persisted task checkpoints.

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
