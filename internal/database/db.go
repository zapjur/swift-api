package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func ConnectDB() (*sql.DB, error) {
	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		connStr = "postgres://user:pass@localhost:5432/swift?sslmode=disable"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Println("Error connecting to database:", err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		log.Println("Database ping failed:", err)
		return nil, err
	}

	log.Println("Connected to database successfully!")
	return db, nil
}
