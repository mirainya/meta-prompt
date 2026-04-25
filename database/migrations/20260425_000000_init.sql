-- 元提示词模板表
CREATE TABLE templates (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    stage       VARCHAR(20) NOT NULL,
    prompt      TEXT NOT NULL,
    version     INT NOT NULL DEFAULT 1,
    is_default  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 生成历史表
CREATE TABLE histories (
    id               BIGSERIAL PRIMARY KEY,
    input            TEXT NOT NULL,
    llm_provider     VARCHAR(50) NOT NULL,
    reasoner_output  JSONB NOT NULL,
    generator_output JSONB NOT NULL,
    template_ids     JSONB NOT NULL,
    duration_ms      INT NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- API Key表
CREATE TABLE api_keys (
    id         BIGSERIAL PRIMARY KEY,
    key_hash   VARCHAR(64) NOT NULL UNIQUE,
    name       VARCHAR(100) NOT NULL,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_templates_stage ON templates(stage);
CREATE INDEX idx_templates_is_default ON templates(is_default);
CREATE INDEX idx_histories_created_at ON histories(created_at);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
