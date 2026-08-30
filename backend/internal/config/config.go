package config

import (
	"bufio"
	"fmt"
	"os"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv              string
	HTTPPort            string
	DatabaseURL         string
	RedisAddr           string
	RedisPassword       string
	RedisDB             int
	MinIOEndpoint       string
	MinIOAccessKey      string
	MinIOSecretKey      string
	MinIOBucket         string
	MinIOUseSSL         bool
	JWTSecret           string
	CORSAllowedOrigins  []string
	ShutdownTimeout     time.Duration
}

func Load() (Config, error) {
	_ = loadDotEnv(".env")

	cfg := Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		HTTPPort:           getEnv("HTTP_PORT", "8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      os.Getenv("REDIS_PASSWORD"),
		MinIOEndpoint:      getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey:     getEnv("MINIO_SECRET_KEY", "minioadmin123"),
		MinIOBucket:        getEnv("MINIO_BUCKET", "task-group"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		ShutdownTimeout:    10 * time.Second,
	}

	var err error
	cfg.RedisDB, err = getEnvInt("REDIS_DB", 0)
	if err != nil {
		return Config{}, err
	}

	cfg.MinIOUseSSL, err = getEnvBool("MINIO_USE_SSL", false)
	if err != nil {
		return Config{}, err
	}

	cfg.CORSAllowedOrigins = splitAndTrim(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"))

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}

	if err := validatePort(cfg.HTTPPort); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func loadDotEnv(filename string) error {
	file, err := os.Open(filepath.Clean(filename))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)

		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", key)
	}
	return value, nil
}

func getEnvBool(key string, fallback bool) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean", key)
	}
	return value, nil
}

func splitAndTrim(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func validatePort(port string) error {
	if port == "" {
		return fmt.Errorf("HTTP_PORT is required")
	}

	if _, err := net.LookupPort("tcp", port); err != nil {
		return fmt.Errorf("HTTP_PORT must be a valid TCP port")
	}

	return nil
}
