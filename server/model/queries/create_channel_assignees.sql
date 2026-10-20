CREATE TABLE IF NOT EXISTS app_channel_assignees (
    channel_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    PRIMARY KEY (channel_id, user_id),
    FOREIGN KEY (channel_id) REFERENCES app_channels (id),
    FOREIGN KEY (user_id) REFERENCES app_users (id)
);
