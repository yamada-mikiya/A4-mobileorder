package connectDB

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func NewDB() (*sqlx.DB, func()) {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Could not load .env file")
	}

	dbConn, ok := os.LookupEnv("DATABASE_URL")

	if !ok {
		log.Fatal("Error: DATABASE_URL environment variable is not set")
	}

	db, err := sqlx.Connect("postgres", dbConn)
	if err != nil {
		log.Fatalf("Error: Could not open database connection: %v", err)
	}

	fmt.Println("Successfully connected to the database!")

	closer := func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing the database: %v\n", err)
		}
	}
	return db, closer
}
