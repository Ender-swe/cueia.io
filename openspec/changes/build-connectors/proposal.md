# Cueia mail, calendar, and action review

## Why

The owner can read a complete conversation, save an unsent draft, and inspect exactly what an approval will send or change. An approval issued for one account must never affect another.

## What Changes

### Connections and reads

In sample mode, the fictional adapter opens the workshop thread, attachment, and calendar event already provisioned for the signed-in workspace. Their IDs must agree with `GET /api/workspace`. In live mode, a signed-in owner can separately authorize Google mail or calendar access, whether they used Google or GitHub for Clerk sign-in. Bind short-lived, single-use callback state to the owner and requested grants. Validate the callback before exchanging the code. Store refreshable credentials encrypted on the server. Report account identity, connection generation, read/write grants, and reconnect-required state. Disconnect invalidates credentials and outstanding reviews for that connection. Reconnecting establishes a new generation even for the same address. [Bootstrap sessions and sample records](../bootstrap/proposal.md#access-and-storage)

Search mail and open the entire ordered thread. Return stable thread/message IDs, sender and recipients, subject, decoded readable bodies, timestamps when valid, and lazy attachment references. Keep attachment bytes out of list responses. Permit creating and editing unsent drafts. Distinguish sending a new message from replying to the selected thread. Reuse the attachment's workspace-scoped file ID for a managed document. For an ordinary workspace file selected for a draft, register the selected version. Bind its `fileId` to the review. PDF previews and dispatch read that ID even if the owner later edits the source. [Mail and file storage](../bootstrap/design.md#mail-and-file-storage)

List calendars and read events for a selected calendar and bounded time range. Preserve calendar ID, event ID, version, attendees, time zone, and start/end values. Represent all-day events with date-only values and an exclusive end date. Require an unambiguous offset or zone for timed events. Prepare event create, update, and delete actions. Obtain the current target and version for update/delete and expose conflicts before a mutation.

### Review ledger and receipts

Persist one review per logical action, whether initiated by chat, a task, or the client. Bind the action ID to its owner, connection generation, normalized payload and fingerprint, expiry, task ID when present, and target version. Store what the person sees, including recipient, message body, attachments or event details. Editing a prepared payload requires a fresh review.

Resolve decisions atomically. Re-read ownership, grant, expiry, connection generation, payload fingerprint, task eligibility, and target version before dispatch. Only the winner of the transition to executing may call the adapter. Repeated requests return the saved decision or in-progress state. Persist denial, expiry, confirmed success and provider ID, definitive failure, or uncertain delivery. If the connection drops after dispatch, expose uncertainty and withhold automatic retries. Surface the receipt in Activity and to the waiting task.

### Acceptance

1. Open the sample workshop from the adapter. Compare its thread, message, attachment, and event IDs with the owner's snapshot. Prepare a draft with an ordinary PDF from Files. Record its reviewed `fileId`, edit and register the source again, and verify it gets a different ID. Restart the API and approve concurrently from two sessions. Expect one provider invocation, one outgoing message with the first version's exact bytes, and one receipt. Reopen the receipt by action ID. Deny a second draft and verify it remains unsent.
2. Prepare a calendar update for an all-day event covering June 9 with an exclusive June 10 end. Change its stored version before approval. Expect a conflict, no provider mutation, and unchanged date semantics in the review.
3. Prepare a reply under account A, disconnect, and reconnect as B. The old review cannot send through B. Repeat after expiry and with read-only grants. Simulate a provider that accepts a send and loses the response. A duplicate approval must return the uncertain outcome with no second invocation.

## Capabilities

### New Capabilities

- `mail-calendar-connections` covers account connection, mailbox, drafts, and calendar access.
- `action-review` covers preparation, approval, execution, and receipts for external changes.

### Modified Capabilities

## Impact

Depends on workspace access and storage. Supplies account-scoped reads, drafts, and the shared action ledger to chat, calendar screens, durable tasks, and reviewed document replies. Use a fake clock, mailbox, calendar, and instrumented provider in automated tests.
