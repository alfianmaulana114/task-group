# Backend

Go backend untuk `task-group`.

## Tujuan tahap awal

- baca konfigurasi dari environment
- siapkan HTTP server dasar
- siapkan fondasi untuk koneksi PostgreSQL, Redis, dan MinIO

## Struktur awal

- `cmd/api` = entrypoint aplikasi HTTP
- `internal/config` = konfigurasi runtime
- `internal/server` = HTTP server dan routing
- `migrations` = schema PostgreSQL

## Cara jalanin

Set environment variable dari `backend/.env` lalu jalankan:

```powershell
cd "D:\alfian\Projek Web Liburan\task-group\backend"
go run ./cmd/api
```

