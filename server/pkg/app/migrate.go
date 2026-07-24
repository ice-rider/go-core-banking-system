package app

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
)

// RunMigrations applies database migrations from the given path.
func RunMigrations(dsn, migrationsPath string) {
	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", err)
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to run migrations: %v", err)
	}

	log.Println("Migrations applied successfully")
}
