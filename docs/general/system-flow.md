# End-to-end system flows

## Create and update a task

```text
Browser -> Go API -> authenticate -> tenant and permission check
        -> PostgreSQL transaction
           -> create/update task
           -> insert activity log
           -> insert outbox event
        -> API response with canonical task
Worker -> reads outbox -> Redis Pub/Sub
Go API WebSocket hub -> active authorized project clients
```

## Realtime rules

- REST is used for commands and initial/query data.
- WebSocket is used for live notification of committed changes.
- Browser does not treat a WebSocket event as permission to change data locally without server confirmation.
- Events can be duplicated or received out of order. Clients deduplicate by event ID and refetch if state is unclear.
- Redis failure must not lose a committed task change because the outbox event remains in PostgreSQL for retry.

## Group chat flow

```text
Member sends message -> API verifies project membership -> message stored in PostgreSQL
-> activity/outbox event stored -> worker publishes event -> WebSocket pushes message
-> other authorized project members update their chat timeline
```

## File upload flow

```text
Browser asks API for upload permission -> API validates task/project access and file policy
-> API creates short-lived signed upload URL -> browser uploads direct to object storage
-> browser confirms upload -> API persists attachment metadata in PostgreSQL
```
