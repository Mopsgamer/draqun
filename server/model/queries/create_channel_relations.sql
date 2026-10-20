CREATE TABLE IF NOT EXISTS app_channel_relations (
    parent_id INTEGER NOT NULL,
    child_id INTEGER NOT NULL,
    relation_type VARCHAR(32) NOT NULL DEFAULT 'subchannel',
    PRIMARY KEY (parent_id, child_id),
    FOREIGN KEY (parent_id) REFERENCES app_channels (id),
    FOREIGN KEY (child_id) REFERENCES app_channels (id)
);
