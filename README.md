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

Turbo builds the TypeScript packages in dependency order. The root tasks also run the Go build, vet, and test commands from `apps/server/`.

## Start the local app

On a new checkout, install Bun 1.3.10 and Docker Compose, then run these commands from the repository root.

```sh
bun install --frozen-lockfile
cp .env.example .env
mkdir -p -m 700 "$HOME/.local/share/cueia/sample-disks"
```

Set `VITE_CLERK_PUBLISHABLE_KEY` and `CLERK_ISSUER` in `.env` for your Clerk test instance. Set `CUEIA_TEST_UID` and `CUEIA_TEST_GID` to the output of `id -u` and `id -g` if they differ from the defaults. If you already have a sample database, use its existing `POSTGRES_PASSWORD`.

```sh
bun start
```

Open `http://127.0.0.1:5173/workspace`. Compose runs Vite alongside the API, worker, and PostgreSQL. Edits in `apps/dashboard` and the shared UI packages update the open dashboard automatically. Use `docker compose logs -f dashboard` to see Vite’s output. Run `bun run stop` to stop the whole app. The database volume and sample files persist.

## Starting a deployment

In a separate deployment env file, set `CUEIA_MODE` to `sample` or `live`. Set `POSTGRES_PASSWORD` to a URL-safe secret and `WORKSPACE_NETWORK` to an existing Docker network with private connectivity to that deployment's workspace VMs. Configure `CLERK_ISSUER`, `CLERK_AUTHORIZED_PARTY`, and `CPK_INTELLIGENCE_API_KEY` for the API.

```sh
docker compose --env-file /path/to/deployment.env -f compose.yaml up -d --build --force-recreate
```

Compose waits for healthy PostgreSQL and a successful migration before starting the API and worker. Each mode has its own Compose project and database volume. Set distinct `CUEIA_PORT` values when running both modes at once.

The production workspace provider target is Azure. The internal provider fixture stores each test workspace in its own temporary disk directory. Its allocation and replacement test covers a lost allocation response, a fresh provider instance recovering the binding, rejection of foreign machine and disk references, and replacement retaining the original bytes. Run it from `apps/server/` with `go test -v ./internal/workspace/test -run '^TestAllocationAndReplacement$'`.

### Azure workspace provider

Use a separate resource group, deployment identity and private subnet per mode. The server uses Azure DefaultAzureCredential with permission to deploy a VM, NIC and disk in that group and join the selected subnet. The guest VM has no Azure management identity. Assign its subnet rules to allow private SSH connections from the API and worker. Configure the server with `AZURE_SUBSCRIPTION_ID`, `AZURE_RESOURCE_GROUP`, `AZURE_LOCATION`, `AZURE_SUBNET_ID`, `AZURE_IMAGE_ID`, `AZURE_VM_SIZE`, `AZURE_VM_ADMIN_PUBLIC_KEY` and `CUEIA_DEPLOYMENT_ID`.

The image must mount the retained LUN 0 data disk on boot, initialize only a new empty disk, and use its workspace directory as the working directory for the owner and agent. The private file account is `cueiafiles`. Restrict it to the workspace filesystem and authorize the public half of the deployment's file-client key. Configure the guest to obtain an SSH host certificate for the lower-case principal `<vm-name>-<vmId>` from the deployment's trusted host CA. Keep the CA signing key outside the guest image. Supply the server-side key and pinned public CA through `CUEIA_FILE_CLIENT_KEY_PATH` and `CUEIA_HOST_CA_PATH`.

The adapter checks the VM owner tag, attached disk and NIC before opening a private connection. Replacements retain the data disk. Local isolation checks use the internal fixture. The live two-VM isolation run is deferred.

Build the guest command from `apps/server/` with `CGO_ENABLED=0 go build -o cueia-files ./cmd/fileservice` for the image architecture. Install it as a root-owned executable on the image. Configure sshd to force `/usr/local/bin/cueia-files` for `cueiafiles` and disable forwarding, TTY and password login for that account. The service expects the mounted disk at `/srv/cueia`. Grant the service account access to the disk, share its `workspace/` directory with the owner and agent, and keep the sibling `private/` directory accessible only to the service account. Gate SSH service availability on the mount.

### Dashboard sign-in

Signed-out visitors can choose “Open sample workspace” at `/` to sign in or create an account and continue directly to their private workspace at `/workspace`. Opening `/workspace` directly also preserves the destination through sign-in and sign-up. Direct visits to either form lead to the sample dashboard after authentication. Configure static hosting to serve the dashboard entry point for both auth routes and their nested paths so verification, password recovery, and OAuth callbacks work on refresh.

Supply the publishable key for the deployment’s Clerk instance when building the dashboard. Configure the API's `CLERK_ISSUER` for the same instance and `CLERK_AUTHORIZED_PARTY` for the dashboard origin. For local sample development, use `http://127.0.0.1:5173` as the dashboard origin. Set signup to public, enable email addresses as a sign-in identifier, and enable password authentication. Enable Google and GitHub as social sign-in options. Use the live deployment's own instance settings and dashboard origin when deploying live.

### Interactive dashboard acceptance

Build the dashboard with the Clerk test publishable key configured. Install Chromium from `apps/browser-worker` with `bunx playwright install chromium`. From `apps/server/`, run `go test ./cmd/api/routes/test -run '^TestBrowserJourney$' -count=1 -v` with `CUEIA_TEST_DATABASE_URL` pointing at disposable PostgreSQL and `CUEIA_BROWSER_CLERK_ISSUER` set to the matching Clerk test issuer.

The test starts an isolated sample workspace, serves the built dashboard on loopback, and opens a fresh Chromium window for sign-in. It writes a nested CSV directly to the internal provider's test disk and checks the Files and Tasks panels, signed PDF preview, keyboard focus, and draft preservation.

For unattended dashboard tests, set `CLERK_SECRET_KEY` and `CUEIA_TEST_USER_EMAIL` in the repository root `.env`. Use the development instance's secret key and the email of an existing test user without MFA or pending session tasks. The Bun browser worker loads this ignored file, signs in with Clerk's Playwright helper and runs headlessly. Leave the email empty for the interactive sign-in journey. The Google and GitHub acceptance cases still exercise their respective OAuth flows.

For the fresh-install sample checks, add `compose.test-workspace.yaml` to `COMPOSE_FILE`. Supply an empty private directory in `CUEIA_TEST_DISK_DIR` and its numeric owner and group in `CUEIA_TEST_UID` and `CUEIA_TEST_GID`. Each internal test machine retains its own disk under that directory. The live service uses Azure.

Verify Google sign-in in a regular browser. Run `bun --env-file=../../.env run first-visit.ts` from `apps/browser-worker` against an empty sample deployment using the same Clerk test identity. Set `CLERK_AUTHORIZED_PARTY` to the loopback dashboard origin, `CUEIA_BROWSER_CLERK_ISSUER` to the instance issuer, `COMPOSE_PROJECT_NAME` and `CUEIA_TEST_DISK_DIR` to the deployment values, and `CUEIA_ACCEPTANCE_DIR` to a private temporary directory. The runner keeps the sessions there for the subsequent recovery checks.

For the second owner, sign in through GitHub in a private browser window against the same sample. Set `CUEIA_TEST_BOB_EMAIL` in the root `.env` to that distinct Clerk user's email, then run `bun --env-file=../../.env run second-visit.ts` with the same acceptance directory and deployment settings. The test compares both owners' snapshots, reads the same relative path on each disk, and verifies that Bob cannot read or obtain a preview link for Alice's attachment.

## License

Cueia's original source code is licensed under the GNU Affero General Public License, version 3 only (`AGPL-3.0-only`). See the [full license](LICENSE). Third-party dependencies and assets retain their respective licenses. See [provenance](PROVENANCE.md).
