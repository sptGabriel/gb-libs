package xgorm

import (
	"fmt"

	"github.com/sptGabriel/defaults/xotel"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupOrm(cfg Config) (*gorm.DB, error) {
	dsn := cfg.Url()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connecting to the database: %v", err)
	}

	db, err = xotel.InstrumentDb(db)
	if err != nil {
		return nil, fmt.Errorf("on setup telemetry: %v", err)
	}

	return db, nil
}
