# Cueia web client

## Why

An owner can delegate work in a desktop or mobile browser, close the tab, and return to the saved result. Review screens must show the exact recipient, attachment, and outcome on both large and small screens.

## What Changes

### App shell and session lifecycle

Extend bootstrap's Vite dashboard in `apps/dashboard` under its dark Adobe React Spectrum 2 provider. Connect the Main chat composer to the runtime and populate the existing Tasks panel from the scheduler. Files still replaces the thread list on the left. The right-hand Tasks panel can stay open while the owner browses files or writes a draft. Give each saved thread a React Router v7 URL the owner can revisit directly. Run the type check and production build through root Turbo tasks. A later change will package the same web build for iOS and Android without Expo. [Bootstrap chat layout](../bootstrap/design.md#chat-files-and-tasks)

Use `packages/api-types`, generated from `backend/api/openapi.yaml`, for JSON requests, responses, and errors. Use bootstrap's sign-in flow for the new screens. Exchange the Clerk credential at `POST /api/session`, keep the Cueia bearer in memory, and load the owner's snapshot from `GET /api/workspace`. On return, exchange again using the existing Clerk session. On expiry, stop private reads and return to sign-in. Sign out with `DELETE /api/session` and clear this client's cached private views and credentials. Expose the API origin and Clerk publishable configuration in the public bundle. Bundle original Cueia artwork, avatar variants, and accessible color choices locally. [Bootstrap API routes](../bootstrap/proposal.md#bootstrap-api-routes)

Give the active conversation the primary reading area and keep the composer reachable above the keyboard. Use a navigation menu and sheets on narrow screens, accessible labels and visible focus on controls, and layouts that keep review details readable at desktop widths. Preserve drafts when switching screens. Maintain the reader’s scroll position when browsing older messages and provide a control to jump to new content. Preferences include avatar, tone, and background updates.

### Screens and interactions

| Destination | What the owner can do |
| --- | --- |
| Main chat and Side chats | Open a thread, compose with attachments, stream Markdown, stop a reply, and inspect queued messages. Send switches to Stop during a run and Stop remains usable with an empty draft. |
| Tasks | Show the agent's connection state, Scheduled tasks, and Today's activity. Confirm a reminder in chat only after the scheduler returns its saved ID and due time. The panel entry opens that same task. |
| Mail and Calendar | Search and open a full thread, preview an attachment, edit a draft, select a calendar and date range, and prepare changes for review. Use platform date/time controls and retain all-day dates. |
| Files and Documents | Browse and search live VM directories. Register an ordinary PDF for preview or reuse the `fileId` of a managed document. Fetch a signed URL for that version and open the shared dialog from `packages/components`. Closing it returns focus to the file entry. Documents adds upload, paging, zoom, form fields, platform sharing, and review of the prepared reply. [PDF registration](../bootstrap/design.md#register-a-pdf-from-files) |
| Computer | Switch between browser, terminal, files, and an available desktop. Renew the browser console, reopen a profile, edit a file, inspect command receipts, and open the hosted desktop stream in the Computer screen. |
| Activity and reviews | Inspect progress, evidence, input requests, pause/resume/cancel controls, saved results, and receipts. Display the action payload and account before offering approval or denial. |
| Ideas, Goals, Tracking, and inbox | Review suggestions, open milestones and linked tasks, inspect watch status and observations, check now or change watch state, and follow a notification to its source. |
| Apps and connections | Search capabilities, inspect grants and account identity, connect or disconnect, and see provider availability. |

Render mail, search, browser, desktop, and sourced choice cards with their evidence and stable IDs. Open plans, comparisons, finance results, documents, and receipts directly from history. Fetch expiring links when needed. Show loading, empty, provider-disabled, disconnected, expired-session, conflict, and retry states where they arise. After an ambiguous external write, reopen its receipt before offering any further action.

### Browser acceptance

1. Open the sample workspace in Main chat. Ask for a reminder 30 minutes before the seeded workshop. Compare the chat confirmation with the ID and time in Scheduled tasks. Type a new draft, open Files alongside Tasks, preview the form, and close it. Verify focus returns to the file and the draft and task entry remain visible. Close the tab, let the reminder run, then reopen the workspace and read its outcome under the same task ID.
2. On a narrow viewport or device, type a draft, show the keyboard, switch threads and return, stop a streaming reply, and inspect pending messages. Verify the draft, tap targets, focus, and scroll-to-latest behavior. Reopen a historical tool card after its original link expires. Create a nested file directly in the workspace while the tab is backgrounded. Refresh Files and find it by name without uploading it. Verify an unavailable VM displays a retry state.
3. Disable a provider and check its disabled view. Make a configured provider unreachable and check its unavailable view. Expire the Cueia session while a private route is open and verify the next private read returns the client to sign-in. With a valid session, return to a backgrounded tab after a task completes and verify the refreshed result.
4. Once mail and document review are available, run the workshop mail scenario in a desktop browser and in browsers on iOS and Android, each with fresh sample state. Open the attachment, delegate the fill, supply any missing answers, and close the tab. Let the worker prepare the review. Open the thread URL in a new tab, preview the filled fields, review the exact recipient and attachment, approve, and reopen the receipt from chat. Verify one outgoing message with the reviewed `fileId`. Record the browser, device or viewport, and observed results for each run.

## Capabilities

### New Capabilities

- `workspace-clients` covers browser navigation, responsive screens, tool cards, and document handling.

### Modified Capabilities

## Impact

The first working client connects bootstrap's shell to the runtime and durable task APIs. Saved threads, connectors, document review, browser, computer, and personal workspace screens follow their owning proposals. Use the stable identities returned by those APIs for navigation.
