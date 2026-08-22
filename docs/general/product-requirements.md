# Product requirements

## Product

Task Group is a multi-organization, real-time collaborative project-management SaaS for small teams. It combines projects, tasks, contextual discussion, project group chat, and delivery visibility.

## Target users

- Organization owner: owns workspace and its settings.
- Admin: manages members and organization administration.
- Project manager: manages a project's work and members.
- Member: completes work, comments, and chats in projects they joined.

## Product goals

- Give a team one reliable place to plan, assign, discuss, and track work.
- Show production-oriented full-stack engineering without unnecessary services.
- Make project changes visible to other active users in real time.

## MVP scope

1. Authentication and secure sessions.
2. Organizations and memberships.
3. Projects and project memberships.
4. Tasks with title, description, status, priority, assignee, creator, and due date.
5. Kanban board using fixed initial statuses: backlog, todo, in_progress, in_review, done.
6. Task comments and activity log.
7. Real-time task and comment updates.
8. Project group chat: one automatic channel per project.
9. Docker-based local environment and tests for critical flows.

## Later scope

Attachments, mentions, notifications, sprints, dependencies, calendar, search, analytics, import CSV, email reminders, monitoring, and cloud deployment.

## Explicit non-goals for early versions

Custom workflows, custom role builder, billing, SSO, mobile application, integrations, microservices, Kafka, Elasticsearch, Kubernetes, CQRS, and event sourcing.

## Primary user journey

Register -> create organization -> invite members -> create project -> add members -> create and assign tasks -> collaborate through comments/chat -> move tasks on the board -> observe live updates and activity history.
