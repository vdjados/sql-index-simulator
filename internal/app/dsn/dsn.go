package dsn

import (
	"fmt"
	"os"
)

// FromEnv формирует строку подключения к PostgreSQL на основе переменных окружения.
// Используется та же схема, что и в примере rip2026-lab2-postgres-integration.
func FromEnv() string {
	host := os.Getenv("DB_HOST")
	if host == "" || host == "localhost" {
		host = "127.0.0.1"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}

	pass := os.Getenv("DB_PASS")
	if pass == "" {
		pass = "postgres"
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "mydb"
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbname)
}

