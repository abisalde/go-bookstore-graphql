package config

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"

	"entgo.io/ent/dialect"
	"github.com/abisalde/go-bookstore-graphql/ent"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db   *ent.Client
	once sync.Once
)

// Connect initializes the Ent client and runs migrations.
// It should be called once at application startup.
func Connect() *ent.Client {
	once.Do(func() {
		// Create data directory if it doesn't exist
		if err := os.MkdirAll("data", 0755); err != nil {
			log.Fatalf("🛑 Failed to create data directory: %v", err)
		}

		// Connect to SQLite database file
		dbPath := filepath.Join("data", "bookstore.db")
		client, err := ent.Open(
			dialect.SQLite,
			"file:"+dbPath+"?cache=shared&_fk=1",
			ent.Log(func(args ...any) {
				log.Println(args...)
			}),
		)

		if err != nil {
			log.Fatalf("🛑 Failed opening connection to sqlite: %v", err)
		}

		// Run the auto migration tool.
		if err := client.Schema.Create(context.Background()); err != nil {
			log.Fatalf("🛑 Failed creating schema resources: %v", err)
		}

		db = client
	})
	return db
}

// GetDB returns the initialized Ent client.
func GetDB() *ent.Client {
	if db == nil {
		log.Fatal("🛑 Database not connected. Call config.Connect() first.")
	}
	return db
}

func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
