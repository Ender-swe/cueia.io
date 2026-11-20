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

Use lowercase, hyphen-separated change names such as `model-task-recovery`. Follow the existing proposal headings, including `Why`, `What Changes`, `Capabilities`, and `Impact`. Describe observable behavior and use numbered steps for acceptance scenarios.

## Testing Guidelines

Use the proposal’s acceptance scenarios as the basis for tests. For bootstrap, sign in twice, change a preference, restart the API, and verify that record IDs, attachment bytes, and the preference survive. Revoke one session and verify that only the other retains access. Cueia has yet to establish a test framework, filename convention, or coverage threshold.

## Commits and Pull Requests

The initial commit uses a scoped subject with an imperative summary. Follow that pattern, for example `docs(bootstrap): Clarify session recovery`. In a pull request, link the issue listed in `PROPOSAL-ORDER.md`, identify the affected proposal, and describe how you validated the change.
