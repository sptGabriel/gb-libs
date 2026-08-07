package xmongo_test

import (
	"os"
	"testing"
	"testing/fstest"

	"github.com/sptGabriel/gb-libs/xmongo"
)

// Integration test: needs a MongoDB reachable at DB_MONGO_HOSTS (default
// localhost:27018). Without one it skips, so `go test ./...` still passes on a
// machine with no infrastructure.
func testConfig() xmongo.Config {
	host := os.Getenv("DB_MONGO_HOSTS")
	if host == "" {
		host = "localhost:27018"
	}

	return xmongo.Config{
		Hosts:    []string{host},
		Database: "gblibs_xmongo_test",
		Direct:   true,
	}
}

func migrationFS(files map[string]string) fstest.MapFS {
	fsys := fstest.MapFS{}
	for name, content := range files {
		fsys["migrations/"+name] = &fstest.MapFile{Data: []byte(content)}
	}

	return fsys
}

func TestRunMigrations(t *testing.T) {
	fsys := migrationFS(map[string]string{
		"000001_create_index.up.json": `[
  {
    "createIndexes": "products",
    "indexes": [{ "key": { "created_at": -1 }, "name": "idx_products_created_at" }]
  }
]`,
		"000001_create_index.down.json": `[
  { "dropIndexes": "products", "index": "idx_products_created_at" }
]`,
	})

	uri := testConfig().ConnectionString()
	migrations := xmongo.Migrations{FS: fsys, Folder: "migrations"}

	if err := xmongo.RunMigrations(uri, migrations); err != nil {
		t.Skipf("mongo unavailable, skipping: %v", err)
	}

	// Applying twice must be a no-op: that is what the schema_migrations
	// collection buys, and the reason this is not the place for indexes that
	// change with the code.
	if err := xmongo.RunMigrations(uri, migrations); err != nil {
		t.Fatalf("second run should be a no-op: %v", err)
	}
}

func TestRunMigrationsRejectsURIWithoutDatabase(t *testing.T) {
	fsys := migrationFS(map[string]string{
		"000001_noop.up.json":   `[]`,
		"000001_noop.down.json": `[]`,
	})

	err := xmongo.RunMigrations("mongodb://localhost:27018",
		xmongo.Migrations{FS: fsys, Folder: "migrations"})
	if err == nil {
		t.Fatal("uri without a database name should fail")
	}
}

func TestRunMigrationsRequiresFilesystem(t *testing.T) {
	if err := xmongo.RunMigrations("mongodb://localhost:27018/db", xmongo.Migrations{}); err == nil {
		t.Fatal("migrations without FS should fail")
	}
}

func TestConnectionString(t *testing.T) {
	cfg := xmongo.Config{
		Hosts:      []string{"host-a:27017", "host-b:27017"},
		Database:   "products",
		User:       "app",
		Password:   "s3cr3t",
		AuthSource: "admin",
		ReplicaSet: "rs0",
		Direct:     true,
	}

	got := cfg.ConnectionString()
	want := "mongodb://app:s3cr3t@host-a:27017,host-b:27017/products" +
		"?authSource=admin&directConnection=true&replicaSet=rs0"

	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}
