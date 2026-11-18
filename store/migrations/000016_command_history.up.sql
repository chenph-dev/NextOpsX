CREATE TABLE command_history (
    id TEXT PRIMARY KEY,
    host_id TEXT NOT NULL,
    session_id TEXT,
    source TEXT NOT NULL DEFAULT 'ai',
    command TEXT NOT NULL,
    exit_code INTEGER,
    output TEXT,
    ts INTEGER NOT NULL
);

CREATE INDEX idx_command_history_host_ts ON command_history (host_id, ts DESC);

INSERT INTO command_history (id, host_id, session_id, source, command, exit_code, output, ts)
SELECT c.id, v.host_id, c.session_id, 'ai', c.command, c.exit_code, c.output, c.ts
FROM commands c
JOIN conversations v ON v.id = c.session_id;
