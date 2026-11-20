# Repository Guidelines

## Project Structure

Cueia is in the planning stage. Start with `PROPOSAL-ORDER.md` to find the proposal, its dependencies, and its GitHub issue. Each active change has a directory under `openspec/changes/` containing a `proposal.md`. Accepted specifications belong in `openspec/specs/`, and completed changes go in `openspec/changes/archive/`.

The bootstrap proposal calls for a TypeScript workspace at the repository root. Cueia application source, tests, and assets have yet to be added. The separate `openfuse/` checkout has its own tooling. Create original Cueia artwork, fixtures, and expected results, and record their provenance.

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

Use the proposal’s acceptance scenarios as the basis for tests. For bootstrap, sign in twice, change a preference, restart the API, and verify that record IDs, attachment bytes, and the preference survive. Revoke one session and verify that only the other retains access. Cueia has yet to establish a test framework, filename convention, or coverage threshold.

## Commits and Pull Requests

The initial commit uses a scoped subject with an imperative summary. Follow that pattern, for example `docs(bootstrap): Clarify session recovery`. In a pull request, link the issue listed in `PROPOSAL-ORDER.md`, identify the affected proposal, and describe how you validated the change.
