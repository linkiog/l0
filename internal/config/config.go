package config

import (
	"os"
)

type Config struct {
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresHost     string
	PostgresPort     string

	KafkaBrokers string
	KafkaTopic   string
}

func NewConfig() *Config {
	cfg := &Config{
		PostgresUser:     getEnv("POSTGRES_USER", "myuser"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "mypassword"),
		PostgresDB:       getEnv("POSTGRES_DB", "mydb"),
		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		KafkaBrokers:     getEnv("KAFKA_BOOTSTRAP_SERVERS", "kafka_local:9092"),
		KafkaTopic:       getEnv("KAFKA_TOPIC", "orders"),
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
