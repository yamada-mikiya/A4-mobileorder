package services_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found")
	}

	if err := setupTestDatabase(); err != nil {
		log.Printf("failed to setup test database: %s", err)
		log.Println("Skipping database-dependent tests")
		os.Exit(0)
	}

	os.Exit(m.Run())
}

func setupTestDatabase() error {
	dbConn, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		return fmt.Errorf("DATABASE_URL is not set")
	}

	db, err := sqlx.Connect("postgres", dbConn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	schema, err := os.ReadFile("../repositories/testdata/schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	db.MustExec(string(schema))

	return nil
}
