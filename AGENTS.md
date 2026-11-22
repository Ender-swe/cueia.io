# Repository Guidelines

## Project Structure

Start with `PROPOSAL-ORDER.md` to find the proposal, its dependencies, and its GitHub issue. Each active change has a directory under `openspec/changes/` containing a `proposal.md`. Accepted specifications belong in `openspec/specs/`, and completed changes go in `openspec/changes/archive/`.

The repository root contains the TypeScript workspace. The Go server lives in `apps/server/`. The separate `openfuse/` checkout has its own tooling. Create original Cueia artwork, fixtures, and expected results, and record their provenance.

### Go server file layout

Use the session feature as the template for new Go features. Group private feature code under `internal/<feature>/` and create each file when the feature needs it.

```text
apps/server/
  cmd/api/
    main.go               # Wire dependencies, call routes.New, start and stop the server
    test/
      main_test.go
    routes/
      routes.go           # Assemble feature routes and apply shared middleware
      sessions.go         # Bind HTTP methods and paths to handlers with route-specific auth
      test/
        routes_test.go
        sessions_test.go
  internal/session/
    session.go            # Define the feature's domain structs, such as Credential
    handler.go            # Read HTTP requests, call the service, and write responses
    service.go            # Coordinate session exchange, authentication, and revocation
    memory.go             # Keep all session database queries here
    config.go             # Load session-specific configuration
    test/
      config_test.go
      handler_test.go
      service_test.go
      memory_test.go
      issuer.go           # Token issuer used by the session tests
```

Add endpoints to the owning feature's file in `cmd/api/routes/` using paths relative to the shared `/api` prefix. After registering all feature routes, `routes.New` mounts their mux at `/api/` and strips the prefix before dispatch. For example, `sessions.go` registers `POST /session`, and clients call `POST /api/session`.

Name the domain struct file after the feature, as in `session/session.go`. The session route invokes `ExchangeHTTP` in `handler.go`. The handler calls `Service.Exchange` in `service.go`, which uses `createSession` in `memory.go` to persist the session.

Put every Go test in the closest `test/` folder to the code it exercises. Create that folder directly under the owning package when needed. This applies to unit and integration tests alike. A session persistence test belongs in `internal/session/test/`, and a session route registration test belongs in `cmd/api/routes/test/`.

## Development Commands

Run the OpenSpec CLI from the repository root:

- `openspec list` lists active changes.
- `openspec status --change bootstrap` shows which foundation artifacts are ready and which still need work.
- `openspec validate bootstrap --strict` checks the foundation change against the schema. Run it after adding or revising specification deltas.

Application build, run, and test commands will arrive with bootstrap. The proposal specifies a pinned Bun version, installation with `bun install --frozen-lockfile`, and root Turbo tasks for builds, type checks, and tests.

## Style and Naming

- Always read `DESIGN.md` before creating or modifying UI components, styles, or layouts.
- Follow all color, typography, spacing, and component rules in `DESIGN.md`.
- Reuse existing UI primitives before creating new ones.
  Use lowercase, hyphen-separated change names such as `model-task-recovery`. Follow the existing proposal headings, including `Why`, `What Changes`, `Capabilities`, and `Impact`. Describe observable behavior and use numbered steps for acceptance scenarios.

Use the [uncodixfy skill](.agents/skills/uncodixfy/SKILL.md) whenever creating or editing frontend UI.

### Illustrations

When asked for an illustration in a design document or in chat, draw it in ASCII inside a fenced `text` block. For a screen, use `+`, `-`, and `|` to outline the relevant panes, place labels and controls where they appear, and fill in realistic content for the state being discussed.

```text
+---------------+------------------------+
| Files         | File preview           |
+---------------+------------------------+
| > agenda.txt  | Workshop agenda        |
|   notes.txt   | 09:00  Introductions   |
|               | 09:15  Sketch review   |
| [Upload]      | [Download]             |
+---------------+------------------------+
```

### Database schemas

In design documents, group related tables in fenced `text` blocks. Start each table with its name and a colon. List every column, including timestamps, on its own indented line. Align the column name, PostgreSQL type, and constraints. Mark keys with `PK`, `FK -> table.column`, or `UNIQUE`, and show nullability where specified. Put composite constraints below the columns and list candidate indexes separately with their purpose.

When asked to show a schema in chat, give one relevant table as an example and link to the full schema in the design. Show the full schema in chat when explicitly requested.

```text
users:
  id             UUID          PK
  clerk_user_id  TEXT          UNIQUE, NOT NULL
  workspace_id   UUID          UNIQUE, NOT NULL, FK -> workspaces.id
  created_at     TIMESTAMPTZ
```

## Testing Guidelines

Use the proposal’s acceptance scenarios as the basis for tests. For bootstrap, sign in twice, change a preference, restart the API, and verify that record IDs, attachment bytes, and the preference survive. Revoke one session and verify that only the other retains access.

## Commits and Pull Requests

The initial commit uses a scoped subject with an imperative summary. Follow that pattern, for example `docs(bootstrap): Clarify session recovery`. In a pull request, link the issue listed in `PROPOSAL-ORDER.md`, identify the affected proposal, and describe how you validated the change.
