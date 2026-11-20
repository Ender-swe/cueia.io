# Cueia reviewed document replies

## Why

A person can receive a fillable form, supply missing answers, inspect the completed copy, and authorize one reply. The original attachment, derived document, and send receipt each have a distinct identity.

## What Changes

### File intake and typed answers

Accept a PDF upload or import an attachment from an owned message. For mail imports, bind the document to the owner, account connection generation, message, and attachment ID. Repeat imports of that attachment within that connection reuse the document. Store the original bytes, stable file ID, display name, size, and source metadata. Authorize access to each file and issue expiring content links for preview and sharing.

Parse the PDF and expose page count, field names, field types, and existing values. Support text and checkbox fields. List other field types as unsupported. Reject unknown names, wrong types, malformed or encrypted PDFs, unsupported XFA forms, and content beyond decoding limits with a safe validation error. Ask for missing values through the task input interface, retain earlier answers, and validate new answers before generating bytes. OCR for scanned forms is outside this workflow.

### Prepare, review, and recover

1. Publish a new PDF with the validated answers and active actions removed. Record its parent file ID and the task operation that produced it. A retry with the same answers and source reuses the same filled-copy ID, including after file publication succeeds but checkpoint storage fails. Distinct inputs produce a distinct version. The original remains byte-for-byte readable.
2. Prepare a reply against the originating mail thread using the common action ledger. Store the recipient, reply text, account generation, and exact filled-copy ID. Move the task to waiting for review. Show the form preview, recipient, reply, and attachment together. Closing the app leaves the task waiting.
3. Resume from the stored decision or receipt. Approval permits one send of the reviewed payload. Denial, expiry, an account change, or task cancellation prevents a pending send. If delivery cannot be determined, display the uncertain receipt and preserve it through restart. Expose confirmed delivery in both task history and the mail thread.

### Acceptance with an independently authored form

Generate a two-page fictional workshop registration PDF containing a blank attendee name and an unchecked consent box. Attach it to a sample incoming message. Record the source bytes and begin with an empty outgoing mailbox.

- Enter a name and a boolean consent answer. Stop the worker after it publishes the filled copy and before it saves the next checkpoint. Resume with the client closed. Expect one derived copy. Parse both PDFs independently and verify the filled fields, original blank values, page count, and unchanged original bytes.
- Reopen the task, inspect the recipient and the preview, and count zero outgoing messages. Approve once. Interrupt before task completion, restart, and allow two more worker passes. Expect one outgoing reply in the original thread, the exact reviewed attachment bytes, one receipt, and a completed task.
- Repeat separately with denial, expiry, and an account switch before approval. Expect zero sends. Drop a response after simulated provider acceptance and verify that the app reports uncertainty across restarts. Submit a string as the checkbox answer and confirm the error occurs before creating a copy or review.

## Capabilities

### New Capabilities

- `reviewed-document-replies` covers form input, file lineage, approval of a prepared reply, and the resulting receipt.

### Modified Capabilities

## Impact

Workspace storage retains document bytes. Durable tasks own checkpoints and input requests. The mail connector owns account identity, approval decisions, and send receipts. Clients use the same file IDs for preview and review. Actual Google delivery is a separate credentialed acceptance run.
