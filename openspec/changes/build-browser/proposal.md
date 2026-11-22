# Cueia persistent browser sessions

## Why

An owner can read a page from chat, take over its browser, download a PDF, and return to the same profile later. A historical card must still explain what the agent saw even after the owner navigates elsewhere.

## What Changes

### Session and worker contract

Extend the Bun Playwright workspace in `apps/browser-worker` and package its matched Playwright and Chromium versions as a target of the root Dockerfile. Keep its endpoint private and authenticate control requests with a server-held bearer token. Its health endpoint reports process liveness. The Go API derives the workspace from the Cueia bearer and reserves a browser session UUID in that workspace before contacting the worker. It exposes active, closed, error, disabled, or unreachable state as appropriate. Save title, last URL, update time, downloads, and profile data. Close releases the browser process. Reopen uses the same profile and the owner's currently requested address. [Bootstrap repository layout](../archive/2026-11-22-bootstrap/design.md#repository-layout)

Provide operations to open, navigate, read, capture a screenshot, apply input, list/download PDFs, and close. A read returns the final URL, title, visible text up to 100,000 characters, and a truncation flag. Pair navigation with its read so simultaneous requests cannot attribute another page’s text to the caller. Store the observation with its original session, URL, title, and excerpt before returning a chat or task card.

### Destination checks, console, and files

- Allow public HTTP(S) pages on ports 80 and 443. Validate navigation, redirects, subrequests, and downloads. Reject non-public destinations, including mixed public/private DNS answers, and bind connections to an approved resolved address. A blocked request returns a visible failure.
- Bound the worker to three active sessions, 20 saved profiles, a 30-minute idle close, a 20-second navigation deadline, and 64 KiB request bodies. Keep profile data private and persistent across worker restarts.
- Issue an expiring owner-authorized console link. Support refresh without navigation, explicit reopen at an edited address, clicks scaled to screenshot pixels, supported keys, text entry, and scroll. Show disconnection, retain unsent text after failure, and suspend preview polling while hidden.
- Accept up to 20 PDFs per session and 10 MiB per PDF. Check file signature and actual byte count before publication, validate the PDF before document import, and preserve bytes. Keep profiles and pending downloads in private worker storage. Publish an accepted download on the owner's VM and record its committed `fileId` against the workspace, browser session, and download ID. Concurrent imports and retries return that ID. Files and the download card open the same immutable version, even if a later download reuses its filename. Return accepted files alongside rejected, oversized, or interrupted download outcomes. Recover incomplete transfers without exposing partial PDFs. [Bootstrap document publication](../archive/2026-11-22-bootstrap/design.md#publish-bytes-before-committing-their-database-references)

### Acceptance with real Chromium

1. Use a Cueia-authored pair of pages and PDF served by a test transport for an approved public destination. Keep production destination checks enabled. Open page A, set a profile preference, read it, then follow a link through the console to page B and download the PDF. The saved A card still names A and contains its excerpt.
2. Concurrently import the download twice. Verify both responses and its Files entry identify one `fileId`. Import a second PDF with the same filename and different bytes. Reopen both download cards and compare each preview with its original download. Interrupt another transfer, restart the worker with its profile volume, reopen the session, and verify the saved preference and interruption record.
3. Navigate to a fixture redirect ending at a private address and expect rejection. Present an expired console link, renew through the API, and verify the current page is unchanged. Disable the worker and confirm a read fails visibly without creating page evidence.

## Capabilities

### New Capabilities

- `browser-sessions` covers public-page reads, persistent profiles, console takeover, and download import.

### Modified Capabilities

## Impact

Workspace access owns authorization. The browser worker owns profiles and downloads. The workspace file service stores imported bytes on the owner's persistent disk. Chat, watches, and the Computer screen consume observations and session controls.
