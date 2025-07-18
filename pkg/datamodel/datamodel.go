package datamodel

import (
	"database/sql"
	"database/sql/driver"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	mgmtpb "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
)

type TokenState int32

const (
	// State: UNSPECIFIED
	StateUnspecified TokenState = 0
	// State: INACTIVE
	StateInactive TokenState = 1
	// State: ACTIVE
	StateActive TokenState = 2

	// In db, we should use current_time < expire_tiem to determine expire or not
	// We'll convert current_time > expire_tiem to pb's STATE_EXPIRED
	// This can be improved when the sync between db, redis are more robust
	// State: EXPIRED
	// STATE_EXPIRED TokenState = 3
)

type OnboardingStatus mgmtpb.OnboardingStatus

const (
	OnboardingStatusUnspecified OnboardingStatus = 0
	OnboardingStatusInProgress  OnboardingStatus = 1
	OnboardingStatusCompleted   OnboardingStatus = 2
)

// Base contains common columns for all tables
type Base struct {
	UID        uuid.UUID `gorm:"type:uuid;primary_key;<-:create"` // allow read and create, but not update
	CreateTime time.Time `gorm:"autoCreateTime:nano;<-:create"`   // allow read and create, but not update
	UpdateTime time.Time `gorm:"autoUpdateTime:nano"`
	// TODO: support DeleteTime gorm.DeletedAt `sql:"index"`
}

// User defines a user instance in the database
type Owner struct {
	Base
	ID                     string `gorm:"unique;not null;"`
	OwnerType              sql.NullString
	Email                  string `gorm:"unique;not null;"`
	CustomerID             string
	DisplayName            sql.NullString
	CompanyName            sql.NullString
	PublicEmail            sql.NullString
	Bio                    sql.NullString
	Role                   sql.NullString
	NewsletterSubscription bool `gorm:"default:false"`
	CookieToken            sql.NullString
	ProfileAvatar          sql.NullString
	SocialProfileLinks     datatypes.JSON `gorm:"type:jsonb"`
	OnboardingStatus       OnboardingStatus
}

type Password struct {
	Base
	PasswordHash       sql.NullString
	PasswordUpdateTime time.Time
}

func (Password) TableName() string {
	return "owner"
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
	*s = TokenState(mgmtpb.ApiToken_State_value[value.(string)])
	return nil
}

// Value function for custom GORM type PipelineMode
func (s TokenState) Value() (driver.Value, error) {
	return mgmtpb.ApiToken_State(s).String(), nil
}

// Scan function for custom GORM type OnboardingStatus
func (o *OnboardingStatus) Scan(value interface{}) error {
	*o = OnboardingStatus(mgmtpb.OnboardingStatus_value[value.(string)])
	return nil
}

// Value function for custom GORM type OnboardingStatus
func (o OnboardingStatus) Value() (driver.Value, error) {
	return mgmtpb.OnboardingStatus(o).String(), nil
}

type FGAMigration struct {
	StoreID              string `gorm:"column:store_id"`
	AuthorizationModelID string `gorm:"column:authorization_model_id"`
	MD5Hash              string `gorm:"column:md5_hash"`
}

func (FGAMigration) TableName() string {
	return "fga_migrations"
}
