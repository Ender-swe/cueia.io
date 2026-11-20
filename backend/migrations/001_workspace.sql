CREATE TABLE workspaces (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    clerk_user_id text UNIQUE NOT NULL,
    workspace_id uuid UNIQUE NOT NULL REFERENCES workspaces(id),
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE sessions (
    token_digest bytea PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz
);
CREATE TABLE workspace_machines (
    workspace_id uuid PRIMARY KEY REFERENCES workspaces(id),
    provider text NOT NULL,
    region text NOT NULL,
    vm_resource_id text,
    workspace_disk_resource_id text,
    provisioning_state text NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    last_error jsonb,
    power_state text,
    last_heartbeat_at timestamptz,
    UNIQUE (provider, vm_resource_id),
    UNIQUE (provider, workspace_disk_resource_id)
);
CREATE TABLE workspace_preferences (
    workspace_id uuid PRIMARY KEY REFERENCES workspaces(id),
    settings jsonb NOT NULL CHECK (jsonb_typeof(settings) = 'object'),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE mail_threads (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES workspaces(id),
    subject text NOT NULL,
    UNIQUE (workspace_id, id)
);
CREATE TABLE mail_messages (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES workspaces(id),
    thread_id uuid NOT NULL,
    sender_address text NOT NULL,
    body_text text NOT NULL,
    sent_at timestamptz NOT NULL,
    UNIQUE (workspace_id, id),
    FOREIGN KEY (workspace_id, thread_id) REFERENCES mail_threads(workspace_id, id)
);
CREATE TABLE files (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES workspaces(id),
    filename text NOT NULL,
    media_type text NOT NULL,
    byte_size bigint NOT NULL CHECK (byte_size >= 0),
    sha256_digest bytea NOT NULL CHECK (octet_length(sha256_digest) = 32),
    storage_key text NOT NULL,
    UNIQUE (workspace_id, id),
    UNIQUE (workspace_id, storage_key)
);
CREATE TABLE mail_attachments (
    workspace_id uuid NOT NULL REFERENCES workspaces(id),
    message_id uuid NOT NULL,
    file_id uuid NOT NULL,
    PRIMARY KEY (workspace_id, message_id, file_id),
    FOREIGN KEY (workspace_id, message_id) REFERENCES mail_messages(workspace_id, id),
    FOREIGN KEY (workspace_id, file_id) REFERENCES files(workspace_id, id)
);
CREATE TABLE calendar_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES workspaces(id),
    title text NOT NULL,
    starts_at timestamptz NOT NULL,
    ends_at timestamptz NOT NULL CHECK (ends_at > starts_at),
    time_zone text NOT NULL
);
CREATE TABLE claim_records (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES workspaces(id),
    status text NOT NULL CHECK (status IN ('eligible','claimed')),
    version bigint NOT NULL CHECK (version >= 0)
);
CREATE TABLE deployment_signing_keys (
    singleton boolean PRIMARY KEY CHECK (singleton),
    key_material bytea NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON mail_messages (workspace_id, thread_id, sent_at, id);
CREATE INDEX ON calendar_events (workspace_id, starts_at);
