package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Cookie   CookieConfig
	CORS     CORSConfig
	Argon2   Argon2Config
}

type ServerConfig struct {
	Addr string
}

type DatabaseConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name)
}

type JWTConfig struct {
	AccessSecret string
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
}

type CookieConfig struct {
	RefreshName string
	Secure      bool
	SameSite    string
	Path        string
}

type CORSConfig struct {
	AllowOrigins []string
}

type Argon2Config struct {
	MemoryKiB   uint32
	Iterations  uint32
	Threads     uint8
	SaltLength  uint32
	KeyLength   uint32
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	if accessSecret == "" {
		return nil, fmt.Errorf("JWT_ACCESS_SECRET is required")
	}

	accessTTL, err := durationEnv("JWT_ACCESS_TTL", 15*time.Minute)
	if err != nil {
		return nil, err
	}
	refreshTTL, err := durationEnv("REFRESH_TTL", 7*24*time.Hour)
	if err != nil {
		return nil, err
	}

	corsOrigins := splitCSV(envOr("CORS_ALLOW_ORIGINS", "http://localhost:3000"))

	argon2Memory, err := uint32Env("ARGON2_MEMORY_KIB", 65536)
	if err != nil {
		return nil, err
	}
	argon2Iter, err := uint32Env("ARGON2_ITERATIONS", 3)
	if err != nil {
		return nil, err
	}
	argon2Threads, err := uint8Env("ARGON2_THREADS", 2)
	if err != nil {
		return nil, err
	}

	return &Config{
		Server: ServerConfig{
			Addr: envOr("SERVER_ADDR", ":8080"),
		},
		Database: DatabaseConfig{
			User:     envOr("DB_USER", "root"),
			Password: os.Getenv("DB_PASSWORD"),
			Host:     envOr("DB_HOST", "127.0.0.1"),
			Port:     envOr("DB_PORT", "3306"),
			Name:     envOr("DB_NAME", "backend"),
		},
		JWT: JWTConfig{
			AccessSecret: accessSecret,
			AccessTTL:    accessTTL,
			RefreshTTL:   refreshTTL,
		},
		Cookie: CookieConfig{
			RefreshName: envOr("COOKIE_REFRESH_NAME", "refresh_token"),
			Secure:      boolEnv("COOKIE_SECURE", false),
			SameSite:    envOr("COOKIE_SAME_SITE", "Lax"),
			Path:        envOr("COOKIE_REFRESH_PATH", "/api/auth/refresh"),
		},
		CORS: CORSConfig{
			AllowOrigins: corsOrigins,
		},
		Argon2: Argon2Config{
			MemoryKiB:  argon2Memory,
			Iterations: argon2Iter,
			Threads:    argon2Threads,
			SaltLength: 16,
			KeyLength:  32,
		},
	}, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid duration: %w", key, err)
	}
	return d, nil
}

func boolEnv(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func uint32Env(key string, fallback uint32) (uint32, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid integer: %w", key, err)
	}
	return uint32(n), nil
}

func uint8Env(key string, fallback uint8) (uint8, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.ParseUint(v, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid integer: %w", key, err)
	}
	return uint8(n), nil
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
