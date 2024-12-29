package config

import (
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	PORT        string
	StoragePath string
	AccessKeyID string
	SecretKey   string
}

func New() (*Config, error) {
	accessKeyID := os.Getenv("ACCESS_KEY_ID")
	if accessKeyID == "" {
		return nil, fmt.Errorf("ACCESS_KEY_ID environment variable is required")
	}

	secretKey := os.Getenv("SECRET_ACCESS_KEY")
	if secretKey == "" {
		return nil, fmt.Errorf("SECRET_ACCESS_KEY environment variable is required")
	}

	log.Println("Access Key ID:", accessKeyID)
	log.Println("Secret Key:", secretKey)
	log.Println("Storage Path:", os.Getenv("STORAGE_PATH"))
	log.Println("Port:", os.Getenv("PORT"))

	return &Config{
		StoragePath: getEnvDefault("STORAGE_PATH", "./data"),
		PORT:        getEnvDefault("PORT", "8191"),
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
