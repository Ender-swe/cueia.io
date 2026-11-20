# Cueia computer backends and workspace files

## Why

The owner can inspect commands run on their behalf, edit saved files, or open a graphical desktop when configured. Files and receipts must remain attributable to the correct workspace and provider.

## What Changes

### Shared API and local computer

Report the selected provider, availability, running state, network access, and display capability. Support start, status, stop, run a bounded command, inspect receipts, list/create folders, read/write text, and transfer PDFs to or from Documents. Resolve ownership and deployment on the server. The API and task worker address the same computer. When the provider is disabled or unreachable, expose the reason and keep command execution off the API host.

Build the local Linux computer as a dedicated target of the root multi-stage Dockerfile. Run commands as a nonroot user in `/workspace` with networking disabled, read-only system files, bounded temporary storage, dropped capabilities, and a persistent workspace volume. Exclude host directories, the Docker socket, and application credentials from the container. Restrict file operations to the workspace, including traversal through symlinks. Support noninteractive commands with a 30-second deadline and 128 KB output cap. Return exit status, timeout or truncation, and saved output.

Persist a receipt containing owner, provider, computer ID, command, start/end time, and outcome. Save running state before dispatch. On stop or restart, reconcile abandoned commands as interrupted. A stop preserves files and ends processes. Keep a file editor’s unsaved changes when a slower load or save returns. Import/export PDFs as bytes through the authenticated document service.

### Optional hosted desktop

For E2B, require server-side credentials and a stable deployment ID. Associate sandboxes with both deployment and owner before reuse. Display that the desktop has network access. Preserve workspace files across pause/resume. Start the display before issuing access, embed its live stream on web, and hand off to the system browser on native. Allow pointer, keyboard, scroll, and screenshot actions only against a running desktop. Use screenshot pixel coordinates and return a fresh image after each action. Stop closes stream access and ends running programs.

### Acceptance

1. Run a real disposable Docker computer. Create a folder and text file, run a command that reads it, and inspect the receipt. Transfer a Cueia fixture PDF in both directions and compare byte counts and digests. Stop and restart, then reopen both files. Attempt to read outside the workspace and expect denial.
2. Run a command past the deadline and another beyond the output cap. Check timeout and truncation independently. Interrupt a running command, restart the service, and verify the original receipt shows interruption. Open file A then B and delay A’s response until after editing B. B must retain its text.
3. With a hosted transport double, exercise paused, running, and stopped states and rejection of a sandbox belonging to another deployment. For a credentialed run, open a fixture document and compare screenshots before and after. Verify web/native handoff and that stopping closes access. Record the actual provider and platform exercised.

## Capabilities

### New Capabilities

- `workspace-computer` covers computer lifecycle, commands, workspace files, PDF transfer, and receipts.
- `hosted-desktop` covers the optional graphical session, visual actions, and stream access.

### Modified Capabilities

## Impact

Depends on workspace access and document storage. Runtime tools and the Computer screen share the lifecycle and file API. Browser sessions have their own storage, network policy, and controls.
