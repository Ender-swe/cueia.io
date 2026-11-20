# Cueia web client

## Why

An owner can delegate work in a desktop or mobile browser, close the tab, and return to the saved result. Review screens must show the exact recipient, attachment, and outcome on both large and small screens.

## What Changes

### App shell and session lifecycle

Build a responsive React app in TypeScript in the root Bun workspace. Use Vite for local development and production builds, and React Router v7 for client-side navigation. Give each thread a URL the owner can revisit directly. Run the type check and production build through Turborepo. A later change will package the same web build for iOS and Android without Expo.

Use the TypeScript types generated from bootstrap’s OpenAPI contract for JSON requests, responses, and errors from the Go API. Configure only the API origin in the public bundle. Sign in through the API, restore the local session when the app opens, and reopen the owner’s workspace. On expiry, stop private data reads and return to sign-in. Signing out clears cached private views and credentials. Bundle original Cueia artwork, avatar variants, and accessible color choices locally.

Give the active conversation the primary reading area and keep the composer reachable above the keyboard. Use a navigation menu and sheets on narrow screens, accessible labels and visible focus on controls, and layouts that keep review details readable at desktop widths. Preserve drafts when switching screens. Maintain the reader’s scroll position when browsing older messages and provide a control to jump to new content. Preferences include avatar, tone, and background updates.

### Screens and interactions

| Destination | What the owner can do |
| --- | --- |
| Chat and Today | Open the main or a side thread, compose with attachments, stream Markdown, stop a reply, inspect queued messages, and return to current work. Send switches to Stop during a run and remains usable with an empty draft. |
| Mail and Calendar | Search and open a full thread, preview an attachment, edit a draft, select a calendar and date range, and prepare changes for review. Use platform date/time controls and retain all-day dates. |
| Documents | Pick or upload a file, open the original or filled PDF, page and zoom, inspect form fields, share through the platform, and open the review for its prepared reply. |
| Computer | Switch between browser, terminal, files, and an available desktop. Renew the browser console, reopen a profile, edit a file, inspect command receipts, and open the hosted desktop stream in the Computer screen. |
| Activity and reviews | Inspect progress, evidence, input requests, pause/resume/cancel controls, saved results, and receipts. Display the action payload and account before offering approval or denial. |
| Ideas, Goals, Tracking, and inbox | Review suggestions, open milestones and linked tasks, inspect watch status and observations, check now or change watch state, and follow a notification to its source. |
| Apps and connections | Search capabilities, inspect grants and account identity, connect or disconnect, and see provider availability. |

Render mail, search, browser, desktop, and sourced choice cards with their evidence and stable IDs. Open plans, comparisons, finance results, documents, and receipts directly from history. Fetch expiring links when needed. Show loading, empty, provider-disabled, disconnected, expired-session, conflict, and retry states where they arise. After an ambiguous external write, reopen its receipt before offering any further action.

### Browser acceptance

1. Run the workshop mail scenario in a desktop browser and in browsers on iOS and Android, starting each run with fresh sample state. Open the attachment, delegate the fill, close the tab, and let the worker finish. Open the thread URL in a new tab, preview the filled fields, review the exact recipient and attachment, approve, and reopen the receipt from chat. Verify one outgoing message with the reviewed document ID.
2. On a narrow viewport or device, type a draft, show the keyboard, switch threads and return, stop a streaming reply, and inspect pending messages. Verify the draft, tap targets, focus, and scroll-to-latest behavior. Reopen a historical tool card after its original link expires.
3. Disable a provider and check its unavailable view. Expire the session and reload a private route to verify the sign-in view. With a valid session, return to a backgrounded tab after a task completes and verify the refreshed result. Record the browser, device or viewport, and observed results for each run.

## Capabilities

### New Capabilities

- `workspace-clients` covers browser navigation, responsive screens, tool cards, and document handling.

### Modified Capabilities

## Impact

The shell can start after workspace access is available. The durable task, conversation, connector, document, browser, computer, and personal workspace APIs own the data displayed here. Use the stable identities returned by those APIs for navigation.
