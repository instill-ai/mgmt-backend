package datamodel

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// Base contains common columns for all tables
type Base struct {
	UID        uuid.UUID      `gorm:"type:uuid;primary_key;<-:create"` // allow read and create
	CreateTime time.Time      `gorm:"autoCreateTime:nano"`
	UpdateTime time.Time      `gorm:"autoUpdateTime:nano"`
	DeleteTime gorm.DeletedAt `sql:"index"`
}

// User defines a user instance in the database
type User struct {
	Base
	ID                     string `gorm:"unique;not null;"`
	Email                  sql.NullString
	CompanyName            sql.NullString
	Role                   sql.NullString
	UsageDataCollection    bool `gorm:"default:false"`
	NewsletterSubscription bool `gorm:"default:false"`
}
