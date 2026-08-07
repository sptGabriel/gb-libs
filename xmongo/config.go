// Package xmongo holds the MongoDB pieces that mirror what xpg does for
// Postgres. Right now that means versioned migrations; the client and the
// transactioner live in the service until there is a second consumer for them.
package xmongo

import (
	"net/url"
	"strings"
	"time"
)

// Config describes how to reach MongoDB.
//
// Transactions require a replica set: a standalone mongod rejects
// startTransaction. Single-node replica sets are fine, and are what local
// development normally runs.
type Config struct {
	Hosts       []string      `env:"DB_MONGO_HOSTS" envSeparator:","`
	Database    string        `env:"DB_MONGO_DATABASE"`
	User        string        `env:"DB_MONGO_USER"`
	Password    string        `env:"DB_MONGO_PASSWORD"`
	AuthSource  string        `env:"DB_MONGO_AUTH_SOURCE" envDefault:"admin"`
	ReplicaSet  string        `env:"DB_MONGO_REPLICA_SET"`
	Direct      bool          `env:"DB_MONGO_DIRECT" envDefault:"false"`
	PoolMinSize uint64        `env:"DB_MONGO_POOL_MIN_SIZE" envDefault:"2"`
	PoolMaxSize uint64        `env:"DB_MONGO_POOL_MAX_SIZE" envDefault:"10"`
	Timeout     time.Duration `env:"DB_MONGO_TIMEOUT" envDefault:"5s"`
}

// ConnectionString builds the mongodb:// URI. It carries the password, so the
// result should not be logged.
func (c Config) ConnectionString() string {
	u := url.URL{
		Scheme: "mongodb",
		Host:   strings.Join(c.Hosts, ","),
		Path:   "/" + c.Database,
	}

	if c.User != "" && c.Password != "" {
		u.User = url.UserPassword(c.User, c.Password)
	}

	query := url.Values{}

	if c.User != "" && c.Password != "" {
		authSource := c.AuthSource
		if authSource == "" {
			authSource = "admin"
		}

		query.Set("authSource", authSource)
	}

	if c.ReplicaSet != "" {
		query.Set("replicaSet", c.ReplicaSet)
	}

	// Without this the driver follows discovery and dials the address the node
	// advertises — in development that is the port inside the container, not
	// the mapped one.
	if c.Direct {
		query.Set("directConnection", "true")
	}

	u.RawQuery = query.Encode()

	return u.String()
}
