CREATE TABLE IF NOT EXISTS app_group_channels (
    group_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (group_id, channel_id),
    FOREIGN KEY (group_id) REFERENCES app_groups (id),
    FOREIGN KEY (channel_id) REFERENCES app_channels (id)
);
