# Deployment and storage

## Environments

- Local development: Docker Desktop runs frontend, API, worker, PostgreSQL, Redis, and MinIO.
- Production: CI builds container images; source code is not mounted; configuration and secrets are injected by the deployment platform.

## Recommended production separation

- Nuxt and Go API: stateless containers.
- Worker: a separate container from the same Go codebase.
- PostgreSQL and Redis: managed services when budget permits.
- Attachments: managed S3-compatible object storage.
- HTTPS/CDN/WAF: an edge provider such as Cloudflare.

## S3-compatible object storage

S3 is object storage, not a relational database. Store attachment bytes there and retain metadata, ownership, and authorization in PostgreSQL.

MinIO is an S3-compatible server used locally through Docker. It lets development code use the same bucket/object API shape as cloud object storage.

Amazon S3 is suitable for production and is pay-as-you-go, not permanently free. AWS currently advertises Free Tier credits for eligible new accounts, but storage, requests, and data transfer can incur charges. Set AWS Budgets and billing alerts before use.

Use private buckets, least-privilege IAM, signed URLs, encryption, lifecycle rules, and backup/versioning appropriate to the product. Do not put relational data, sessions, or application state in S3.

## Production deployment checklist

- Run database migrations before a compatible application release.
- Use health/readiness checks for rollout.
- Keep secrets outside images and the repository.
- Create PostgreSQL backups and test restoring them.
- Configure logs, metrics, error alerts, and cost alerts.
- Keep object-storage buckets private and restrict direct access.
