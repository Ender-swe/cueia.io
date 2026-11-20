# Cueia native and web clients

## Why

An owner can delegate work on a phone or in a browser, close the client, and return to the saved result. Review screens must make the exact recipient, attachment, and outcome visible on all three platforms.

## What Changes

### App shell and session lifecycle

Build an Expo/React Native client for iOS, Android, and web in the root Bun workspace. Run type checks and platform exports through Turborepo. Configure only the API origin in the public bundle. Sign in through the API, restore the local session on launch, and reopen the owner’s workspace. On expiry, stop private data reads and return to sign-in. Signing out clears cached private views and credentials. Bundle original Cueia artwork, avatar variants, and accessible color choices locally.

Give the active conversation the primary reading area and keep the composer reachable above the keyboard. Use a navigation menu and sheets on narrow screens, accessible labels and visible focus on controls, and layouts that keep review details readable at desktop widths. Preserve drafts when switching screens. Maintain the reader’s scroll position when browsing older messages and provide a control to jump to new content. Preferences include avatar, tone, and background updates.

### Screens and interactions

| Destination | What the owner can do |
| --- | --- |
| Chat and Today | Open the main or a side thread, compose with attachments, stream Markdown, stop a reply, inspect queued messages, and return to current work. Send switches to Stop during a run and remains usable with an empty draft. |
| Mail and Calendar | Search and open a full thread, preview an attachment, edit a draft, select a calendar and date range, and prepare changes for review. Use platform date/time controls and retain all-day dates. |
| Documents | Pick or upload a file, open the original or filled PDF, page and zoom, inspect form fields, share through the platform, and open the review for its prepared reply. |
| Computer | Switch between browser, terminal, files, and an available desktop. Renew the browser console, reopen a profile, edit a file, inspect command receipts, and hand a native desktop session to the system browser. |
| Activity and reviews | Inspect progress, evidence, input requests, pause/resume/cancel controls, saved results, and receipts. Display the action payload and account before offering approval or denial. |
| Ideas, Goals, Tracking, and inbox | Review suggestions, open milestones and linked tasks, inspect watch status and observations, check now or change watch state, and follow a notification to its source. |
| Apps and connections | Search capabilities, inspect grants and account identity, connect or disconnect, and see provider availability. |

Render mail, search, browser, desktop, and sourced choice cards with their evidence and stable IDs. Open plans, comparisons, finance results, documents, and receipts directly from history. Fetch expiring links when needed. Show loading, empty, provider-disabled, disconnected, expired-session, conflict, and retry states where they arise. After an ambiguous external write, reopen its receipt before offering any further action.

### Cross-platform acceptance

1. Use the same fictional workshop mail on each platform with fresh sample state. Open the attachment, delegate the fill, close the app, and finish processing on the worker. Relaunch, preview the filled fields, review the exact recipient and attachment, approve, and reopen the receipt from chat. Verify one outgoing message with the reviewed document ID.
2. On a narrow viewport or device, type a draft, show the keyboard, switch threads and return, stop a streaming reply, and inspect pending messages. Verify the draft, tap targets, focus, and scroll-to-latest behavior. Reopen a historical tool card after its original link expires.
3. Disable a provider, expire the session, and return from the background with a task newly completed. Check the associated unavailable, sign-in, and refreshed-result views. Record web interaction, iOS device/simulator interaction, Android device/emulator interaction, and platform bundle exports as distinct results.

## Capabilities

### New Capabilities

- `workspace-clients` covers cross-platform navigation, screens, tool cards, and document handling.

### Modified Capabilities

## Impact

The shell can start after workspace access is available. The durable task, conversation, connector, document, browser, computer, and personal workspace APIs own the data displayed here. Use their shared validation and stable identities for navigation.
