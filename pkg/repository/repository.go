package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/redis/go-redis/v9"
	"go.einride.tech/aip/filtering"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/x/paginate"

	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
)

type CtxKey string

const UserUIDCtxKey CtxKey = "userUID"

// Repository interface
type Repository interface {
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

	listOwners(ctx context.Context, ownerType string, pageSize int, pageToken string, filter filtering.Filter) ([]*datamodel.Owner, int64, string, error)
	createOwner(ctx context.Context, ownerType string, user *datamodel.Owner) error
	getOwner(ctx context.Context, ownerType string, id string, includeAvatar bool) (*datamodel.Owner, error)
	getOwnerByUID(ctx context.Context, ownerType string, uid uuid.UUID) (*datamodel.Owner, error)
	updateOwner(ctx context.Context, ownerType string, id string, user *datamodel.Owner) error
	deleteOwner(ctx context.Context, ownerType string, id string) error

	GetUserPasswordHash(ctx context.Context, uid uuid.UUID) (string, time.Time, error)
	UpdateUserPasswordHash(ctx context.Context, uid uuid.UUID, newPassword string, updateTime time.Time) error

	CreateToken(ctx context.Context, token *datamodel.Token) error
	ListTokens(ctx context.Context, owner string, pageSize int64, pageToken string) ([]*datamodel.Token, int64, string, error)
	GetToken(ctx context.Context, owner string, id string) (*datamodel.Token, error)
	DeleteToken(ctx context.Context, owner string, id string) error
	LookupToken(ctx context.Context, token string) (*datamodel.Token, error)

	ListAllValidTokens(ctx context.Context) ([]datamodel.Token, error)

	AddCredit(context.Context, datamodel.Credit) error
	GetRemainingCredit(ctx context.Context, ownerUID uuid.UUID) (float64, error)
	SubtractCredit(ctx context.Context, ownerUID uuid.UUID, amount float64) error
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

func (r *repository) checkPinnedUser(ctx context.Context, db *gorm.DB) *gorm.DB {
	userUID := ctx.Value(UserUIDCtxKey)
	// If the user is pinned, we will use the primary database for querying.
	if !errors.Is(r.redisClient.Get(ctx, fmt.Sprintf("db_pin_user:%s", userUID)).Err(), redis.Nil) {
		db = db.Clauses(dbresolver.Write)
	}
	return db
}

func (r *repository) pinUser(ctx context.Context) {
	userUID := ctx.Value(UserUIDCtxKey)
	// To solve the read-after-write inconsistency problem,
	// we will direct the user to read from the primary database for a certain time frame
	// to ensure that the data is synchronized from the primary DB to the replica DB.
	_ = r.redisClient.Set(ctx, fmt.Sprintf("db_pin_user:%s", userUID), time.Now(), time.Duration(config.Config.Database.Replica.ReplicationTimeFrame)*time.Second)
}

func (r *repository) ListUsers(ctx context.Context, pageSize int, pageToken string, filter filtering.Filter) ([]*datamodel.Owner, int64, string, error) {
	return r.listOwners(ctx, "user", pageSize, pageToken, filter)
}
func (r *repository) CreateUser(ctx context.Context, user *datamodel.Owner) error {
	return r.createOwner(ctx, "user", user)
}
func (r *repository) GetUser(ctx context.Context, id string, includeAvatar bool) (*datamodel.Owner, error) {
	return r.getOwner(ctx, "user", id, includeAvatar)
}
func (r *repository) GetUserByUID(ctx context.Context, uid uuid.UUID) (*datamodel.Owner, error) {
	return r.getOwnerByUID(ctx, "user", uid)
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
	return r.getOwner(ctx, "organization", id, includeAvatar)
}
func (r *repository) GetOrganizationByUID(ctx context.Context, uid uuid.UUID) (*datamodel.Owner, error) {
	return r.getOwnerByUID(ctx, "organization", uid)
}
func (r *repository) UpdateOrganization(ctx context.Context, id string, org *datamodel.Owner) error {
	return r.updateOwner(ctx, "organization", id, org)
}
func (r *repository) DeleteOrganization(ctx context.Context, id string) error {
	return r.deleteOwner(ctx, "organization", id)
}

func (r *repository) GetAllUsers(ctx context.Context) ([]*datamodel.Owner, error) {
	db := r.checkPinnedUser(ctx, r.db)
	logger, _ := logger.GetZapLogger(ctx)
	var users []*datamodel.Owner
	if result := db.Find(users).Where("owner_type = 'user'"); result.Error != nil {
		logger.Error(result.Error.Error())
		return nil, result.Error
	}
	return users, nil
}

func (r *repository) listOwners(ctx context.Context, ownerType string, pageSize int, pageToken string, filter filtering.Filter) ([]*datamodel.Owner, int64, string, error) {

	db := r.checkPinnedUser(ctx, r.db)

	logger, _ := logger.GetZapLogger(ctx)
	totalSize := int64(0)
	if result := db.Model(&datamodel.Owner{}).Where("owner_type = ?", ownerType).Count(&totalSize); result.Error != nil {
		logger.Error(result.Error.Error())
		return nil, totalSize, "", result.Error
	}

	queryBuilder := db.Model(&datamodel.Owner{}).Order("create_time DESC, id DESC")
	queryBuilder = queryBuilder.Where("owner_type = ?", ownerType)

	if pageSize > 0 {
		queryBuilder = queryBuilder.Limit(pageSize)
	}

	if pageToken != "" {
		// TODO: check pageToken in handler
		createTime, uid, err := paginate.DecodeToken(pageToken)
		if err != nil {
			return nil, totalSize, "", ErrPageTokenDecode
		}
		queryBuilder = queryBuilder.Where("(create_time,uid) < (?::timestamp, ?)", createTime, uid)
	}

	var owners []*datamodel.Owner
	var createTime time.Time

	rows, err := queryBuilder.Rows()
	if err != nil {
		logger.Error(err.Error())
		return nil, totalSize, "", err
	}
	defer rows.Close()
	for rows.Next() {
		var item datamodel.Owner
		if err = db.ScanRows(rows, &item); err != nil {
			logger.Error(err.Error())
			return nil, totalSize, "", err
		}
		createTime = item.CreateTime

		owners = append(owners, &item)
	}

	nextPageToken := ""
	if len(owners) > 0 {

		lastUID := (owners)[len(owners)-1].UID
		lastItem := &datamodel.Owner{}
		if result := db.Model(&datamodel.Owner{}).
			Omit("profile_avatar").
			Where("owner_type = ?", ownerType).
			Order("create_time ASC, uid ASC").
			Limit(1).Find(lastItem); result.Error != nil {
			return nil, 0, "", result.Error
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

	r.pinUser(ctx)
	db := r.checkPinnedUser(ctx, r.db)

	if ownerType != owner.OwnerType.String {
		return ErrOwnerTypeNotMatch
	}

	logger, _ := logger.GetZapLogger(ctx)
	if result := db.Model(&datamodel.Owner{}).Create(owner); result.Error != nil {
		logger.Error(result.Error.Error())
		return result.Error
	}
	return nil
}

func (r *repository) getOwner(ctx context.Context, ownerType string, id string, includeAvatar bool) (*datamodel.Owner, error) {

	db := r.checkPinnedUser(ctx, r.db)

	var owner datamodel.Owner
	queryBuilder := db.Model(&datamodel.Owner{}).Where("owner_type = ?", ownerType).Where("id = ?", id)
	if !includeAvatar {
		queryBuilder = queryBuilder.Omit("profile_avatar")
	}
	if result := queryBuilder.First(&owner); result.Error != nil {
		return nil, result.Error
	}
	return &owner, nil
}

func (r *repository) getOwnerByUID(ctx context.Context, ownerType string, uid uuid.UUID) (*datamodel.Owner, error) {

	db := r.checkPinnedUser(ctx, r.db)

	var owner datamodel.Owner
	if result := db.Model(&datamodel.Owner{}).Omit("profile_avatar").Where("owner_type = ?", ownerType).Where("uid = ?", uid.String()).First(&owner); result.Error != nil {
		return nil, result.Error
	}
	return &owner, nil
}

func (r *repository) updateOwner(ctx context.Context, ownerType string, id string, owner *datamodel.Owner) error {

	r.pinUser(ctx)
	db := r.checkPinnedUser(ctx, r.db)

	logger, _ := logger.GetZapLogger(ctx)
	if result := db.Select("*").Omit("UID").Omit("password_hash").Model(&datamodel.Owner{}).Where("owner_type = ?", ownerType).Where("id = ?", id).Updates(owner); result.Error != nil {
		logger.Error(result.Error.Error())
		return result.Error
	}
	return nil
}

func (r *repository) deleteOwner(ctx context.Context, ownerType string, id string) error {

	r.pinUser(ctx)
	db := r.checkPinnedUser(ctx, r.db)

	logger, _ := logger.GetZapLogger(ctx)
	result := db.Model(&datamodel.Owner{}).
		Where("owner_type = ?", ownerType).
		Where("id = ?", id).
		Delete(&datamodel.Owner{})

	if result.Error != nil {
		logger.Error(result.Error.Error())
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrNoDataDeleted
	}

	return nil
}

// GetUser gets a user by ID
// Return error types
//   - codes.NotFound
func (r *repository) GetUserPasswordHash(ctx context.Context, uid uuid.UUID) (string, time.Time, error) {

	db := r.checkPinnedUser(ctx, r.db)

	var pw datamodel.Password
	if result := db.First(&pw, "uid = ?", uid.String()); result.Error != nil {
		return "", time.Time{}, result.Error
	}
	return pw.PasswordHash.String, pw.PasswordUpdateTime, nil
}

func (r *repository) UpdateUserPasswordHash(ctx context.Context, uid uuid.UUID, newPassword string, updateTime time.Time) error {

	r.pinUser(ctx)
	db := r.checkPinnedUser(ctx, r.db)

	logger, _ := logger.GetZapLogger(ctx)
	if result := db.Select("*").Omit("UID").Model(&datamodel.Password{}).Where("uid = ?", uid.String()).Updates(datamodel.Password{
		PasswordHash:       sql.NullString{String: newPassword, Valid: true},
		PasswordUpdateTime: updateTime,
	}); result.Error != nil {
		logger.Error(result.Error.Error())
		return result.Error
	}
	return nil
}

// TODO: use general filter
func (r *repository) ListAllValidTokens(ctx context.Context) (tokens []datamodel.Token, err error) {

	db := r.checkPinnedUser(ctx, r.db)

	queryBuilder := db.Model(&datamodel.Token{}).Where("state = ?", datamodel.TokenState(mgmtPB.ApiToken_STATE_ACTIVE))
	queryBuilder.Where("expire_time >= ?", time.Now())
	rows, err := queryBuilder.Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item datamodel.Token
		if err = db.ScanRows(rows, &item); err != nil {
			return nil, err
		}
		// createTime = item.CreateTime
		tokens = append(tokens, item)
	}

	return tokens, nil
}

func (r *repository) ListTokens(ctx context.Context, owner string, pageSize int64, pageToken string) (tokens []*datamodel.Token, totalSize int64, nextPageToken string, err error) {

	db := r.checkPinnedUser(ctx, r.db)

	if result := db.Model(&datamodel.Token{}).Where("owner = ?", owner).Count(&totalSize); result.Error != nil {
		return nil, 0, "", err
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
			return nil, 0, "", err
		}
		queryBuilder = queryBuilder.Where("(create_time,uid) < (?::timestamp, ?)", createTime, uid)
	}

	var createTime time.Time
	rows, err := queryBuilder.Rows()
	if err != nil {
		return nil, 0, "", err
	}
	defer rows.Close()
	for rows.Next() {
		var item datamodel.Token
		if err = db.ScanRows(rows, &item); err != nil {
			return nil, 0, "", err
		}
		createTime = item.CreateTime
		tokens = append(tokens, &item)
	}

	if len(tokens) > 0 {
		lastUID := (tokens)[len(tokens)-1].UID
		lastItem := &datamodel.Token{}
		if result := db.Model(&datamodel.Token{}).
			Where("owner = ?", owner).
			Order("create_time ASC, uid ASC").
			Limit(1).Find(lastItem); result.Error != nil {
			return nil, 0, "", err
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

	r.pinUser(ctx)
	db := r.checkPinnedUser(ctx, r.db)

	logger, _ := logger.GetZapLogger(ctx)
	if result := db.Model(&datamodel.Token{}).Create(token); result.Error != nil {
		logger.Error(result.Error.Error())
		return result.Error
	}
	return nil
}

func (r *repository) GetToken(ctx context.Context, owner string, id string) (*datamodel.Token, error) {

	db := r.checkPinnedUser(ctx, r.db)

	queryBuilder := db.Model(&datamodel.Token{}).Where("id = ? AND owner = ?", id, owner)
	var token datamodel.Token
	if result := queryBuilder.First(&token); result.Error != nil {
		return nil, result.Error
	}
	return &token, nil
}

func (r *repository) LookupToken(ctx context.Context, accessToken string) (*datamodel.Token, error) {

	db := r.checkPinnedUser(ctx, r.db)

	queryBuilder := db.Model(&datamodel.Token{}).Where("access_token = ?", accessToken)
	var token datamodel.Token
	if result := queryBuilder.First(&token); result.Error != nil {
		return nil, result.Error
	}
	return &token, nil
}

func (r *repository) DeleteToken(ctx context.Context, owner string, id string) error {

	r.pinUser(ctx)
	db := r.checkPinnedUser(ctx, r.db)

	result := db.Model(&datamodel.Token{}).
		Where("id = ? AND owner = ?", id, owner).
		Delete(&datamodel.Token{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrNoDataDeleted
	}

	return nil
}

func (r *repository) AddCredit(ctx context.Context, credit datamodel.Credit) error {
	r.pinUser(ctx)
	db := r.checkPinnedUser(ctx, r.db)

	return db.Create(credit).Error
}

type remainingCredit struct {
	Total float64

	// ExpireTime will be used for subtraction, when we group by this column.
	ExpireTime sql.NullTime
}

// GetRemainingCredit is computed as the sum of entries of a owner that aren't
// expired.
func (r *repository) GetRemainingCredit(ctx context.Context, ownerUID uuid.UUID) (float64, error) {
	db := r.checkPinnedUser(ctx, r.db)

	var result remainingCredit
	q := db.Model(datamodel.Credit{}).Select("sum(amount) as total").
		Where("owner_uid = ?", ownerUID).
		Where("expire_time is null or expire_time > ?", time.Now()).
		Group("owner_uid")

	if err := q.First(&result).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, err
		}
		return 0, nil
	}

	return result.Total, nil
}

// ErrNotEnoughCredit will be returned when trying to subtract more credit than
// what's available. The owner's remaining credit will be set to zero.
var ErrNotEnoughCredit = fmt.Errorf("not enough credit")

// Because entries might expire (users get free monthly credit), when
// subtracting credit we need to:
// 1. Cancel out credit that has an expiration date. The negative entry will
// have an expiration date.
// 2. If there's remaining credit to be subtracted, cancel out non-expiring
// credit.
//
// That way, when computing the remaining credit we'll only take into account
// credit that doesn't expire or credit within the period.
// If the number of entries grows too big over time, this won't be as efficient
// as keeping a separate table with the balance. For now, this is simple than
// recomputing the balance when an entry expires.
func (r *repository) SubtractCredit(ctx context.Context, ownerUID uuid.UUID, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("only positive amounts are allowed")
	}

	r.pinUser(ctx)
	db := r.checkPinnedUser(ctx, r.db).WithContext(ctx)

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := acquireTxLock(ctx, tx, "credit:"+ownerUID.String()); err != nil {
			return fmt.Errorf("cannot acquire lock: %w", err)
		}

		q := tx.Model(datamodel.Credit{}).
			Select("sum(amount) as total", "expire_time").
			Where("owner_uid = ?", ownerUID).
			Where("expire_time is null or expire_time > ?", time.Now()).
			Group("expire_time")

		rows, err := q.Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		// Instead of entering a single negative amount, we cancel first the
		// credit entries that have an expiration date. The negative entries
		// will have, in this case, the same expiration date as the positive
		// one.
		entriesToCancel := []*datamodel.Credit{}
		for rows.Next() {
			remaining := new(remainingCredit)
			if err := tx.ScanRows(rows, remaining); err != nil {
				return err
			}

			entry := &datamodel.Credit{
				OwnerUID:   ownerUID,
				ExpireTime: remaining.ExpireTime,
			}
			entriesToCancel = append(entriesToCancel, entry)

			diff := remaining.Total - amount
			if diff >= 0 { // credit is enough.
				entry.Amount = -amount
				amount = 0

				rows.Close()
				break
			}

			// set credit to zero and continue subtracting the remaining amount.
			entry.Amount = -remaining.Total
			amount = -diff

		}

		for _, entry := range entriesToCancel {
			if err := tx.Create(entry).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	if amount > 0 {
		return ErrNotEnoughCredit
	}

	return nil
}

// Acquire an exclusive advisory lock for the provided key. This allows us to
// lock a resource without having an associated row in the database, e.g.
// locking the credit ledger of a user (where we fetch the records with a GROUP
// BY statement, which doesn't allow for locks).
func acquireTxLock(ctx context.Context, tx *gorm.DB, key string) error {
	hash := fnv.New64()
	if _, err := hash.Write([]byte(key)); err != nil {
		return err
	}
	hashedKey := int64(hash.Sum64())

	return tx.WithContext(ctx).Exec("SELECT pg_advisory_xact_lock(?)", hashedKey).Error
}
