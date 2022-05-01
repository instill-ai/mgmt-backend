package datamodel

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
)

// Base contains common columns for all tables
type Base struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreateTime time.Time
	UpdateTime time.Time
	DeleteTime *time.Time `sql:"index"`
}

type User struct {
	Base
	Email                  sql.NullString
	Login                  string `gorm:"unique;not null;"`
	CompanyName            sql.NullString
	Role                   sql.NullString
	UsageDataCollection    bool `gorm:"default:false"`
	NewsletterSubscription bool `gorm:"default:false"`
}
