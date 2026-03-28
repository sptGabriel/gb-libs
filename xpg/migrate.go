package xpg

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// Migrations groups information about SQL migrations.
type Migrations struct {
	// Folder name of the folder in which the migration .sql files are located
	Folder string

	// FS filesystem representing a migrations folder
	FS fs.FS
}

func RunMigrations(connString string, migrations Migrations) error {
	pgxConfig, err := pgx.ParseConfig(connString)
	if err != nil {
		return fmt.Errorf("error parsing pgx ConnString: %w", err)
	}

	return RunMigrationsByConnConfig(pgxConfig, migrations)
}

func RunMigrationsByConnConfig(pgxConfig *pgx.ConnConfig, migrations Migrations) error {
	handler, err := GetMigrationHandlerByConnConfig(pgxConfig, migrations)
	if err != nil {
		return err
	}

	if err := handler.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	srcErr, dbErr := handler.Close()
	if srcErr != nil {
		return fmt.Errorf("[migrations] failed to close DB source: %w", srcErr)
	}
	if dbErr != nil {
		return fmt.Errorf("[migrations] failed to close migrations repositories connection: %w", dbErr)
	}

	return nil
}

func GetMigrationHandler(connString string, migrations Migrations) (*migrate.Migrate, error) {
	pgxConfig, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("error parsing pgx ConnString: %w", err)
	}

	return GetMigrationHandlerByConnConfig(pgxConfig, migrations)
}

func GetMigrationHandlerByConnConfig(pgxConfig *pgx.ConnConfig, migrations Migrations) (*migrate.Migrate, error) {
	driver, err := postgres.WithInstance(stdlib.OpenDB(*pgxConfig), &postgres.Config{
		DatabaseName: pgxConfig.Database,
	})
	if err != nil {
		return nil, fmt.Errorf("[migrations] failed to get postgres driver: %w", err)
	}

	if migrations.FS == nil {
		handler, mErr := migrate.NewWithDatabaseInstance("file://"+migrations.Folder, pgxConfig.Database, driver)
		if mErr != nil {
			return nil, fmt.Errorf("[migrations] failed to create migrate: %w", mErr)
		}

		return handler, nil
	}

	source, err := httpfs.New(http.FS(migrations.FS), migrations.Folder)
	if err != nil {
		return nil, fmt.Errorf("[migrations] failed to create httpfs driver: %w", err)
	}

	handler, err := migrate.NewWithInstance("httpfs", source, pgxConfig.Database, driver)
	if err != nil {
		return nil, fmt.Errorf("[migrations] failed to create migrate source instance: %w", err)
	}

	return handler, nil
}
