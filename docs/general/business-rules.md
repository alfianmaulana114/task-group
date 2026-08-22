# Business rules

## Tenancy and access

- A user may belong to multiple organizations.
- Every project belongs to exactly one organization.
- A user may access a project only when they are an active member of that organization and project, unless an organization-level policy explicitly grants access.
- Every query and mutation must be scoped to the current organization; an ID from another organization must never reveal data.

## Roles

- `owner`: full organization control; ownership cannot be removed without transfer.
- `admin`: organization administration and member management.
- `project_manager`: manages assigned projects, memberships, tasks, and sprints.
- `member`: accesses assigned projects, works on tasks, comments, and chats.
- Backend permission checks are authoritative; hidden UI is not authorization.

## Tasks

- A task belongs to one project and has one creator.
- A task may have zero or one assignee initially.
- Only authorized project members may create or modify a task.
- A task can move through the initial fixed statuses: `backlog`, `todo`, `in_progress`, `in_review`, `done`.
- A task with unfinished blocking dependencies cannot be treated as ready; the UI must show why.
- Every meaningful change creates an activity entry.

## Comments and chat

- Only project members can read or write task comments and the project's group chat.
- Each project has one system-created project chat channel.
- A message is persisted before it is broadcast in real time.
- Members may edit/delete only their own content; managers/admins may moderate according to policy.
- Deletes are soft deletes when audit history is required.

## Files

- Database stores file metadata only; bytes are stored in object storage.
- The server validates file size, allowed MIME type, and authorization before creating an upload URL.

## Events and jobs

- A completed business transaction writes an outbox event in the same PostgreSQL transaction.
- A worker publishes the outbox event after commit and retries safely.
- All jobs and event consumers must be idempotent.
