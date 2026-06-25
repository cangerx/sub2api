-- Multi-group routing: bind one API key to multiple groups with priority/weight.
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS multi_group_routing BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS api_key_group_bindings (
    id          BIGSERIAL PRIMARY KEY,
    api_key_id  BIGINT NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    group_id    BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    priority    INTEGER NOT NULL DEFAULT 0,
    weight      INTEGER NOT NULL DEFAULT 100,
    enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS apikeygroupbinding_api_key_id_group_id
    ON api_key_group_bindings (api_key_id, group_id);
CREATE INDEX IF NOT EXISTS apikeygroupbinding_api_key_id
    ON api_key_group_bindings (api_key_id);
CREATE INDEX IF NOT EXISTS apikeygroupbinding_group_id
    ON api_key_group_bindings (group_id);
