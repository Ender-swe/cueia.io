# Cueia computer backends and workspace files

## Why

The owner can inspect commands run on their behalf, edit saved files, or open a graphical desktop when configured. Files and receipts must remain attributable to the correct workspace and provider.

## What Changes

### Computer access to the existing workspace

Report the selected provider, availability, running state, network access, and display capability. Support start, status, stop, run a bounded command, inspect receipts, list/create folders, read/write text, and transfer PDFs to or from Documents. Resolve ownership and deployment on the server. The API, task worker, terminal and desktop resolve the same private VM from the [bootstrap machine binding](../bootstrap/design.md#bootstrap-database-tables). Run user and agent commands in that workspace. Their files appear in Files on refresh. When the provider is disabled or unreachable, expose the reason and keep command execution off the API host.

Package the Linux command environment as a dedicated root Dockerfile target. In a cloud deployment, run it inside the owner's existing workspace VM with its persistent workspace filesystem at `/workspace`. Use a disposable, per-owner Docker fixture for local contract tests. Run commands as a nonroot user with networking disabled for this command environment, read-only system files, bounded temporary storage, and dropped capabilities. Exclude host directories, the Docker socket, and application credentials. Restrict file operations to the workspace root, including symlink resolution. Keep a 30-second command deadline and 128 KB output cap, and return exit status, timeout or truncation, and saved output.

Persist a receipt containing owner, provider, computer ID, command, start/end time, and outcome. Save running state before dispatch. On stop or restart, reconcile abandoned commands as interrupted. A stop preserves files and ends processes. Keep a file editor's unsaved changes when a slower load or save returns. Text edits write the ordinary workspace path. Importing a PDF from that path into Documents captures its current bytes. Exporting a selected document reads the immutable version identified by its `fileId`. [Bootstrap file and document operations](../bootstrap/proposal.md#bootstrap-api-routes)

### Optional hosted desktop

For E2B, require server-side credentials and a stable deployment ID. Enable hosted desktop only when the configured provider can expose the owner's bound workspace machine and persistent filesystem. Verify the machine, workspace, and deployment before issuing desktop access. Display that the desktop has network access. Preserve workspace files across pause/resume. Start the display before issuing access and embed its live stream in the web client. Allow pointer, keyboard, scroll, and screenshot actions only against a running desktop. Use screenshot pixel coordinates and return a fresh image after each action. Stop closes stream access and ends running programs.

### Acceptance

1. Run a disposable per-owner Docker fixture. Create a folder and text file, run a command that reads it, and inspect the receipt. Refresh Files in a second session and read the same text without registering a document. Stop and restart the computer and reopen the file. Attempt to read outside the workspace and expect denial.
2. Run a command past the deadline and another beyond the output cap. Check timeout and truncation independently. Interrupt a running command, restart the service, and verify the original receipt shows interruption. Open file A then B and delay A’s response until after editing B. B must retain its text.
3. With a hosted transport double, exercise paused, running, and stopped states and rejection of a sandbox belonging to another deployment. For a credentialed run, create a file in the owner's desktop and read it through the workspace API. Open a fixture document and compare screenshots before and after. Verify the stream opens in the web client and stopping the desktop closes access. Record the actual provider and browser exercised.
4. Import a fixture PDF from the terminal's workspace into Documents. Edit the PDF at the same path and import it again. After restarting the computer with its disk retained, export each `fileId` to a separate ordinary file. Compare each export's byte count and digest to the corresponding source version.

## Capabilities

### New Capabilities

- `workspace-computer` covers computer lifecycle, commands, workspace files, PDF transfer, and receipts.
- `hosted-desktop` covers the optional graphical session, visual actions, and stream access.

### Modified Capabilities

## Impact

Depends on workspace access and document storage. Runtime tools and the Computer screen share the lifecycle and file API. Browser sessions have their own storage, network policy, and controls.
