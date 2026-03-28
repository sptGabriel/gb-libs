package config

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/caarlos0/env/v11"
)

var (
	once        sync.Once
	appConfig   *AppConfig
	allowedEnvs []string = []string{"local", "development", "testing", "homolog", "staging", "production"}
)

type AppConfig struct {
	OTEL
	PGDB        PGDB
	Name        string `env:"APP_NAME,required"`
	Version     string `env:"APP_VERSION,required"`
	Environment string `env:"ENVIRONMENT,required"`
	extras      map[string]string
}

type OTEL struct {
	Enabled         bool   `env:"OTEL_ENABLED"`
	LogPath         string `env:"OTEL_LOG_PATH"`
	LogEndpoint     string `env:"OTEL_LOG_ENDPOINT"`
	OTLPSecure      bool   `env:"OTEL_OTLP_SECURE"`
	MetricsEndPoint string `env:"OTEL_METRICS_ENDPOINT"`
	TraceEndpoint   string `env:"OTEL_TRACE_ENDPOINT"`
}

type PGDB struct {
	DatabaseName        string        `env:"DB_NAME"`
	User                string        `env:"DB_USER"`
	Password            string        `env:"DB_PASS"`
	Host                string        `env:"DB_HOST"`
	Port                uint16        `env:"DB_PORT" envDefault:"5433"`
	Schema              string        `env:"DB_SCHEMA" envDefault:"public"`
	PoolMinSize         int           `env:"DB_POOL_MIN_SIZE" envDefault:"2"`
	PoolMaxSize         int           `env:"DB_POOL_MAX_SIZE" envDefault:"10"`
	PoolMaxConnLifetime time.Duration `env:"DB_POOL_MAX_CONN_LIFETIME" envDefault:"30m"`
	PoolMaxConnIdleTime time.Duration `env:"DB_POOL_MAX_CONN_IDLE_TIME" envDefault:"5m"`
	SSLMode             string        `env:"DB_SSLMODE" envDefault:"disable"`
}

func (c PGDB) ConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.DatabaseName,
		c.SSLMode,
		c.Schema,
	)
}

func (c *AppConfig) LoadDynamic(keys ...string) {
	for _, key := range keys {
		value := strings.TrimSpace(os.Getenv(key))
		if value == "" {
			log.Printf("%s is not set in .env or system enviroment", key)
		}

		c.extras[key] = value
	}

}

func (c *AppConfig) GetExtra(key string) string { return c.extras[key] }

func LoadConfig() *AppConfig {
	once.Do(func() {
		cfg, err := env.ParseAs[AppConfig]()
		if err != nil {
			panic("error to load config")
		}

		cfg.extras = make(map[string]string)

		envName := strings.ToLower(strings.TrimSpace(cfg.Environment))
		if !slices.Contains(allowedEnvs, envName) {
			panic(fmt.Sprintf("invalid env value %s", envName))
		}

		cfg.Environment = envName

		appConfig = &cfg
	})

	return appConfig
}
