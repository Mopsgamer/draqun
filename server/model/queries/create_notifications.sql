CREATE TABLE IF NOT EXISTS app_notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    actor_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    message_id INTEGER NOT NULL,
    type VARCHAR(32) NOT NULL DEFAULT 'mention',
    title VARCHAR(255) NOT NULL,
    content VARCHAR(510) NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES app_users (id),
    FOREIGN KEY (actor_id) REFERENCES app_users (id),
    FOREIGN KEY (group_id) REFERENCES app_groups (id),
    FOREIGN KEY (channel_id) REFERENCES app_channels (id),
    FOREIGN KEY (message_id) REFERENCES app_group_messages (id)
);
