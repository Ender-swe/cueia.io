# Cueia reviewed document replies

## Why

A person can receive a fillable form, supply missing answers, inspect the completed copy, and authorize one reply. The original attachment, derived document, and send receipt each have a distinct identity.

## What Changes

### File intake and typed answers

Accept a PDF upload, a workspace PDF selected from Files, or an attachment from an owned message. For an ordinary PDF, register the version at its normalized workspace-relative path. Concurrent registrations of the same path and bytes return one `fileId`. An edit yields a new version on registration. For a managed attachment, reuse the `fileId` already shared by the mail message and its Files entry. Save that input ID at intake so parsing, filling, and recovery all read the same bytes. [Bootstrap PDF registration](../archive/2026-11-22-bootstrap/design.md#register-a-pdf-from-files)

For mail imports, bind the document to the owner, account connection generation, message, and attachment ID. Repeat imports of that attachment within that connection reuse the document. Keep original and filled bytes on the owner's persistent VM disk. Store each version's ID, display name, size, and source metadata in PostgreSQL. Issue an expiring read-only link for the selected version when the owner previews or shares it. A send requires a separate ledger decision.

Parse the PDF and expose page count, field names, field types, and existing values. Support text and checkbox fields. List other field types as unsupported. Reject unknown names, wrong types, malformed or encrypted PDFs, unsupported XFA forms, and content beyond decoding limits with a safe validation error. Ask for missing values through the task input interface, retain earlier answers, and validate new answers before generating bytes. OCR for scanned forms is outside this workflow.

### Prepare, review, and recover

1. Publish a new PDF with the validated answers and active actions removed. Record its parent file ID and the task operation that produced it. A retry with the same answers and source reuses the same filled-copy ID, including after file publication succeeds but checkpoint storage fails. Distinct inputs produce a distinct version. The original remains byte-for-byte readable.
2. Prepare a reply against the originating mail thread using the common action ledger. Store the recipient, reply text, account generation, and exact filled-copy ID. Move the task to waiting for review. Show the form preview, recipient, reply, and attachment together. Closing the app leaves the task waiting.
3. Resume from the stored decision or receipt. Approval permits one send of the reviewed payload. Denial, expiry, an account change, or task cancellation prevents a pending send. If delivery cannot be determined, display the uncertain receipt and preserve it through restart. Expose confirmed delivery in both task history and the mail thread.

### Acceptance with an independently authored form

Use the signed-in owner's bootstrap workshop invitation and its attached form. Author the shared sample PDF with two pages, a blank attendee name, and an unchecked consent box. Record its expected bytes and field values independently and begin with an empty outgoing mailbox.

1. Open the attachment from Mail and Files and check both identify the same original. Enter a name and a boolean consent answer. Stop the worker after it publishes the filled copy and before it saves the next checkpoint. Resume with the client closed. Refresh Files from a second session and expect one derived copy alongside the original. Check its parent ID and parse both PDFs independently to verify the filled fields, original blank values, page count, and original bytes.
2. Reopen the task, inspect the recipient and the filled-copy preview, and count zero outgoing messages. Approve once. Interrupt before task completion, restart, and allow two more worker passes. Expect one outgoing reply in the original thread, the exact reviewed attachment bytes, one receipt, and a completed task.
3. Prepare another reply to the workshop invitation, this time selecting an ordinary copy of the blank PDF from Files as the input. Record its `fileId`, fill the form, and record the filled-copy `fileId` shown in review. Edit and register the ordinary source again, then restart. Open each of the three IDs. The original still matches the blank fixture, the review still opens the exact filled copy, and only the newest ID returns the edited source bytes.
4. Repeat the mail review separately with denial, expiry, and an account switch before approval. Expect zero sends. Drop a response after simulated provider acceptance and verify that the app reports uncertainty across restarts. Submit a string as the checkbox answer and confirm the error occurs before creating a copy or review.

## Capabilities

### New Capabilities

- `reviewed-document-replies` covers form input, file lineage, approval of a prepared reply, and the resulting receipt.

### Modified Capabilities

## Impact

Workspace storage retains document bytes. Durable tasks own checkpoints and input requests. The mail connector owns account identity, approval decisions, and send receipts. Clients use the same file IDs for preview and review. Actual Google delivery is a separate credentialed acceptance run.
