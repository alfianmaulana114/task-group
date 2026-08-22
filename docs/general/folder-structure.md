# Repository and documentation structure

## Repository pillars

```text
task-group/
├── frontend/     Nuxt application; created by the Nuxt initializer later
├── backend/      Go module; created and organized when backend setup begins
├── api/          OpenAPI contract shared by frontend and backend
├── infra/        Docker, local infrastructure, and later deployment assets
└── docs/         Product and engineering context
```

Do not create framework-level source folders manually before their official initializer runs. Nuxt and Go setup will create the appropriate files at the correct time.

## Documentation pillars

```text
docs/
├── general/      Product decisions independent of implementation
├── frontend/     Nuxt UI, state, API consumption, and realtime behaviour
└── backend/      Go architecture, data, API, security, and deployment
```

## Ownership

- `general`: scope, actors, requirements, business rules, use cases, and end-to-end flows.
- `frontend`: what users see and how client state is synchronized. It does not define authorization rules.
- `backend`: persistence, authorization, API contracts, jobs, and operational behaviour. Backend is the source of truth.
- `api/openapi.yaml`: the executable agreement between frontend and backend. It must be updated with any API change.
