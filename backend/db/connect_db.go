package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func NewDB() *sql.DB {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Could not load .env file")
	}

	dbConn, ok := os.LookupEnv("DATABASE_URL")

	if !ok {
		log.Fatal("Error: DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", dbConn)
	if err != nil {
		log.Fatalf("Error: Could not open database connection: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error: Could not ping database: %v", err)
	}

	fmt.Println("Successfully connected to the database!")
	return db
}
