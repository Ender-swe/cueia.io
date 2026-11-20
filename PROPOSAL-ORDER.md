# Cueia proposal order

Start with **bootstrap**. Write its specs and design, then its task list. The first milestone is a sample workspace you can sign in to, restart, and reopen with the same records. After sign-out, verify that the old session is rejected.

This recommended start order follows the [five-stage build plan](openspec/changes/bootstrap/proposal.md#build-order-and-ownership).

| Order | Stage | Proposal | Issue | Reason |
| --- | --- | --- | --- | --- |
| 1 | Foundation | [bootstrap](openspec/changes/bootstrap/proposal.md) | [#1](https://github.com/Ender-swe/cueia.io/issues/1) | Establish authentication, durable storage, and sample data. |
| 2 | First working client | [model-task-recovery](openspec/changes/model-task-recovery/proposal.md) | [#10](https://github.com/Ender-swe/cueia.io/issues/10) | Give delegated jobs durable state, cancellation, and restart recovery. |
| 3 | First working client | [build-agent-runtime](openspec/changes/build-agent-runtime/proposal.md) | [#2](https://github.com/Ender-swe/cueia.io/issues/2) | Run delegated tasks and stream their results. |
| 4 | First working client | [build-clients](openspec/changes/build-clients/proposal.md) | [#4](https://github.com/Ender-swe/cueia.io/issues/4) | Submit a job, close the app, and reopen its saved result. The shell can start after bootstrap and gain features as their APIs land. |
| 5 | Conversations and tools | [map-conversations](openspec/changes/map-conversations/proposal.md) | [#9](https://github.com/Ender-swe/cueia.io/issues/9) | Connect the composer and runtime to saved threads, queues, and replay. |
| 6 | Conversations and tools | [build-connectors](openspec/changes/build-connectors/proposal.md) | [#6](https://github.com/Ender-swe/cueia.io/issues/6) | Supply mail, calendar, and the approval ledger needed for reviewed document replies. |
| 7 | Conversations and tools | [build-browser](openspec/changes/build-browser/proposal.md) | [#3](https://github.com/Ender-swe/cueia.io/issues/3) | Supply page reads for watches and PDF downloads for Documents. |
| 8 | Conversations and tools | [build-computer](openspec/changes/build-computer/proposal.md) | [#5](https://github.com/Ender-swe/cueia.io/issues/5) | Add persistent workspace files and desktop access. This can proceed alongside the browser. |
| 9 | Personal workflows | [build-personal-workspace](openspec/changes/build-personal-workspace/proposal.md) | [#7](https://github.com/Ender-swe/cueia.io/issues/7) | Add saved results, goals, and the inbox that receives watch alerts. |
| 10 | Personal workflows | [trace-document-review](openspec/changes/trace-document-review/proposal.md) | [#12](https://github.com/Ender-swe/cueia.io/issues/12) | Combine durable tasks, mail, PDF handling, and approval to send one reviewed reply. |
| 11 | Personal workflows | [characterize-watches](openspec/changes/characterize-watches/proposal.md) | [#8](https://github.com/Ender-swe/cueia.io/issues/8) | Combine task scheduling, page reads, and inbox delivery. This can proceed alongside document review. |
| 12 | Release | [ship](openspec/changes/ship/proposal.md) | [#11](https://github.com/Ender-swe/cueia.io/issues/11) | Run clean-machine builds and check parity across the completed workflows. |
