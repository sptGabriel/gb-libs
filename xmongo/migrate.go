package xmongo

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mongodb" // mongodb:// driver
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Migrations groups information about MongoDB migrations, mirroring xpg.
//
// The files are JSON rather than SQL: each one is an array of commands for
// db.runCommand, numbered the same way (000001_name.up.json plus its .down.json).
// A schema_migrations collection records what already ran, and migrate takes a
// lock before applying — two instances booting together do not collide.
//
//	//go:embed migrations/*.json
//	var migrationsFS embed.FS
//
//	err := xmongo.RunMigrations(cfg.ConnectionString(), xmongo.Migrations{
//	    FS:     migrationsFS,
//	    Folder: "migrations",
//	})
//
// Use this for what changes data — renaming a field, backfilling, fixing a type.
// Indexes and validators are a different problem: they are idempotent by nature
// and change together with the code that uses them, so they belong in a
// desired-state helper that runs on every boot. Versioning "make sure this index
// exists" only adds ceremony to an operation that was always safe to repeat.
type Migrations struct {
	// Folder is the directory inside FS holding the migration files.
	Folder string
	// FS is the filesystem with the files, usually an embed.FS.
	FS fs.FS
}

type migrateSettings struct {
	collection      string
	transactionMode bool
}

// Option tweaks how migrations are applied.
type Option func(*migrateSettings)

// WithMigrationsCollection overrides the collection that records applied
// versions. Defaults to schema_migrations.
func WithMigrationsCollection(name string) Option {
	return func(s *migrateSettings) { s.collection = name }
}

// WithTransaction wraps each migration in a transaction, which requires a
// replica set.
//
// Off by default because not every migration fits in one: MongoDB transactions
// are bounded in time and oplog size, and a backfill over millions of documents
// blows through both. Turn it on for small migrations, where applying half of it
// would be worse than applying none.
func WithTransaction() Option {
	return func(s *migrateSettings) { s.transactionMode = true }
}

// RunMigrations applies pending migrations. It is idempotent across runs: what
// already ran is recorded and does not run again.
func RunMigrations(uri string, migrations Migrations, opts ...Option) error {
	if migrations.FS == nil {
		return fmt.Errorf("migrations: filesystem is required")
	}

	folder := migrations.Folder
	if folder == "" {
		folder = "."
	}

	source, err := iofs.New(migrations.FS, folder)
	if err != nil {
		return fmt.Errorf("migrations: reading %q: %w", folder, err)
	}

	target, err := migrateURI(uri, opts)
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, target)
	if err != nil {
		return fmt.Errorf("migrations: opening target: %w", err)
	}

	migrateErr := m.Up()
	if migrateErr != nil && !errors.Is(migrateErr, migrate.ErrNoChange) {
		// Close before leaving so neither the connection nor the lock is left
		// hanging.
		_, _ = m.Close()

		return fmt.Errorf("migrations: applying: %w", migrateErr)
	}

	sourceErr, dbErr := m.Close()
	if sourceErr != nil {
		return fmt.Errorf("migrations: closing source: %w", sourceErr)
	}

	if dbErr != nil {
		return fmt.Errorf("migrations: closing target: %w", dbErr)
	}

	return nil
}

// migrateURI appends the parameters the migration driver reads. They travel in
// the query string because that is the only channel migrate.New offers.
func migrateURI(uri string, opts []Option) (string, error) {
	s := &migrateSettings{}
	for _, opt := range opts {
		opt(s)
	}

	parsed, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("migrations: parsing uri: %w", err)
	}

	if strings.Trim(parsed.Path, "/") == "" {
		return "", fmt.Errorf("migrations: uri has no database name")
	}

	query := parsed.Query()

	if s.collection != "" {
		query.Set("x-migrations-collection", s.collection)
	}

	if s.transactionMode {
		query.Set("x-transaction-mode", "true")
	}

	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}
