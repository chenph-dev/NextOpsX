-- 上次选中的资产，用于启动时恢复工作台。
ALTER TABLE ui_settings ADD COLUMN last_asset_id TEXT NOT NULL DEFAULT '';
