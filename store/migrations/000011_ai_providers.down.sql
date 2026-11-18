-- 回滚多供应商：恢复单行 ai_config（不恢复数据），移除 conversations.provider_id。
ALTER TABLE conversations DROP COLUMN provider_id;

DROP TABLE IF EXISTS ai_providers;

CREATE TABLE IF NOT EXISTS ai_config (
    id                INTEGER PRIMARY KEY CHECK (id = 1),
    provider          TEXT NOT NULL,
    model             TEXT NOT NULL,
    base_url          TEXT NOT NULL,
    api_key_encrypted BLOB
);
