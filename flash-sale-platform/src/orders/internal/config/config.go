package config

import (
	"log"
	"os"
	"time"
)

type Config struct {
	// Database
	DBUser     string
	DBPassword string
	DBHost     string
	DBName     string

	// Redis
	RedisAddr     string
	RedisPassword string

	// SQS
	SQSQueueURL string
	AWSRegion   string

	// Server
	Port string

	// Locking
	LockingStrategy string // "none", "pessimistic", "optimistic", "queue"
	LockTimeout     time.Duration
}

func Load() *Config {
	cfg := &Config{
		DBUser:     getEnv("DB_USER", ""),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBHost:     getEnv("DB_HOST", ""),
		DBName:     getEnv("DB_NAME", ""),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		SQSQueueURL: getEnv("SQS_QUEUE_URL", ""),
		AWSRegion:   getEnv("AWS_REGION", "us-east-1"),

		Port: getEnv("PORT", "8080"),

		LockingStrategy: getEnv("LOCKING_STRATEGY", "pessimistic"),
		LockTimeout:     getEnvDuration("LOCK_TIMEOUT", 5*time.Second),
	}

	log.Printf("Configuration loaded:")
	log.Printf("DB Host: %s", cfg.DBHost)
	log.Printf("Redis: %s", cfg.RedisAddr)
	log.Printf("Locking Strategy: %s", cfg.LockingStrategy)

	return cfg
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}
