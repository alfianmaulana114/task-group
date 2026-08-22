# Data model

## Principles

- PostgreSQL is relational source of truth.
- Use UUID or UUIDv7 identifiers consistently.
- Every tenant-owned table carries `organization_id` where practical and is queried with tenant scope.
- Use foreign keys, unique constraints, indexes for common filters, and `created_at`/`updated_at` timestamps.

## Core entities

```text
users
organizations
organization_members
projects
project_members
tasks
task_dependencies
comments
activity_logs
refresh_sessions
chat_channels
chat_channel_members
chat_messages
attachments
notifications
outbox_events
```

## Key relationships

```text
user --< organization_members >-- organization --< projects
project --< project_members >-- user
project --< tasks --< comments
task --< task_dependencies >-- task
project -- 1 chat_channel --< chat_messages
user --< chat_messages
project/task --< attachments
```

## Important constraints

- `organization_members (organization_id, user_id)` is unique.
- `project_members (project_id, user_id)` is unique.
- `chat_channels.project_id` is unique for a project channel.
- A task dependency cannot reference itself and must not create a cycle.
- Refresh token values are stored only as hashes.
- Attachment storage keys are unique and never derived solely from user-provided file names.
