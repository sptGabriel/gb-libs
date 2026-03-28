package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	Id        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"not null; default:now()"`
	UpdatedAt time.Time      `gorm:"not null; default:now()"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
