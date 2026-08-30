# Backend Migrations

Folder ini berisi migration SQL PostgreSQL untuk project `task-group`.

## File

- `0001_initial_schema.up.sql` = membuat schema awal
- `0001_initial_schema.down.sql` = menghapus schema awal

## Cara pakai

Jika nanti kamu pakai migration tool, arahkan tool itu ke folder ini.

Kalau belum pakai tool, file ini tetap berguna sebagai blueprint schema yang bisa dibaca sebelum coding repository, service, dan handler.

## Urutan kerja yang disarankan

1. Jalankan Docker Compose untuk PostgreSQL
2. Terapkan migration
3. Seed data minimal
4. Baru mulai bikin endpoint backend
