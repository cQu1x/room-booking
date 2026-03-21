package config

import (
	"fmt"
	"os"
)

type Config struct {
	App AppConfig
	DB  DBConfig
	JWT JWTConfig
}

type AppConfig struct {
	Port string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

type JWTConfig struct {
	Secret string
}

func LoadConfig() Config {
	return Config{
		App: AppConfig{
			Port: GetEnv("APP_PORT", "8080"),
		},
		DB: DBConfig{
			Host:     GetEnv("DB_HOST", "localhost"),
			Port:     GetEnv("DB_PORT", "5432"),
			User:     GetEnv("DB_USER", "postgres"),
			Password: GetEnv("DB_PASSWORD", "password"),
			Name:     GetEnv("DB_NAME", "booking"),
			SSLMode:  GetEnv("DB_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			Secret: GetEnv("JWT_SECRET", "secret"),
		},
	}

}

func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
