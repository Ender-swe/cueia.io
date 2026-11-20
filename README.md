# Cueia

Cueia is a planned personal workspace for web, iOS, and Android. The owner will be able to delegate work in a conversation, close the app, and return to the saved result. For example, the agent can prepare a filled PDF from a mail attachment, show the owner the proposed reply and attachment, and send them after approval.

## Project status

This repository contains the OpenSpec proposals for a ground-up rebuild. The application is in the planning stage. The intended deliverable is a self-contained TypeScript workspace in `cueia/` that a developer can build, run, and test on a clean machine.

## Where to start

Begin with the foundation. Its first milestone is to sign in to a seeded sample workspace, restart the API and recover the same records, then sign out and verify that the old session is rejected. [Bootstrap proposal](openspec/changes/bootstrap-cueia/proposal.md)

The next milestone connects durable tasks, the agent runtime, and the first client so a delegated job survives closing the app. The roadmap covers all 12 proposals through release and identifies which features can be built in parallel. [Proposal order](PROPOSAL-ORDER.md)

With the OpenSpec CLI installed, run these commands from the repository root to list active changes and check which planning artifacts the foundation still needs.

```sh
openspec list
openspec status --change bootstrap-cueia
```

## Development

Install Go 1.26.2 and Bun 1.3.10. Run these commands at the repository root.

```sh
bun install --frozen-lockfile
bun run build
bun run type-check
bun run test
```

Turbo builds the TypeScript packages in dependency order. The root tasks also run the Go build, vet, and test commands from `backend/`.

## Starting a deployment

Set `POSTGRES_PASSWORD` to a URL-safe secret and `WORKSPACE_NETWORK` to an existing Docker network with private connectivity to that deployment's workspace VMs. Configure `CLERK_ISSUER`, `CLERK_AUTHORIZED_PARTY`, and `CPK_INTELLIGENCE_API_KEY` for the API. Load the environment for the desired mode and run `scripts/start sample` or `scripts/start live`. Set `CUEIA_PORT` to distinct ports when running both simultaneously. Both publish the API on loopback.

Each startup stops the application processes before running a fresh migration job against healthy PostgreSQL. The script starts the API and worker after that job succeeds. Sample and live use separate Compose project names, database volumes, and provider allocation namespaces.

The production workspace provider target is Azure. The internal provider fixture stores each test workspace in its own temporary disk directory. Its tests cover a lost allocation response, a fresh provider instance recovering the binding, rejection of foreign machine and disk references, and replacement retaining the original bytes. Run them from `backend/` with `go test ./internal/workspace/provider/...`.

### Azure workspace provider

Use a separate resource group, deployment identity and private subnet per mode. The server uses Azure DefaultAzureCredential with permission to deploy a VM, NIC and disk in that group and join the selected subnet. The guest VM has no Azure management identity. Assign its subnet rules to allow private SSH connections from the API and worker. Configure the server with `AZURE_SUBSCRIPTION_ID`, `AZURE_RESOURCE_GROUP`, `AZURE_LOCATION`, `AZURE_SUBNET_ID`, `AZURE_IMAGE_ID`, `AZURE_VM_SIZE`, `AZURE_VM_ADMIN_PUBLIC_KEY` and `CUEIA_DEPLOYMENT_ID`.

The image must mount the retained LUN 0 data disk on boot, initialize only a new empty disk, and use its workspace directory as the working directory for the owner and agent. The private file account is `cueiafiles`. Restrict it to the workspace filesystem and authorize the public half of the deployment's file-client key. Configure the guest to obtain an SSH host certificate for the lower-case principal `<vm-name>-<vmId>` from the deployment's trusted host CA. Keep the CA signing key outside the guest image. Supply the server-side key and pinned public CA through `CUEIA_FILE_CLIENT_KEY_PATH` and `CUEIA_HOST_CA_PATH`.

The adapter checks the VM owner tag, attached disk and NIC before opening a private connection. Replacements retain the data disk. Local isolation checks use the internal fixture. The live two-VM isolation run is deferred.

Build the guest command from `backend/` with `CGO_ENABLED=0 go build -o cueia-files ./cmd/fileservice` for the image architecture. Install it as a root-owned executable on the image. Configure sshd to force `/usr/local/bin/cueia-files` for `cueiafiles` and disable forwarding, TTY and password login for that account. The service expects the mounted disk at `/srv/cueia`. Grant the service account access to the disk, share its `workspace/` directory with the owner and agent, and keep the sibling `private/` directory accessible only to the service account. Gate SSH service availability on the mount.

### Dashboard sign-in

Set `VITE_CLERK_PUBLISHABLE_KEY` in `apps/dashboard/.env.local` to the publishable key for the Clerk instance used by that deployment. Configure the API's `CLERK_ISSUER` for the same instance and `CLERK_AUTHORIZED_PARTY` for the dashboard origin. For local sample development, use `http://127.0.0.1:5173` as the dashboard origin. Set signup to public, enable email addresses as a sign-in identifier, and enable password authentication. Enable Google and GitHub as social sign-in options. Use the live deployment's own instance settings and dashboard origin when deploying live.

With the API running, start the dashboard with `cd apps/dashboard && bun run dev --port 5173 --strictPort` and open `http://127.0.0.1:5173`. Vite proxies API requests to loopback port 8080. Set `CUEIA_API_ORIGIN` to point the development or preview server at a different API. Keep API and provider credentials in the server environment. The dashboard build needs only the publishable key.

### Interactive dashboard acceptance

Build the dashboard with the Clerk test publishable key configured. Install Chromium from `apps/browser-worker` with `bunx playwright install chromium`. From `backend/`, run `go test ./cmd/api -run '^TestBrowserJourney$' -count=1 -v` with `CUEIA_TEST_DATABASE_URL` pointing at disposable PostgreSQL and `CUEIA_BROWSER_CLERK_ISSUER` set to the matching Clerk test issuer.

The test starts an isolated sample workspace, serves the built dashboard on loopback, and opens a fresh Chromium window for sign-in. It writes a nested CSV directly to the internal provider's test disk and checks the Files and Tasks panels, signed PDF preview, keyboard focus, and draft preservation.
