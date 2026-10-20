CREATE TABLE IF NOT EXISTS app_user_channels (
    user_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, channel_id),
    FOREIGN KEY (user_id) REFERENCES app_users (id),
    FOREIGN KEY (channel_id) REFERENCES app_channels (id)
);
