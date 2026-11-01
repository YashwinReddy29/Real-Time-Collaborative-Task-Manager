package main

import (
	"fmt"
	"log"
	"os"

	"taskflow/internal/repository"
)

func main() {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		log.Fatal("TEST_DATABASE_URL must be set")
	}

	pool, err := repository.NewPool(url)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := repository.Migrate(pool); err != nil {
		log.Fatal(err)
	}

	if err := repository.MigratePasswordResets(pool); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Test database schema initialized successfully")
}
