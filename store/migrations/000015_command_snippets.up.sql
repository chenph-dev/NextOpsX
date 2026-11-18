CREATE TABLE command_snippets (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    group_name TEXT NOT NULL DEFAULT '',
    command TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX idx_command_snippets_group_name ON command_snippets (group_name, name);
