# Frontend architecture

## Role of the frontend

The Nuxt application presents the product and consumes the Go API. It improves usability but never decides security. Backend responses are canonical.

## Initial responsibilities

- Authentication-aware routes and session-aware UI.
- Organization and project selection.
- Project board, task detail, comments, and project group chat.
- Optimistic UI only where rollback/refetch behaviour is defined.
- REST data loading, WebSocket subscription, and accessible loading/error/empty states.

## Data and realtime behaviour

1. Fetch the initial canonical page/data through REST.
2. Subscribe to authorized organization/project events after session is established.
3. Apply a supported event only once using its event ID.
4. Refetch an affected resource/list if an event payload is incomplete, stale, or conflicts with current client state.
5. On reconnect, refetch visible project/task/chat data; never assume missed events are recoverable solely from the socket.

## API boundary

- Generate or validate TypeScript API types from `api/openapi.yaml` later; do not manually duplicate Go structs as long-term contracts.
- Keep API client, session handling, and WebSocket handling centralized rather than scattered through page components.
- A `401` leads to session renewal/login flow; a `403` shows denied access; a `404` does not leak cross-tenant existence.

## UX requirements

- Main workflow is desktop-first; mobile supports viewing and simple updates.
- Keyboard navigation, meaningful labels, focus states, and color-independent status cues are required.
- A pending mutation clearly indicates saving; failure restores/refetches state and explains the next action.
