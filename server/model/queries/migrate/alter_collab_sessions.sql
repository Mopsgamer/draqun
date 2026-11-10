ALTER TABLE app_collab_sessions ADD COLUMN is_camera_enabled BOOLEAN NOT NULL DEFAULT 0;
ALTER TABLE app_collab_sessions ADD COLUMN is_screen_shared BOOLEAN NOT NULL DEFAULT 0;
ALTER TABLE app_collab_sessions ADD COLUMN is_file_shared BOOLEAN NOT NULL DEFAULT 0;
