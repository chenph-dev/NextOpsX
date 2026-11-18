-- 多 AI 供应商：用 ai_providers 表替换单行 ai_config；
-- 对话记录其使用的供应商 ID（空串 = 使用默认供应商）。
CREATE TABLE IF NOT EXISTS ai_providers (
    id                TEXT PRIMARY KEY,
    name              TEXT NOT NULL UNIQUE,
    protocol          TEXT NOT NULL,              -- openai | claude
    model             TEXT NOT NULL,
    base_url          TEXT NOT NULL DEFAULT '',
    api_key_encrypted BLOB,
    is_default        INTEGER NOT NULL DEFAULT 0,
    created_at        INTEGER NOT NULL,
    updated_at        INTEGER NOT NULL
);

ALTER TABLE conversations ADD COLUMN provider_id TEXT NOT NULL DEFAULT '';

DROP TABLE IF EXISTS ai_config;
