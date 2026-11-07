CREATE TABLE IF NOT EXISTS app_collab_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    subchannel_name VARCHAR(64) NOT NULL DEFAULT 'main',
    input_muted BOOLEAN NOT NULL DEFAULT 0,
    output_muted BOOLEAN NOT NULL DEFAULT 0,
    video_enabled BOOLEAN NOT NULL DEFAULT 0,
    bitrate_kbps INTEGER NOT NULL DEFAULT 64,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(user_id, channel_id)
);
