package config

import (
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	ListenAddr  string
	StoragePath string
	AccessKeyID string
	SecretKey   string
}

func New() (*Config, error) {
	accessKeyID := os.Getenv("S3_ACCESS_KEY_ID")
	if accessKeyID == "" {
		return nil, fmt.Errorf("S3_ACCESS_KEY_ID environment variable is required")
	}

	secretKey := os.Getenv("S3_SECRET_ACCESS_KEY")
	if secretKey == "" {
		return nil, fmt.Errorf("S3_SECRET_ACCESS_KEY environment variable is required")
	}

	return &Config{
		ListenAddr:  getEnvDefault("S3_LISTEN_ADDR", ":8000"),
		StoragePath: getEnvDefault("S3_STORAGE_PATH", "./data"),
		AccessKeyID: accessKeyID,
		SecretKey:   secretKey,
	}, nil
}

func getEnvDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
