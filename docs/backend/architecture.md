# Architecture

## Chosen style

One repository and one Go modular monolith. This is not microservices. The API and worker are separate runtime processes that share one codebase and one database model.

## Components

- Nuxt frontend: user interface; browser calls the Go API and holds WebSocket connections.
- Go API: REST endpoints, authentication, authorization, business rules, and WebSocket connection hub.
- PostgreSQL: durable source of truth.
- Redis: cache, Asynq queue, and Pub/Sub between API instances. It is never the durable source of truth.
- Go worker: asynchronous jobs, retries, reminders, imports, and outbox publishing.
- S3-compatible object storage: attachment bytes; MinIO locally, cloud object storage in production.

## Module boundary

Each backend module owns its handlers, services, validation, queries, and tests: auth, organizations, projects, tasks, comments, chat, notifications, realtime, and platform infrastructure.

## Request rule

`HTTP handler -> authentication -> tenant/permission check -> service -> PostgreSQL transaction -> response`.

The handler contains transport concerns only. The service contains business rules. SQL is generated through sqlc and accessed behind module-level query/repository code.

## Reliable real-time rule

For an important change, the database transaction updates the business data, inserts the activity log, and inserts an outbox event. The worker later publishes that event to Redis. Each Go API instance receives it and delivers it to its connected WebSocket clients.

This prevents a Redis outage from losing a task update after PostgreSQL has committed.

## Event envelope

Every WebSocket event includes `id`, `type`, `organization_id`, `project_id` when relevant, `occurred_at`, and a compact `data` payload. Clients must tolerate duplicated events and refetch on uncertainty.
