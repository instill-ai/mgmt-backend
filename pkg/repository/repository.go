package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/redis/go-redis/v9"
	"go.einride.tech/aip/filtering"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/x/paginate"

	mgmtpb "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
	errorsx "github.com/instill-ai/x/errors"
)

const (
	// DefaultPageSize is the default pagination page size when page size is
	// not assigned
	DefaultPageSize = 100

	// MaxPageSize is the maximum pagination page size if the assigned value is
	// over this number
	MaxPageSize = 1000
)

type CtxKey string

// UserUIDCtxKey must be present in the context in order to pin a user to the
// primary database.
// TODO we shouldn't force parent packages to be aware of this context key. All
// authenticated requests and endpoints modifying a user resource should
// add this key (defined in a common package) to the context.
const UserUIDCtxKey CtxKey = "userUID"

// Repository interface
type Repository interface {
	PinUser(context.Context)
	CheckPinnedUser(context.Context, *gorm.DB) *gorm.DB

	GetAllUsers(ctx context.Context) ([]*datamodel.Owner, error)

	ListUsers(ctx context.Context, pageSize int, pageToken string, filter filtering.Filter) ([]*datamodel.Owner, int64, string, error)
	CreateUser(ctx context.Context, user *datamodel.Owner) error
	GetUser(ctx context.Context, id string, includeAvatar bool) (*datamodel.Owner, error)
	GetUserByUID(ctx context.Context, uid uuid.UUID) (*datamodel.Owner, error)
	UpdateUser(ctx context.Context, id string, user *datamodel.Owner) error
	DeleteUser(ctx context.Context, id string) error

	ListOrganizations(ctx context.Context, pageSize int, pageToken string, filter filtering.Filter) ([]*datamodel.Owner, int64, string, error)
	CreateOrganization(ctx context.Context, user *datamodel.Owner) error
	GetOrganization(ctx context.Context, id string, includeAvatar bool) (*datamodel.Owner, error)
	GetOrganizationByUID(ctx context.Context, uid uuid.UUID) (*datamodel.Owner, error)
	UpdateOrganization(ctx context.Context, id string, user *datamodel.Owner) error
	DeleteOrganization(ctx context.Context, id string) error

	GetOwner(ctx context.Context, id string, includeAvatar bool) (*datamodel.Owner, error)
	GetOwnerByUID(ctx context.Context, uid uuid.UUID) (*datamodel.Owner, error)

	listOwners(ctx context.Context, ownerType string, pageSize int, pageToken string, filter filtering.Filter) ([]*datamodel.Owner, int64, string, error)
	createOwner(ctx context.Context, ownerType string, user *datamodel.Owner) error
	updateOwner(ctx context.Context, ownerType string, id string, user *datamodel.Owner) error
	deleteOwner(ctx context.Context, ownerType string, id string) error

	GetUserPasswordHash(ctx context.Context, uid uuid.UUID) (string, time.Time, error)
	UpdateUserPasswordHash(ctx context.Context, uid uuid.UUID, newPassword string, updateTime time.Time) error

	CreateToken(ctx context.Context, token *datamodel.Token) error
	ListTokens(ctx context.Context, owner string, pageSize int64, pageToken string) ([]*datamodel.Token, int64, string, error)
	GetToken(ctx context.Context, owner string, id string) (*datamodel.Token, error)
	DeleteToken(ctx context.Context, owner string, id string) error
	LookupToken(ctx context.Context, token string) (*datamodel.Token, error)
	UpdateTokenLastUseTime(ctx context.Context, accessToken string) error

	ListAllValidTokens(ctx context.Context) ([]datamodel.Token, error)
}

type repository struct {
	db          *gorm.DB
	redisClient *redis.Client
}

// NewRepository initiates a repository instance
func NewRepository(db *gorm.DB, redisClient *redis.Client) Repository {
	return &repository{
		db:          db,
		redisClient: redisClient,
	}
}

// CheckPinnedUser uses the primary database for querying if the user is
// pinned. This is used to solve read-after-write inconsistency problems on
// multi-region setups.
func (r *repository) CheckPinnedUser(ctx context.Context, db *gorm.DB) *gorm.DB {
	userUID := ctx.Value(UserUIDCtxKey)
	if !errors.Is(r.redisClient.Get(ctx, fmt.Sprintf("db_pin_user:%s", userUID)).Err(), redis.Nil) {
		db = db.Clauses(dbresolver.Write)
	}
	return db.WithContext(ctx)
}

// PinUser sets the primary database as the read of the user for a certain
// period to ensure that the data is synchronized from the primary DB to the
// replica DB.
func (r *repository) PinUser(ctx context.Context) {
	userUID := ctx.Value(UserUIDCtxKey)
	_ = r.redisClient.Set(ctx, fmt.Sprintf("db_pin_user:%s", userUID), time.Now(), time.Duration(config.Config.Database.Replica.ReplicationTimeFrame)*time.Second)
}

func (r *repository) ListUsers(ctx context.Context, pageSize int, pageToken string, filter filtering.Filter) ([]*datamodel.Owner, int64, string, error) {
	return r.listOwners(ctx, "user", pageSize, pageToken, filter)
}
func (r *repository) CreateUser(ctx context.Context, user *datamodel.Owner) error {
	return r.createOwner(ctx, "user", user)
}

// ownerWithType is used to wrap an owner fetch with a tyupe check.
func ownerWithType(o *datamodel.Owner, ownerType string) (*datamodel.Owner, error) {
	if !o.OwnerType.Valid || o.OwnerType.String != ownerType {
		return nil, errorsx.RepositoryErr(gorm.ErrRecordNotFound)
	}

	return o, nil
}

func (r *repository) GetUser(ctx context.Context, id string, includeAvatar bool) (*datamodel.Owner, error) {
	owner, err := r.GetOwner(ctx, id, includeAvatar)
	if err != nil {
		return nil, err
	}

	return ownerWithType(owner, "user")
}

func (r *repository) GetUserByUID(ctx context.Context, uid uuid.UUID) (*datamodel.Owner, error) {
	owner, err := r.GetOwnerByUID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return ownerWithType(owner, "user")
}

func (r *repository) UpdateUser(ctx context.Context, id string, user *datamodel.Owner) error {
	return r.updateOwner(ctx, "user", id, user)
}

func (r *repository) DeleteUser(ctx context.Context, id string) error {
	return r.deleteOwner(ctx, "user", id)
}

func (r *repository) ListOrganizations(ctx context.Context, pageSize int, pageToken string, filter filtering.Filter) ([]*datamodel.Owner, int64, string, error) {
	return r.listOwners(ctx, "organization", pageSize, pageToken, filter)
}
func (r *repository) CreateOrganization(ctx context.Context, org *datamodel.Owner) error {
	return r.createOwner(ctx, "organization", org)
}

func (r *repository) GetOrganization(ctx context.Context, id string, includeAvatar bool) (*datamodel.Owner, error) {
	owner, err := r.GetOwner(ctx, id, includeAvatar)
	if err != nil {
		return nil, err
	}

	return ownerWithType(owner, "organization")
}

func (r *repository) GetOrganizationByUID(ctx context.Context, uid uuid.UUID) (*datamodel.Owner, error) {
	owner, err := r.GetOwnerByUID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return ownerWithType(owner, "organization")
}

func (r *repository) UpdateOrganization(ctx context.Context, id string, org *datamodel.Owner) error {
	return r.updateOwner(ctx, "organization", id, org)
}
func (r *repository) DeleteOrganization(ctx context.Context, id string) error {
	return r.deleteOwner(ctx, "organization", id)
}

func (r *repository) GetAllUsers(ctx context.Context) ([]*datamodel.Owner, error) {
	db := r.CheckPinnedUser(ctx, r.db)
	var users []*datamodel.Owner
	if err := db.Find(users).Where("owner_type = 'user'").Error; err != nil {
		return nil, errorsx.RepositoryErr(fmt.Errorf("getting all users: %w", err))
	}
	return users, nil
}

func (r *repository) listOwners(ctx context.Context, ownerType string, pageSize int, pageToken string, filter filtering.Filter) ([]*datamodel.Owner, int64, string, error) {

	db := r.CheckPinnedUser(ctx, r.db)

	totalSize := int64(0)
	if err := db.Model(&datamodel.Owner{}).Where("owner_type = ?", ownerType).Count(&totalSize).Error; err != nil {
		return nil, totalSize, "", errorsx.RepositoryErr(fmt.Errorf("counting owners: %w", err))
	}

	queryBuilder := db.Model(&datamodel.Owner{}).Order("create_time DESC, id DESC")
	queryBuilder = queryBuilder.Where("owner_type = ?", ownerType)

	var expr *clause.Expr
	var err error
	if expr, err = r.transpileFilter(filter, "owner"); err != nil {
		return nil, 0, "", err
	}
	if expr != nil {
		queryBuilder = queryBuilder.Where(expr)
	}

	if pageSize > 0 {
		queryBuilder = queryBuilder.Limit(pageSize)
	}

	if pageToken != "" {
		// TODO: check pageToken in handler
		createTime, uid, err := paginate.DecodeToken(pageToken)
		if err != nil {
			return nil, totalSize, "", errorsx.NewPageTokenErr(err)
		}
		queryBuilder = queryBuilder.Where("(create_time,uid) < (?::timestamp, ?)", createTime, uid)
	}

	var owners []*datamodel.Owner
	var createTime time.Time

	rows, err := queryBuilder.Rows()
	if err != nil {
		return nil, totalSize, "", errorsx.RepositoryErr(fmt.Errorf("querying owners: %w", err))
	}
	defer rows.Close()
	for rows.Next() {
		var item datamodel.Owner
		if err = db.ScanRows(rows, &item); err != nil {
			return nil, totalSize, "", errorsx.RepositoryErr(fmt.Errorf("scanning owner row: %w", err))
		}
		createTime = item.CreateTime

		owners = append(owners, &item)
	}

	nextPageToken := ""
	if len(owners) > 0 {

		lastUID := (owners)[len(owners)-1].UID
		lastItem := &datamodel.Owner{}
		queryBuilder := db.Model(&datamodel.Owner{}).
			Omit("profile_avatar").
			Where("owner_type = ?", ownerType).
			Order("create_time ASC, uid ASC")
		var expr *clause.Expr
		var err error
		if expr, err = r.transpileFilter(filter, "owner"); err != nil {
			return nil, 0, "", err
		}
		if expr != nil {
			queryBuilder = queryBuilder.Where(expr)
		}
		if err := queryBuilder.Limit(1).Find(lastItem).Error; err != nil {
			return nil, 0, "", errorsx.RepositoryErr(fmt.Errorf("finding last owner for pagination: %w", err))
		}
		if lastItem.UID.String() == lastUID.String() {
			nextPageToken = ""
		} else {
			nextPageToken = paginate.EncodeToken(createTime, lastUID.String())
		}

		return owners, totalSize, nextPageToken, nil
	}

	return owners, totalSize, "", nil
}

func (r *repository) createOwner(ctx context.Context, ownerType string, owner *datamodel.Owner) error {
	r.PinUser(ctx)
	db := r.CheckPinnedUser(ctx, r.db)

	if ownerType != owner.OwnerType.String {
		return errorsx.ErrOwnerTypeNotMatch
	}

	if err := db.Model(&datamodel.Owner{}).Create(owner).Error; err != nil {
		return errorsx.RepositoryErr(fmt.Errorf("inserting owner: %w", err))
	}

	return nil
}

func (r *repository) GetOwner(ctx context.Context, id string, includeAvatar bool) (*datamodel.Owner, error) {
	db := r.CheckPinnedUser(ctx, r.db)

	// Temporarily disable logging for admin ID queries
	if id == "admin" {
		db = db.Session(&gorm.Session{Logger: db.Logger.LogMode(logger.Silent)})
	}

	var owner datamodel.Owner
	queryBuilder := db.Model(&datamodel.Owner{}).Where("id = ?", id)
	if !includeAvatar {
		queryBuilder = queryBuilder.Omit("profile_avatar")
	}
	if err := queryBuilder.First(&owner).Error; err != nil {
		return nil, errorsx.RepositoryErr(fmt.Errorf("getting owner by id: %w", err))
	}
	return &owner, nil
}

func (r *repository) GetOwnerByUID(ctx context.Context, uid uuid.UUID) (*datamodel.Owner, error) {
	db := r.CheckPinnedUser(ctx, r.db)

	var owner datamodel.Owner
	if err := db.Model(&datamodel.Owner{}).
		Omit("profile_avatar").
		Where("uid = ?", uid.String()).
		First(&owner).
		Error; err != nil {

		return nil, errorsx.RepositoryErr(fmt.Errorf("getting owner by uid: %w", err))
	}
	return &owner, nil
}

func (r *repository) updateOwner(ctx context.Context, ownerType string, id string, owner *datamodel.Owner) error {

	r.PinUser(ctx)
	db := r.CheckPinnedUser(ctx, r.db)

	if err := db.Omit("UID").
		Omit("password_hash").
		Model(&datamodel.Owner{}).
		Where("owner_type = ?", ownerType).
		Where("id = ?", id).
		Updates(owner).
		Error; err != nil {

		return errorsx.RepositoryErr(fmt.Errorf("updating owner: %w", err))
	}
	return nil
}

func (r *repository) deleteOwner(ctx context.Context, ownerType string, id string) error {

	r.PinUser(ctx)
	db := r.CheckPinnedUser(ctx, r.db)

	result := db.Model(&datamodel.Owner{}).
		Where("owner_type = ?", ownerType).
		Where("id = ?", id).
		Delete(&datamodel.Owner{})

	if result.Error != nil {
		return errorsx.RepositoryErr(fmt.Errorf("deleting owner: %w", result.Error))
	}

	if result.RowsAffected == 0 {
		return errorsx.ErrNoDataDeleted
	}

	return nil
}

// GetUser gets a user by ID
// Return error types
//   - codes.NotFound
func (r *repository) GetUserPasswordHash(ctx context.Context, uid uuid.UUID) (string, time.Time, error) {

	db := r.CheckPinnedUser(ctx, r.db)

	var pw datamodel.Password
	if err := db.First(&pw, "uid = ?", uid.String()).Error; err != nil {
		return "", time.Time{}, errorsx.RepositoryErr(fmt.Errorf("getting password hash: %w", err))
	}
	return pw.PasswordHash.String, pw.PasswordUpdateTime, nil
}

func (r *repository) UpdateUserPasswordHash(ctx context.Context, uid uuid.UUID, newPassword string, updateTime time.Time) error {

	r.PinUser(ctx)
	db := r.CheckPinnedUser(ctx, r.db)

	if err := db.Select("*").
		Omit("UID").
		Model(&datamodel.Password{}).
		Where("uid = ?", uid.String()).
		Updates(datamodel.Password{
			PasswordHash:       sql.NullString{String: newPassword, Valid: true},
			PasswordUpdateTime: updateTime,
		}).Error; err != nil {

		return errorsx.RepositoryErr(fmt.Errorf("updating password hash: %w", err))
	}

	return nil
}

// TODO: use general filter
func (r *repository) ListAllValidTokens(ctx context.Context) (tokens []datamodel.Token, err error) {

	db := r.CheckPinnedUser(ctx, r.db)

	queryBuilder := db.Model(&datamodel.Token{}).Where("state = ?", datamodel.TokenState(mgmtpb.ApiToken_STATE_ACTIVE))
	queryBuilder.Where("expire_time >= ?", time.Now())
	rows, err := queryBuilder.Rows()
	if err != nil {
		return nil, errorsx.RepositoryErr(fmt.Errorf("querying valid tokens: %w", err))
	}
	defer rows.Close()
	for rows.Next() {
		var item datamodel.Token
		if err = db.ScanRows(rows, &item); err != nil {
			return nil, errorsx.RepositoryErr(fmt.Errorf("scanning token row: %w", err))
		}
		// createTime = item.CreateTime
		tokens = append(tokens, item)
	}

	return tokens, nil
}

func (r *repository) ListTokens(ctx context.Context, owner string, pageSize int64, pageToken string) (tokens []*datamodel.Token, totalSize int64, nextPageToken string, err error) {

	db := r.CheckPinnedUser(ctx, r.db)

	if err := db.Model(&datamodel.Token{}).Where("owner = ?", owner).Count(&totalSize).Error; err != nil {
		return nil, 0, "", errorsx.RepositoryErr(fmt.Errorf("counting tokens: %w", err))
	}

	queryBuilder := db.Model(&datamodel.Token{}).Order("create_time DESC, uid DESC").Where("owner = ?", owner)

	if pageSize == 0 {
		pageSize = DefaultPageSize
	} else if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	queryBuilder = queryBuilder.Limit(int(pageSize))

	if pageToken != "" {
		createTime, uid, err := paginate.DecodeToken(pageToken)
		if err != nil {
			return nil, 0, "", errorsx.NewPageTokenErr(err)
		}
		queryBuilder = queryBuilder.Where("(create_time,uid) < (?::timestamp, ?)", createTime, uid)
	}

	var createTime time.Time
	rows, err := queryBuilder.Rows()
	if err != nil {
		return nil, 0, "", errorsx.RepositoryErr(fmt.Errorf("querying tokens: %w", err))
	}
	defer rows.Close()
	for rows.Next() {
		var item datamodel.Token
		if err = db.ScanRows(rows, &item); err != nil {
			return nil, 0, "", errorsx.RepositoryErr(fmt.Errorf("scanning token row: %w", err))
		}
		createTime = item.CreateTime
		tokens = append(tokens, &item)
	}

	if len(tokens) > 0 {
		lastUID := (tokens)[len(tokens)-1].UID
		lastItem := &datamodel.Token{}
		if err := db.Model(&datamodel.Token{}).
			Where("owner = ?", owner).
			Order("create_time ASC, uid ASC").
			Limit(1).Find(lastItem).Error; err != nil {

			return nil, 0, "", errorsx.RepositoryErr(fmt.Errorf("finding last token for pagination: %w", err))
		}
		if lastItem.UID.String() == lastUID.String() {
			nextPageToken = ""
		} else {
			nextPageToken = paginate.EncodeToken(createTime, lastUID.String())
		}
	}

	return tokens, totalSize, nextPageToken, nil
}

func (r *repository) CreateToken(ctx context.Context, token *datamodel.Token) error {

	r.PinUser(ctx)
	db := r.CheckPinnedUser(ctx, r.db)

	if err := db.Model(&datamodel.Token{}).Create(token).Error; err != nil {
		return errorsx.RepositoryErr(fmt.Errorf("creating token: %w", err))
	}
	return nil
}

func (r *repository) GetToken(ctx context.Context, owner string, id string) (*datamodel.Token, error) {

	db := r.CheckPinnedUser(ctx, r.db)

	queryBuilder := db.Model(&datamodel.Token{}).Where("id = ? AND owner = ?", id, owner)
	var token datamodel.Token
	if err := queryBuilder.First(&token).Error; err != nil {
		return nil, errorsx.RepositoryErr(fmt.Errorf("getting token: %w", err))
	}
	return &token, nil
}

func (r *repository) LookupToken(ctx context.Context, accessToken string) (*datamodel.Token, error) {

	db := r.CheckPinnedUser(ctx, r.db)

	queryBuilder := db.Model(&datamodel.Token{}).Where("access_token = ?", accessToken)
	var token datamodel.Token
	if err := queryBuilder.First(&token).Error; err != nil {
		return nil, errorsx.RepositoryErr(fmt.Errorf("looking up token: %w", err))
	}
	return &token, nil
}

func (r *repository) DeleteToken(ctx context.Context, owner string, id string) error {

	r.PinUser(ctx)
	db := r.CheckPinnedUser(ctx, r.db)

	result := db.Model(&datamodel.Token{}).
		Where("id = ? AND owner = ?", id, owner).
		Delete(&datamodel.Token{})

	if result.Error != nil {
		return errorsx.RepositoryErr(fmt.Errorf("deleting token: %w", result.Error))
	}

	if result.RowsAffected == 0 {
		return errorsx.ErrNoDataDeleted
	}

	return nil
}

func (r *repository) UpdateTokenLastUseTime(ctx context.Context, accessToken string) error {
	r.PinUser(ctx)
	db := r.CheckPinnedUser(ctx, r.db)

	result := db.Model(&datamodel.Token{}).
		Where("access_token = ?", accessToken).
		Update("last_use_time", time.Now())

	if result.Error != nil {
		return errorsx.RepositoryErr(fmt.Errorf("updating token last use time: %w", result.Error))
	}

	if result.RowsAffected == 0 {
		return errorsx.ErrNoDataUpdated
	}

	return nil
}

// TranspileFilter transpiles a parsed AIP filter expression to GORM DB clauses
func (r *repository) transpileFilter(filter filtering.Filter, tableName string) (*clause.Expr, error) {
	return (&Transpiler{
		filter:    filter,
		tableName: tableName,
	}).Transpile()
}
