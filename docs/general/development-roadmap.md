# Development roadmap

## Phase 0: documentation and local foundation

Finalize product requirements, business rules, ERD, API contract, Docker Compose, environment template, and README. Acceptance: a newcomer can understand the intended system before code is added.

## Phase 1: platform foundation

Set up Go API, Nuxt frontend, PostgreSQL migration workflow, health checks, configuration, logging, and basic CI. Acceptance: one command starts the local stack and API reaches PostgreSQL.

## Phase 2: identity and tenant model

Implement registration, login, sessions, organization membership, roles, and authorization middleware. Acceptance: cross-organization access is blocked by tests.

## Phase 3: project and task vertical slice

Implement project membership, tasks, board, assignment, status updates, activity logs, API documentation, and critical tests. Acceptance: a manager can create a project and team members can manage permitted work.

## Phase 4: collaboration

Add comments, WebSocket events, outbox worker, and one project group chat per project. Acceptance: a change from one browser appears on another without refresh and survives temporary Redis failure.

## Phase 5: supporting workflows

Add attachments, notifications, mentions, sprints, dependencies, search, and dashboard metrics only where requirements are confirmed.

## Phase 6: production hardening

Add E2E tests, rate limits, operational dashboards, backups, deployment pipeline, and runbook. Acceptance: a clean deployment and rollback procedure is documented and tested.
