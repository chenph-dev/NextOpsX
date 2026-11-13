-- 界面偏好：单行配置（id=1）。theme: system | light | dark。
CREATE TABLE IF NOT EXISTS ui_settings (
    id    INTEGER PRIMARY KEY CHECK (id = 1),
    theme TEXT NOT NULL DEFAULT 'system'
);
