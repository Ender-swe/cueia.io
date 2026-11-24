CREATE TABLE conversations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES workspaces(id),
    kind text NOT NULL CHECK (kind IN ('main', 'side')),
    title text NOT NULL,
    source_kind text CHECK (source_kind IN ('mail_thread', 'calendar_event')),
    source_id uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, id),
    CHECK ((source_kind IS NULL) = (source_id IS NULL))
);
CREATE UNIQUE INDEX ON conversations (workspace_id) WHERE kind = 'main';

CREATE TABLE tasks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES workspaces(id),
    conversation_id uuid NOT NULL,
    task_type text NOT NULL,
    request jsonb NOT NULL,
    initial_plan jsonb NOT NULL,
    current_run integer NOT NULL CHECK (current_run > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, id),
    FOREIGN KEY (workspace_id, conversation_id) REFERENCES conversations(workspace_id, id)
);
CREATE INDEX ON tasks (workspace_id, conversation_id, created_at, id);

CREATE TABLE task_runs (
    workspace_id uuid NOT NULL REFERENCES workspaces(id),
    task_id uuid NOT NULL,
    run_number integer NOT NULL CHECK (run_number > 0),
    state text NOT NULL CHECK (state IN ('queued', 'scheduled', 'running', 'waiting_input', 'paused',
                                         'succeeded', 'failed', 'cancelled', 'expired')),
    version bigint NOT NULL CHECK (version >= 0),
    resume_state text,
    handler_version integer NOT NULL CHECK (handler_version > 0),
    checkpoint_version integer NOT NULL CHECK (checkpoint_version > 0),
    checkpoint jsonb NOT NULL,
    progress jsonb NOT NULL,
    due_at timestamptz,
    time_zone text,
    expires_at timestamptz,
    failed_attempts integer NOT NULL DEFAULT 0 CHECK (failed_attempts BETWEEN 0 AND 3),
    retry_not_before timestamptz,
    lease_token uuid,
    lease_expires_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    finished_at timestamptz,
    PRIMARY KEY (workspace_id, task_id, run_number),
    FOREIGN KEY (workspace_id, task_id) REFERENCES tasks(workspace_id, id),
    CHECK (resume_state IS NULL OR state = 'paused'),
    CHECK ((lease_token IS NOT NULL) = (state = 'running')),
    CHECK ((lease_expires_at IS NOT NULL) = (state = 'running')),
    CHECK (state <> 'scheduled' OR (due_at IS NOT NULL AND time_zone IS NOT NULL))
);
ALTER TABLE tasks ADD FOREIGN KEY (workspace_id, id, current_run)
    REFERENCES task_runs(workspace_id, task_id, run_number) DEFERRABLE INITIALLY DEFERRED;
CREATE INDEX ON task_runs (state, due_at, retry_not_before) WHERE state IN ('queued', 'scheduled');
CREATE INDEX ON task_runs (lease_expires_at) WHERE state = 'running';
CREATE INDEX ON task_runs (expires_at) WHERE state IN ('queued', 'scheduled', 'running', 'waiting_input', 'paused');

CREATE TABLE task_attempts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES workspaces(id),
    task_id uuid NOT NULL,
    run_number integer NOT NULL,
    lease_token uuid UNIQUE NOT NULL,
    started_at timestamptz NOT NULL DEFAULT now(),
    finished_at timestamptz,
    finish_reason text,
    checkpoint_version integer NOT NULL,
    checkpoint jsonb NOT NULL,
    progress jsonb NOT NULL,
    error jsonb,
    FOREIGN KEY (workspace_id, task_id, run_number) REFERENCES task_runs(workspace_id, task_id, run_number)
);
CREATE UNIQUE INDEX ON task_attempts (workspace_id, task_id, run_number) WHERE finished_at IS NULL;
CREATE INDEX ON task_attempts (workspace_id, task_id, run_number, started_at, id);

CREATE TABLE task_inputs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES workspaces(id),
    task_id uuid NOT NULL,
    run_number integer NOT NULL,
    state text NOT NULL CHECK (state IN ('pending', 'answered', 'closed')),
    prompt jsonb NOT NULL,
    answer_schema jsonb NOT NULL,
    answer jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    answered_at timestamptz,
    FOREIGN KEY (workspace_id, task_id, run_number) REFERENCES task_runs(workspace_id, task_id, run_number),
    CHECK ((answer IS NOT NULL) = (state = 'answered')),
    CHECK ((answered_at IS NOT NULL) = (state = 'answered'))
);
CREATE UNIQUE INDEX ON task_inputs (workspace_id, task_id, run_number) WHERE state = 'pending';

CREATE TABLE task_outcomes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES workspaces(id),
    task_id uuid NOT NULL,
    run_number integer NOT NULL,
    status text NOT NULL CHECK (status IN ('succeeded', 'failed', 'cancelled', 'expired')),
    message text NOT NULL,
    result jsonb NOT NULL,
    error_code text,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, task_id, run_number),
    FOREIGN KEY (workspace_id, task_id, run_number) REFERENCES task_runs(workspace_id, task_id, run_number)
);
CREATE INDEX ON task_outcomes (workspace_id, created_at, id);

CREATE TABLE mutation_responses (
    workspace_id uuid NOT NULL REFERENCES workspaces(id),
    operation_scope text NOT NULL,
    request_key text NOT NULL,
    request_fingerprint bytea NOT NULL,
    http_status integer NOT NULL,
    response jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (workspace_id, operation_scope, request_key)
);
