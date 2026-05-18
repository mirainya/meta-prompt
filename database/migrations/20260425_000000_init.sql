-- ============================================================
-- 初始化：核心表结构
-- ============================================================

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id         BIGSERIAL PRIMARY KEY,
    username   VARCHAR(50) NOT NULL UNIQUE,
    password   VARCHAR(200) NOT NULL,
    role       VARCHAR(20) NOT NULL DEFAULT 'user',
    credits    INT NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 元提示词模板表
CREATE TABLE IF NOT EXISTS templates (
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
CREATE TABLE IF NOT EXISTS histories (
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT REFERENCES users(id),
    input            TEXT NOT NULL,
    llm_provider     VARCHAR(50) NOT NULL,
    status           VARCHAR(20) NOT NULL DEFAULT 'pending',
    reasoner_output  JSONB,
    generator_output JSONB,
    template_ids     JSONB,
    duration_ms      INT NOT NULL DEFAULT 0,
    webhook_url      TEXT,
    webhook_secret   TEXT,
    source           VARCHAR(20) NOT NULL DEFAULT 'web',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- API Key 表
CREATE TABLE IF NOT EXISTS api_keys (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT REFERENCES users(id),
    key_hash      VARCHAR(64) NOT NULL UNIQUE,
    name          VARCHAR(100) NOT NULL,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    rate_limit    INT NOT NULL DEFAULT 60,
    credits_quota INT NOT NULL DEFAULT -1,
    last_used_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 模板版本历史表
CREATE TABLE IF NOT EXISTS template_versions (
    id          BIGSERIAL PRIMARY KEY,
    template_id BIGINT NOT NULL REFERENCES templates(id) ON DELETE CASCADE,
    prompt      TEXT NOT NULL,
    version     INT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 渠道源表
CREATE TABLE IF NOT EXISTS channel_sources (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(100) NOT NULL UNIQUE,
    base_url   VARCHAR(500) NOT NULL,
    api_key    VARCHAR(500) NOT NULL,
    proxy_url  VARCHAR(500),
    enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 渠道模型表
CREATE TABLE IF NOT EXISTS channel_models (
    id               BIGSERIAL PRIMARY KEY,
    source_id        BIGINT NOT NULL REFERENCES channel_sources(id) ON DELETE CASCADE,
    model_code       VARCHAR(100) NOT NULL UNIQUE,
    model_type       VARCHAR(20) NOT NULL DEFAULT 'chat',
    billing_type     VARCHAR(20) NOT NULL DEFAULT 'per_call',
    credits_per_call INT NOT NULL DEFAULT 1,
    token_price      INT NOT NULL DEFAULT 1,
    enabled          BOOLEAN NOT NULL DEFAULT TRUE,
    is_default       BOOLEAN NOT NULL DEFAULT FALSE,
    synced_at        TIMESTAMPTZ
);

-- ============================================================
-- 索引
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_templates_stage ON templates(stage);
CREATE INDEX IF NOT EXISTS idx_templates_is_default ON templates(is_default);
CREATE INDEX IF NOT EXISTS idx_histories_created_at ON histories(created_at);
CREATE INDEX IF NOT EXISTS idx_histories_user_id ON histories(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_template_versions_tid ON template_versions(template_id);
CREATE INDEX IF NOT EXISTS idx_channel_models_source ON channel_models(source_id);
