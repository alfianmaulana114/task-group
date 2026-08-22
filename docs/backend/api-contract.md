# API contract

## Convention

- REST base path: `/api/v1`.
- JSON request/response bodies.
- Version the human- and machine-readable contract in `../api/openapi.yaml`.
- Use cursor pagination for messages and optionally task feeds; expose stable error codes.

## Initial endpoint groups

```text
POST /auth/register
POST /auth/login
POST /auth/refresh
POST /auth/logout

GET,POST          /organizations
GET,PATCH         /organizations/:organization_id
POST              /organizations/:organization_id/members

GET,POST          /projects
GET,PATCH,DELETE  /projects/:project_id
POST              /projects/:project_id/members

GET,POST          /projects/:project_id/tasks
GET,PATCH,DELETE  /tasks/:task_id
POST              /tasks/:task_id/comments

GET               /projects/:project_id/chat
GET,POST          /chat/channels/:channel_id/messages
PATCH,DELETE      /chat/messages/:message_id
POST              /chat/channels/:channel_id/read
```

## Contract rules

- IDs in a request never bypass authorization or tenant checks.
- A successful mutation returns the canonical updated resource.
- Validation errors use `400` or `422`; unauthenticated is `401`; forbidden is `403`; absent/not visible is `404`; conflicts are `409`.
- List endpoints expose pagination metadata and only accept allowlisted filters/sort fields.
