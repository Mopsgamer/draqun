ALTER TABLE app_group_messages ADD COLUMN channel_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE app_group_messages ADD COLUMN attachment_url TEXT NOT NULL DEFAULT '';
ALTER TABLE app_group_messages ADD COLUMN attachment_name TEXT NOT NULL DEFAULT '';
ALTER TABLE app_group_messages ADD COLUMN attachment_size INTEGER NOT NULL DEFAULT 0;
ALTER TABLE app_group_messages ADD COLUMN attachment_type TEXT NOT NULL DEFAULT '';
