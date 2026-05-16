-- API Key 表扩展
ALTER TABLE api_keys ADD COLUMN user_id BIGINT REFERENCES users(id);
ALTER TABLE api_keys ADD COLUMN rate_limit INT NOT NULL DEFAULT 60;
ALTER TABLE api_keys ADD COLUMN credits_quota INT NOT NULL DEFAULT -1;
ALTER TABLE api_keys ADD COLUMN last_used_at TIMESTAMPTZ;

-- 模板版本历史表
CREATE TABLE template_versions (
    id          BIGSERIAL PRIMARY KEY,
    template_id BIGINT NOT NULL REFERENCES templates(id) ON DELETE CASCADE,
    prompt      TEXT NOT NULL,
    version     INT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_template_versions_tid ON template_versions(template_id);

-- History 表增加来源和 webhook 字段
ALTER TABLE histories ADD COLUMN webhook_url TEXT;
ALTER TABLE histories ADD COLUMN webhook_secret TEXT;
ALTER TABLE histories ADD COLUMN source VARCHAR(20) NOT NULL DEFAULT 'web';
