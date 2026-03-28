package xotel

import (
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/gorm"
)

func InstrumentDb(db *gorm.DB, opts ...otelgorm.Option) (*gorm.DB, error) {
	if err := db.Use(otelgorm.NewPlugin(opts...)); err != nil {
		return db, err
	}
	return db, nil
}
