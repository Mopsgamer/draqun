CREATE TABLE IF NOT EXISTS app_server_config (
    id INTEGER PRIMARY KEY DEFAULT 1,
    max_message_size INTEGER NOT NULL DEFAULT 8000,
    user_token_expiration INTEGER NOT NULL DEFAULT 180,
    max_file_size INTEGER NOT NULL DEFAULT 10,
    notice_message TEXT NOT NULL DEFAULT '',
    lockdown INTEGER NOT NULL DEFAULT 0
);
