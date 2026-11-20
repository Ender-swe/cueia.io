# Cueia

Cueia is a planned personal workspace for web, iOS, and Android. The owner will be able to delegate work in a conversation, close the app, and return to the saved result. For example, the agent can prepare a filled PDF from a mail attachment, show the owner the proposed reply and attachment, and send them after approval.

## Project status

This repository contains the OpenSpec proposals for a ground-up rebuild. The application is in the planning stage. The intended deliverable is a self-contained TypeScript workspace in `cueia/` that a developer can build, run, and test on a clean machine.

## Where to start

Begin with the foundation. Its first milestone is to sign in to a seeded sample workspace, restart the API and recover the same records, then sign out and verify that the old session is rejected. [Bootstrap proposal](openspec/changes/bootstrap-cueia/proposal.md)

The next milestone connects durable tasks, the agent runtime, and the first client so a delegated job survives closing the app. The roadmap covers all 12 proposals through release and identifies which features can be built in parallel. [Proposal order](PROPOSAL-ORDER.md)

With the OpenSpec CLI installed, run these commands from the repository root to list active changes and check which planning artifacts the foundation still needs.

```sh
openspec list
openspec status --change bootstrap-cueia
```
