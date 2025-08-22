package repositories_test

import (
	"log"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func NewTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	dbConn, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		t.Skip("DATABASE_URL environment variable is not set")
	}

	db, err := sqlx.Connect("postgres", dbConn)
	if err != nil {
		t.Skipf("Could not open database connection: %v", err)
	}

	return db
}

func TestMain(m *testing.M) {
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found")
	}

	dbConn, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sqlx.Connect("postgres", dbConn)
	if err != nil {
		log.Printf("failed to connect to db for schema setup: %s", err)
		log.Println("Skipping database-dependent tests")
		os.Exit(0) // テストをスキップして正常終了
	}

	schema, err := os.ReadFile("./testdata/schema.sql")
	if err != nil {
		log.Fatalf("failed to read schema file: %s", err)
	}
	db.MustExec(string(schema))

	db.Close()

	os.Exit(m.Run())
}
