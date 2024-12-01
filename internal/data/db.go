// internal/data/db.go

package data

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var (
	db   *sql.DB
	once sync.Once
)

// Config holds database configuration
type Config struct {
	URL   string
	Token string
}

// Initialize sets up the database connection
func Initialize(cfg Config) error {
	var err error
	once.Do(func() {
		connStr := fmt.Sprintf("%s?authToken=%s", cfg.URL, cfg.Token)
		log.Printf("Attempting to connect to database with URL: %s", cfg.URL)
		db, err = sql.Open("libsql", connStr)
		if err != nil {
			log.Printf("Error opening database: %v", err)
			return
		}

		// Test the connection
		err = db.Ping()
		if err != nil {
			log.Printf("Error connecting to the database: %v", err)
			return
		}
		log.Printf("Successfully connected to database")
		// Set connection pool settings
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(25)
	})

	return err
}

// GetDB returns the database instance
func GetDB() *sql.DB {
	return db
}

// Close closes the database connection
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
