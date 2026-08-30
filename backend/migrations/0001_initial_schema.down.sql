BEGIN;

DROP TRIGGER IF EXISTS chat_messages_set_updated_at ON chat_messages;
DROP TRIGGER IF EXISTS chat_channel_members_set_updated_at ON chat_channel_members;
DROP TRIGGER IF EXISTS chat_channels_set_updated_at ON chat_channels;
DROP TRIGGER IF EXISTS comments_set_updated_at ON comments;
DROP TRIGGER IF EXISTS tasks_set_updated_at ON tasks;
DROP TRIGGER IF EXISTS project_members_set_updated_at ON project_members;
DROP TRIGGER IF EXISTS projects_set_updated_at ON projects;
DROP TRIGGER IF EXISTS organization_members_set_updated_at ON organization_members;
DROP TRIGGER IF EXISTS organizations_set_updated_at ON organizations;
DROP TRIGGER IF EXISTS users_set_updated_at ON users;

DROP TABLE IF EXISTS outbox_events;
DROP TABLE IF EXISTS refresh_sessions;
DROP TABLE IF EXISTS activity_logs;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS attachments;
DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS chat_channel_members;
DROP TABLE IF EXISTS chat_channels;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS task_dependencies;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS users;

DROP FUNCTION IF EXISTS set_updated_at;

COMMIT;
