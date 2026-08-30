# Infrastructure

Folder ini dipakai buat service pendukung project:

- PostgreSQL
- Redis
- MinIO

## Cara Pakai

1. Salin `infra/.env.example` jadi `infra/.env`
2. Sesuaikan password kalau perlu
3. Jalankan dari folder `infra`

```powershell
cd "D:\alfian\Projek Web Liburan\task-group\infra"
docker compose up -d
```

## Port Default

- PostgreSQL: `5432`
- Redis: `6379`
- MinIO API: `9000`
- MinIO Console: `9001`

## Env Untuk App

- Backend pakai `backend/.env.example`
- Frontend pakai `frontend/.env.example`

Kalau nanti backend sudah punya config loader, arahkan `DATABASE_URL`, `REDIS_ADDR`, dan variabel MinIO ke nilai yang sesuai.
