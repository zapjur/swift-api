package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func ConnectDB() (*sql.DB, error) {
	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		return nil, fmt.Errorf("DB_URL environment variable is required")
	}

	var db *sql.DB

	err := retry(10, 2*time.Second, func() error {
		var err error
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Println("Failed to open connection:", err)
			return err
		}
		err = db.Ping()
		if err != nil {
			log.Println("Failed to ping database:", err)
			return err
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}

	log.Println("Connected to database successfully!")
	return db, nil
}

func retry(maxRetries int, delay time.Duration, fn func() error) error {
	for i := 0; i < maxRetries; i++ {
		err := fn()
		if err == nil {
			return nil
		}
		log.Printf("Retry %d/%d failed, retrying in %s...\n", i+1, maxRetries, delay)
		time.Sleep(delay)
	}
	return fmt.Errorf("all %d retry attempts failed", maxRetries)
}
