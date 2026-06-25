-- Video async gateway foundation.
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

CREATE TABLE IF NOT EXISTS video_call_templates (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    create_method   VARCHAR(10)  NOT NULL DEFAULT 'POST',
    create_path     VARCHAR(500) NOT NULL,
    query_method    VARCHAR(10)  NOT NULL DEFAULT 'GET',
    query_path      VARCHAR(500) NOT NULL,
    content_method  VARCHAR(10),
    content_path    VARCHAR(500),
    cancel_method   VARCHAR(10),
    cancel_path     VARCHAR(500),
    status_mapping  JSONB NOT NULL DEFAULT '{}',
    result_mapping  JSONB NOT NULL DEFAULT '{}',
    error_mapping   JSONB NOT NULL DEFAULT '{}',
    poll_config     JSONB NOT NULL DEFAULT '{}',
    timeout_config  JSONB NOT NULL DEFAULT '{}',
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS videocalltemplate_name ON video_call_templates (name);
CREATE INDEX IF NOT EXISTS videocalltemplate_status ON video_call_templates (status);

CREATE TABLE IF NOT EXISTS video_models (
    id                  BIGSERIAL PRIMARY KEY,
    public_model        VARCHAR(100) NOT NULL,
    display_name        VARCHAR(100),
    template_id         BIGINT NOT NULL REFERENCES video_call_templates(id) ON DELETE RESTRICT,
    upstream_model_id   VARCHAR(150),
    request_shape       VARCHAR(50) NOT NULL,
    status              VARCHAR(20) NOT NULL DEFAULT 'active',
    capabilities        JSONB NOT NULL DEFAULT '{}',
    defaults            JSONB NOT NULL DEFAULT '{}',
    limits              JSONB NOT NULL DEFAULT '{}',
    supported_options   JSONB NOT NULL DEFAULT '{}',
    extra_body_allow    JSONB NOT NULL DEFAULT '[]',
    sort_order          INTEGER NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS videomodel_public_model ON video_models (public_model);
CREATE INDEX IF NOT EXISTS videomodel_template_id ON video_models (template_id);
CREATE INDEX IF NOT EXISTS videomodel_status ON video_models (status);
CREATE INDEX IF NOT EXISTS videomodel_sort_order ON video_models (sort_order);

CREATE TABLE IF NOT EXISTS video_generation_tasks (
    id                         BIGSERIAL PRIMARY KEY,
    public_id                  VARCHAR(80) NOT NULL,
    user_id                    BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    api_key_id                 BIGINT NOT NULL REFERENCES api_keys(id) ON DELETE RESTRICT,
    group_id                   BIGINT REFERENCES groups(id) ON DELETE SET NULL,
    account_id                 BIGINT NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    channel_id                 BIGINT REFERENCES channels(id) ON DELETE SET NULL,
    video_model_id             BIGINT NOT NULL REFERENCES video_models(id) ON DELETE RESTRICT,
    requested_model            VARCHAR(150) NOT NULL,
    upstream_model             VARCHAR(150) NOT NULL,
    upstream_task_id           VARCHAR(200),
    status                     VARCHAR(20) NOT NULL DEFAULT 'queued',
    progress                   INTEGER NOT NULL DEFAULT 0,
    billing_state              VARCHAR(20) NOT NULL DEFAULT 'reserved',
    request_payload            JSONB NOT NULL DEFAULT '{}',
    upstream_request_payload   JSONB,
    upstream_response_payload  JSONB,
    result_payload             JSONB,
    error_code                 VARCHAR(100),
    error_message              TEXT,
    content_url                TEXT,
    upstream_content_url       TEXT,
    local_content_url          TEXT,
    billing_mode               VARCHAR(20) NOT NULL,
    unit_price                 NUMERIC(20,10) NOT NULL DEFAULT 0,
    unit_seconds               NUMERIC(20,10),
    requested_seconds          INTEGER,
    billable_seconds           INTEGER,
    reserved_cost              NUMERIC(20,10) NOT NULL DEFAULT 0,
    estimated_cost             NUMERIC(20,10) NOT NULL DEFAULT 0,
    actual_cost                NUMERIC(20,10) NOT NULL DEFAULT 0,
    idempotency_key            VARCHAR(128),
    submitted_at               TIMESTAMPTZ,
    started_at                 TIMESTAMPTZ,
    completed_at               TIMESTAMPTZ,
    expires_at                 TIMESTAMPTZ,
    next_poll_at               TIMESTAMPTZ,
    poll_count                 INTEGER NOT NULL DEFAULT 0,
    locked_until               TIMESTAMPTZ,
    created_at                 TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                 TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS videogenerationtask_public_id ON video_generation_tasks (public_id);
CREATE UNIQUE INDEX IF NOT EXISTS videogenerationtask_idempotency_key ON video_generation_tasks (idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX IF NOT EXISTS videogenerationtask_status_next_poll_at ON video_generation_tasks (status, next_poll_at);
CREATE INDEX IF NOT EXISTS videogenerationtask_user_id_created_at ON video_generation_tasks (user_id, created_at);
CREATE INDEX IF NOT EXISTS videogenerationtask_api_key_id_created_at ON video_generation_tasks (api_key_id, created_at);
CREATE INDEX IF NOT EXISTS videogenerationtask_billing_state ON video_generation_tasks (billing_state);
CREATE INDEX IF NOT EXISTS videogenerationtask_video_model_id ON video_generation_tasks (video_model_id);
CREATE INDEX IF NOT EXISTS videogenerationtask_upstream_task_id ON video_generation_tasks (upstream_task_id);

ALTER TABLE channel_model_pricing
    ADD COLUMN IF NOT EXISTS unit_seconds NUMERIC(20,10);

COMMENT ON COLUMN channel_model_pricing.unit_seconds IS 'Video segment billing unit seconds when billing_mode=segment.';

ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS video_task_id VARCHAR(80),
    ADD COLUMN IF NOT EXISTS video_seconds INTEGER,
    ADD COLUMN IF NOT EXISTS video_size VARCHAR(32),
    ADD COLUMN IF NOT EXISTS video_billing_units INTEGER;

CREATE INDEX IF NOT EXISTS idx_usage_logs_video_task_id ON usage_logs (video_task_id);
