package platform

import (
	"context"
	"fmt"
	"time"

	"github.com/USERNAME-GITHUB-KAMU/task-group/backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
)

type Dependencies struct {
	Postgres *pgxpool.Pool
	Redis    *redis.Client
	MinIO    *minio.Client
	Bucket   string
}

func Connect(ctx context.Context, cfg config.Config) (*Dependencies, error) {
	pgPool, err := connectPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	redisClient := connectRedis(cfg)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		pgPool.Close()
		redisClient.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	minioClient, err := connectMinIO(cfg)
	if err != nil {
		pgPool.Close()
		redisClient.Close()
		return nil, err
	}

	if err := ensureBucket(ctx, minioClient, cfg.MinIOBucket); err != nil {
		pgPool.Close()
		redisClient.Close()
		return nil, err
	}

	return &Dependencies{
		Postgres: pgPool,
		Redis:    redisClient,
		MinIO:    minioClient,
		Bucket:   cfg.MinIOBucket,
	}, nil
}

func (d *Dependencies) Close() {
	if d == nil {
		return
	}

	if d.Redis != nil {
		d.Redis.Close()
	}

	if d.Postgres != nil {
		d.Postgres.Close()
	}
}

func (d *Dependencies) Ready(ctx context.Context) error {
	if d == nil {
		return fmt.Errorf("dependencies are not initialized")
	}

	if err := d.Postgres.Ping(ctx); err != nil {
		return fmt.Errorf("postgres ping: %w", err)
	}

	if err := d.Redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}

	exists, err := d.MinIO.BucketExists(ctx, d.Bucket)
	if err != nil {
		return fmt.Errorf("minio bucket check: %w", err)
	}
	if !exists {
		return fmt.Errorf("minio bucket %q not ready", d.Bucket)
	}

	return nil
}

func connectPostgres(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 1
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}

func connectRedis(cfg config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
}

func connectMinIO(cfg config.Config) (*minio.Client, error) {
	client, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("connect minio: %w", err)
	}

	return client, nil
}

func ensureBucket(ctx context.Context, client *minio.Client, bucket string) error {
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("check minio bucket: %w", err)
	}
	if exists {
		return nil
	}

	if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create minio bucket: %w", err)
	}

	return nil
}
