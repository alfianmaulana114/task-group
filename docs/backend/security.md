# Security

## Authentication

- Passwords are hashed with Argon2id or bcrypt; plaintext is never stored.
- Use a short-lived access token and rotating refresh token.
- Store refresh token in a `HttpOnly`, `Secure`, `SameSite` cookie; store only its hash in PostgreSQL.
- Revoke sessions on logout, password change, and suspicious security events.

## Authorization and tenancy

- Authenticate first, then verify organization scope, then verify permission/ownership.
- Frontend route guards are for experience only; Go enforces every action.
- Do not expose cross-organization resources through predictable IDs, cache keys, events, or logs.

## Web and API controls

- Validate all request payloads and allowlist sortable/filterable fields.
- Use parameterized sqlc queries; never concatenate user input into SQL.
- Configure CORS narrowly, set security headers, apply rate limits to authentication and upload endpoints, and create a request/correlation ID.
- Do not log passwords, tokens, session cookies, or object-storage credentials.

## Uploads

- Validate authorization, declared and detected MIME type, byte size, and permitted extension.
- Prefer short-lived signed upload/download URLs.
- Keep buckets private; do not make all attachments public.
- Consider malware scanning before production launch.

## Secrets

- `.env` is local only and never committed.
- Commit `.env.example` with placeholders only.
- Production uses a cloud secret manager or platform-managed secrets.
