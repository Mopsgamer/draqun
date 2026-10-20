CREATE TABLE IF NOT EXISTS app_user_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    actor_id INTEGER NOT NULL,
    action VARCHAR(64) NOT NULL,
    details VARCHAR(510) NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES app_users (id)
);
