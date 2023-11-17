package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"go.einride.tech/aip/filtering"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/x/paginate"

	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1alpha"
)

// Repository interface
type Repository interface {
	GetAllUsers(ctx context.Context) ([]*datamodel.Owner, error)

	ListUsers(ctx context.Context, pageSize int, pageToken string, filter filtering.Filter) ([]*datamodel.Owner, int64, string, error)
	CreateUser(ctx context.Context, user *datamodel.Owner) error
	GetUser(ctx context.Context, id string) (*datamodel.Owner, error)
	GetUserByUID(ctx context.Context, uid uuid.UUID) (*datamodel.Owner, error)
	UpdateUser(ctx context.Context, id string, user *datamodel.Owner) error
	DeleteUser(ctx context.Context, id string) error

	ListOrganizations(ctx context.Context, pageSize int, pageToken string, filter filtering.Filter) ([]*datamodel.Owner, int64, string, error)
	CreateOrganization(ctx context.Context, user *datamodel.Owner) error
	GetOrganization(ctx context.Context, id string) (*datamodel.Owner, error)
	GetOrganizationByUID(ctx context.Context, uid uuid.UUID) (*datamodel.Owner, error)
	UpdateOrganization(ctx context.Context, id string, user *datamodel.Owner) error
	DeleteOrganization(ctx context.Context, id string) error

	listOwners(ctx context.Context, ownerType string, pageSize int, pageToken string, filter filtering.Filter) ([]*datamodel.Owner, int64, string, error)
	createOwner(ctx context.Context, ownerType string, user *datamodel.Owner) error
	getOwner(ctx context.Context, ownerType string, id string) (*datamodel.Owner, error)
	getOwnerByUID(ctx context.Context, ownerType string, uid uuid.UUID) (*datamodel.Owner, error)
	updateOwner(ctx context.Context, ownerType string, id string, user *datamodel.Owner) error
	deleteOwner(ctx context.Context, ownerType string, id string) error

	GetUserPasswordHash(ctx context.Context, uid uuid.UUID) (string, time.Time, error)
	UpdateUserPasswordHash(ctx context.Context, uid uuid.UUID, newPassword string, updateTime time.Time) error

	CreateToken(ctx context.Context, token *datamodel.Token) error
	ListTokens(ctx context.Context, owner string, pageSize int64, pageToken string) ([]*datamodel.Token, int64, string, error)
	GetToken(ctx context.Context, owner string, id string) (*datamodel.Token, error)
	DeleteToken(ctx context.Context, owner string, id string) error

	ListAllValidTokens(ctx context.Context) ([]datamodel.Token, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository initiates a repository instance
func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) ListUsers(ctx context.Context, pageSize int, pageToken string, filter filtering.Filter) ([]*datamodel.Owner, int64, string, error) {
	return r.listOwners(ctx, "user", pageSize, pageToken, filter)
}
func (r *repository) CreateUser(ctx context.Context, user *datamodel.Owner) error {
	return r.createOwner(ctx, "user", user)
}
func (r *repository) GetUser(ctx context.Context, id string) (*datamodel.Owner, error) {
	return r.getOwner(ctx, "user", id)
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
func (r *repository) GetOrganization(ctx context.Context, id string) (*datamodel.Owner, error) {
	return r.getOwner(ctx, "organization", id)
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
	logger, _ := logger.GetZapLogger(ctx)
	var users []*datamodel.Owner
	if result := r.db.Find(users).Where("owner_type = 'user'"); result.Error != nil {
		logger.Error(result.Error.Error())
		return users, status.Errorf(codes.Internal, "error %v", result.Error)
	}
	return users, nil
}

func (r *repository) listOwners(ctx context.Context, ownerType string, pageSize int, pageToken string, filter filtering.Filter) ([]*datamodel.Owner, int64, string, error) {

	logger, _ := logger.GetZapLogger(ctx)
	totalSize := int64(0)
	if result := r.db.Model(&datamodel.Owner{}).Where("owner_type = ?", ownerType).Count(&totalSize); result.Error != nil {
		logger.Error(result.Error.Error())
		return nil, totalSize, "", status.Errorf(codes.Internal, "error %v", result.Error)
	}

	queryBuilder := r.db.Model(&datamodel.Owner{}).Order("create_time DESC, id DESC")
	queryBuilder = queryBuilder.Where("owner_type = ?", ownerType)

	if pageSize > 0 {
		queryBuilder = queryBuilder.Limit(pageSize)
	}

	if pageToken != "" {
		createTime, uid, err := paginate.DecodeToken(pageToken)
		if err != nil {
			return nil, totalSize, "", status.Errorf(codes.InvalidArgument, "Invalid page token: %s", err.Error())
		}
		queryBuilder = queryBuilder.Where("(create_time,uid) < (?::timestamp, ?)", createTime, uid)
	}

	var owners []*datamodel.Owner
	var createTime time.Time

	rows, err := queryBuilder.Rows()
	if err != nil {
		logger.Error(err.Error())
		return nil, totalSize, "", status.Errorf(codes.Internal, "error %v", err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var item datamodel.Owner
		if err = r.db.ScanRows(rows, &item); err != nil {
			logger.Error(err.Error())
			return nil, totalSize, "", status.Errorf(codes.Internal, "error %v", err.Error())
		}
		createTime = item.CreateTime

		owners = append(owners, &item)
	}

	nextPageToken := ""
	if len(owners) > 0 {

		lastUID := (owners)[len(owners)-1].UID
		lastItem := &datamodel.Owner{}
		if result := r.db.Model(&datamodel.Owner{}).
			Where("owner_type = ?", ownerType).
			Order("create_time ASC, uid ASC").
			Limit(1).Find(lastItem); result.Error != nil {
			return nil, 0, "", status.Errorf(codes.Internal, result.Error.Error())
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

	if ownerType != owner.OwnerType.String {
		return fmt.Errorf("wrong ownerType")
	}

	logger, _ := logger.GetZapLogger(ctx)
	if result := r.db.Model(&datamodel.Owner{}).Create(owner); result.Error != nil {
		var pgErr *pgconn.PgError
		if errors.As(result.Error, &pgErr) {
			if pgErr.Code == "23505" {
				return status.Errorf(codes.AlreadyExists, pgErr.Message)
			}
		}
		logger.Error(result.Error.Error())
		return status.Errorf(codes.Internal, "error %v", result.Error)
	}
	return nil
}

func (r *repository) getOwner(ctx context.Context, ownerType string, id string) (*datamodel.Owner, error) {
	var owner datamodel.Owner
	if result := r.db.Model(&datamodel.Owner{}).Where("owner_type = ?", ownerType).Where("id = ?", id).First(&owner); result.Error != nil {
		return nil, status.Error(codes.NotFound, "the owner is not found")
	}
	return &owner, nil
}

func (r *repository) getOwnerByUID(ctx context.Context, ownerType string, uid uuid.UUID) (*datamodel.Owner, error) {
	var owner datamodel.Owner
	if result := r.db.Model(&datamodel.Owner{}).Where("owner_type = ?", ownerType).Where("uid = ?", uid.String()).First(&owner); result.Error != nil {
		return nil, status.Error(codes.NotFound, "the owner is not found")
	}
	return &owner, nil
}

func (r *repository) updateOwner(ctx context.Context, ownerType string, id string, owner *datamodel.Owner) error {
	logger, _ := logger.GetZapLogger(ctx)
	if result := r.db.Select("*").Omit("UID").Omit("password_hash").Model(&datamodel.Owner{}).Where("owner_type = ?", ownerType).Where("id = ?", id).Updates(owner); result.Error != nil {
		logger.Error(result.Error.Error())
		return status.Errorf(codes.Internal, "error %v", result.Error)
	}
	return nil
}

func (r *repository) deleteOwner(ctx context.Context, ownerType string, id string) error {
	logger, _ := logger.GetZapLogger(ctx)
	result := r.db.Model(&datamodel.Owner{}).
		Where("owner_type = ?", ownerType).
		Where("id = ?", id).
		Delete(&datamodel.Owner{})

	if result.Error != nil {
		logger.Error(result.Error.Error())
		return status.Errorf(codes.Internal, "error %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return status.Errorf(codes.NotFound, "the owner with id %s is not found", id)
	}

	return nil
}

// GetUser gets a user by ID
// Return error types
//   - codes.NotFound
func (r *repository) GetUserPasswordHash(ctx context.Context, uid uuid.UUID) (string, time.Time, error) {
	var pw datamodel.Password
	if result := r.db.First(&pw, "uid = ?", uid.String()); result.Error != nil {
		return "", time.Time{}, status.Error(codes.NotFound, "the user is not found")
	}
	return pw.PasswordHash.String, pw.PasswordUpdateTime, nil
}

func (r *repository) UpdateUserPasswordHash(ctx context.Context, uid uuid.UUID, newPassword string, updateTime time.Time) error {
	logger, _ := logger.GetZapLogger(ctx)
	if result := r.db.Select("*").Omit("UID").Model(&datamodel.Password{}).Where("uid = ?", uid.String()).Updates(datamodel.Password{
		PasswordHash:       sql.NullString{String: newPassword, Valid: true},
		PasswordUpdateTime: updateTime,
	}); result.Error != nil {
		logger.Error(result.Error.Error())
		return status.Errorf(codes.Internal, "error %v", result.Error)
	}
	return nil
}

// TODO: use general filter
func (r *repository) ListAllValidTokens(ctx context.Context) (tokens []datamodel.Token, err error) {

	queryBuilder := r.db.Model(&datamodel.Token{}).Where("state = ?", datamodel.TokenState(mgmtPB.ApiToken_STATE_ACTIVE))
	queryBuilder.Where("expire_time >= ?", time.Now())
	rows, err := queryBuilder.Rows()
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var item datamodel.Token
		if err = r.db.ScanRows(rows, &item); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		// createTime = item.CreateTime
		tokens = append(tokens, item)
	}

	return tokens, nil
}

func (r *repository) ListTokens(ctx context.Context, owner string, pageSize int64, pageToken string) (tokens []*datamodel.Token, totalSize int64, nextPageToken string, err error) {

	if result := r.db.Model(&datamodel.Token{}).Where("owner = ?", owner).Count(&totalSize); result.Error != nil {
		return nil, 0, "", status.Errorf(codes.Internal, result.Error.Error())
	}

	queryBuilder := r.db.Model(&datamodel.Token{}).Order("create_time DESC, uid DESC").Where("owner = ?", owner)

	if pageSize == 0 {
		pageSize = DefaultPageSize
	} else if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	queryBuilder = queryBuilder.Limit(int(pageSize))

	if pageToken != "" {
		createTime, uid, err := paginate.DecodeToken(pageToken)
		if err != nil {
			return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid page token: %s", err.Error())
		}
		queryBuilder = queryBuilder.Where("(create_time,uid) < (?::timestamp, ?)", createTime, uid)
	}

	var createTime time.Time
	rows, err := queryBuilder.Rows()
	if err != nil {
		return nil, 0, "", status.Errorf(codes.Internal, err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var item datamodel.Token
		if err = r.db.ScanRows(rows, &item); err != nil {
			return nil, 0, "", status.Error(codes.Internal, err.Error())
		}
		createTime = item.CreateTime
		tokens = append(tokens, &item)
	}

	if len(tokens) > 0 {
		lastUID := (tokens)[len(tokens)-1].UID
		lastItem := &datamodel.Token{}
		if result := r.db.Model(&datamodel.Token{}).
			Where("owner = ?", owner).
			Order("create_time ASC, uid ASC").
			Limit(1).Find(lastItem); result.Error != nil {
			return nil, 0, "", status.Errorf(codes.Internal, result.Error.Error())
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
	logger, _ := logger.GetZapLogger(ctx)
	if result := r.db.Model(&datamodel.Token{}).Create(token); result.Error != nil {
		var pgErr *pgconn.PgError
		if errors.As(result.Error, &pgErr) {
			if pgErr.Code == "23505" {
				return status.Errorf(codes.AlreadyExists, pgErr.Message)
			}
		}
		logger.Error(result.Error.Error())
		return status.Errorf(codes.Internal, "error %v", result.Error)
	}
	return nil
}

func (r *repository) GetToken(ctx context.Context, owner string, id string) (*datamodel.Token, error) {
	queryBuilder := r.db.Model(&datamodel.Token{}).Where("id = ? AND owner = ?", id, owner)
	var token datamodel.Token
	if result := queryBuilder.First(&token); result.Error != nil {
		return nil, status.Errorf(codes.NotFound, "[GetToken] The token id %s you specified is not found", id)
	}
	return &token, nil
}

func (r *repository) DeleteToken(ctx context.Context, owner string, id string) error {
	result := r.db.Model(&datamodel.Token{}).
		Where("id = ? AND owner = ?", id, owner).
		Delete(&datamodel.Token{})

	if result.Error != nil {
		return status.Error(codes.Internal, result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return status.Errorf(codes.NotFound, "[DeleteToken] The token id %s you specified is not found", id)
	}

	return nil
}
