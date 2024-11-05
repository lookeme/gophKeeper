package db

import (
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// embedMigrations is a variable of type `embed.FS` that represents
// the embedded file system where the migration files are stored.
// It is used to set the base file system for the `goose` migration
// tool to read the migration files from.
//
// Example usage:
//
//	func StartMigration(pool *pgxpool.Pool) error {
//	    db := stdlib.OpenDBFromPool(pool)
//	    goose.SetBaseFS(embedMigrations)
//	    if err := goose.SetDialect("postgres"); err != nil {
//	        return err
//	    }
//	    if err := goose.Up(db, "migrations"); err != nil {
//	        return err
//	    }
//	    return nil
//	}
//
//go:embed migrations/*.sql
var embedMigrations embed.FS

// StartMigration starts the migration process for a PostgreSQL database using the provided connection pool.
// It opens a database connection from the pool and sets the base file system for embedded migrations.
// Then it sets the dialect to "postgres" and runs the migrations using the specified directory.
// If any error occurs during the migration process, it is returned.
// Otherwise, it returns nil.
func StartMigration(pool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pool)
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}
	return nil
}
