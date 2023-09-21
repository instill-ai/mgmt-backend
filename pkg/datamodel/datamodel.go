package datamodel

import (
	"database/sql"
	"database/sql/driver"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"

	mgmtPB "github.com/instill-ai/protogen-go/base/mgmt/v1alpha"
)

type TokenState int32

const (
	// State: UNSPECIFIED
	STATE_UNSPECIFIED TokenState = 0
	// State: INACTIVE
	STATE_INACTIVE TokenState = 1
	// State: ACTIVE
	STATE_ACTIVE TokenState = 2

	// In db, we should use current_time < expire_tiem to determine expire or not
	// We'll convert current_time > expire_tiem to pb's STATE_EXPIRED
	// This can be improved when the sync between db, redis are more robust
	// State: EXPIRED
	// STATE_EXPIRED TokenState = 3
)

// Base contains common columns for all tables
type Base struct {
	UID        uuid.UUID `gorm:"type:uuid;primary_key;<-:create"` // allow read and create, but not update
	CreateTime time.Time `gorm:"autoCreateTime:nano;<-:create"`   // allow read and create, but not update
	UpdateTime time.Time `gorm:"autoUpdateTime:nano"`
	// TODO: support DeleteTime gorm.DeletedAt `sql:"index"`
}

// User defines a user instance in the database
type User struct {
	Base
	ID                     string `gorm:"unique;not null;"`
	OwnerType              sql.NullString
	Email                  string `gorm:"unique;not null;"`
	CustomerId             string
	FirstName              sql.NullString
	LastName               sql.NullString
	OrgName                sql.NullString
	Role                   sql.NullString
	NewsletterSubscription bool `gorm:"default:false"`
	CookieToken            sql.NullString
}

type Password struct {
	Base
	PasswordHash       sql.NullString
	PasswordUpdateTime time.Time
}

func (Password) TableName() string {
	return "user"
}

// Token defines a api token instance in the database
type Token struct {
	Base
	ID          string
	Owner       string
	AccessToken string
	State       TokenState
	TokenType   string
	LastUseTime time.Time
	ExpireTime  time.Time
}

func (token *Token) BeforeCreate(db *gorm.DB) error {
	uuid, err := uuid.NewV4()
	if err != nil {
		return err
	}
	db.Statement.SetColumn("UID", uuid)
	return nil
}

// Scan function for custom GORM type PipelineMode
func (s *TokenState) Scan(value interface{}) error {
	*s = TokenState(mgmtPB.ApiToken_State_value[value.(string)])
	return nil
}

// Value function for custom GORM type PipelineMode
func (s TokenState) Value() (driver.Value, error) {
	return mgmtPB.ApiToken_State(s).String(), nil
}
