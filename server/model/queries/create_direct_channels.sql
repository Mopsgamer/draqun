CREATE TABLE IF NOT EXISTS app_direct_channels (
    user1_id INTEGER NOT NULL,
    user2_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    PRIMARY KEY (user1_id, user2_id),
    FOREIGN KEY (user1_id) REFERENCES app_users (id),
    FOREIGN KEY (user2_id) REFERENCES app_users (id),
    FOREIGN KEY (channel_id) REFERENCES app_channels (id)
);
