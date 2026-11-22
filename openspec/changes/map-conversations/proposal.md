# Cueia conversation threads and queued messages

## Why

An owner can queue a follow-up while a reply streams and switch conversations without losing their place. A restart must recover saved messages and tool results without sending an old draft.

## What Changes

### History and thread identity

Serve CopilotKit Rich Threads through the authenticated runtime using the server's Intelligence project key. Resolve the workspace from the active Cueia session. Give bootstrap's Main chat one durable conversation ID per workspace before its first send. The sample workshop mail shown there keeps its own mail thread ID as source evidence. Reopening Main chat from another session loads the same conversation. Populate Side chats with the owner's conversations and persist each side thread's distinct ID on first run. List and paginate owned threads, rename them, archive them out of the active list, and restore them with the same IDs. [Bootstrap chat layout](../archive/2026-11-22-bootstrap/design.md#chat-files-and-tasks)

Reconnect by thread ID and replay ordered rich messages, assistant text, tool call/result associations, task and document IDs, and prior selections. Store each file reference with the permanent [workspace ID](../archive/2026-11-22-bootstrap/design.md#resolve-workspace-ownership-at-the-storage-boundary) and stable file ID so a replay can open the referenced version after the workspace machine is replaced. Resolve the current task state and fresh resource links through authenticated endpoints. Keep the legacy single-conversation read/write API at `/api/conversation` isolated in owner storage. Creating, replaying, or archiving a Rich Thread must leave that legacy history intact.

### Composer lifecycle

Keep a draft, ordered removable follow-ups, and run status for each visited thread in the open app. A send on an idle thread starts a run. A send during a run adds the message to that thread’s visible queue. Switching threads restores the selected thread’s draft and queue, even while a different thread streams. Leave the composer editable. Advance only after the active turn has finished and its history is persisted. Dequeue each message once.

Stopping or receiving a run error pauses the queue, including errors delivered as stream events after a request was accepted. Keep the draft and unsent follow-ups available for removal or explicit resume. Stopping chat targets that conversation run. A previously delegated server task retains its own lifecycle. Closing the app discards volatile queues. Relaunch retrieves durable history and reconnects to the selected thread when available.

### Acceptance with a controllable history service

1. Sign in, send “first” on the main thread, queue “second” and “third,” remove “second,” and type an unsent draft. Open Files and Tasks, then restore the thread list and switch to a side conversation. Return and verify the draft and “third” belong to the main thread alone. Stop the first reply and verify no further sends. Resume, permit history persistence, and observe “third” exactly once.
2. Hold the persistence acknowledgement for a naturally completed reply. The queued send count stays zero until acknowledgement. Repeat with a streamed run error and require explicit resume. Close the app with another message pending and verify only saved turns appear after relaunch.
3. Delegate a task from Main chat, stop chat, and let the worker finish with a registered PDF result. Edit and register the source PDF again. After the first preview link expires, open Main chat in a second session. Verify the same conversation ID, completed task, and original `fileId`. The replayed card must obtain a fresh link that serves the first version's exact bytes.
4. Rename, archive, restore, and paginate with at least two threads. Verify the same message IDs and that another owner cannot read them. Seed one legacy message and confirm it remains readable through its own endpoint throughout.
5. Run these checks against a service double in CI. Record live Intelligence persistence and reconnect separately when a configured account is available.

## Capabilities

### New Capabilities

- `conversation-threads` covers thread identity, transcript replay, turn ordering, and control of pending messages.

### Modified Capabilities

## Impact

The runtime supplies streamed turns. The web client owns volatile drafts and queues. The history service owns saved threads, and durable task and file services own the objects linked from messages.
