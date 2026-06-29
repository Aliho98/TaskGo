package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/lib/pq" //  PostgreSQL driver
)

var DB *sql.DB

func Connect() error {
	connStr := "host=localhost port=5432 user=postgres password= dbname=postgres sslmode=disable"
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("Error connecting to database: %v", err)
	}
	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("Error pinging database: %v", err)
	}
	log.Println("Connected to database")
	return nil
}
func RunMigrations(dsn string) error {
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}
	log.Println("Migrations applied.")
	return nil
}

func Close() error {
	if DB != nil {
		DB.Close()
	}
	return nil
}

func GetDB() *sql.DB {
	return DB
}
