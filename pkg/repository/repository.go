package repository

import (
	"errors"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/instill-ai/mgmt-backend/internal/paginate"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/logger"
)

// Repository interface
type Repository interface {
	ListUser(pageSize int, pageToken string) ([]datamodel.User, string, int64, error)
	CreateUser(user *datamodel.User) error
	GetUser(uid uuid.UUID) (*datamodel.User, error)
	GetUserByID(id string) (*datamodel.User, error)
	UpdateUser(uid uuid.UUID, user *datamodel.User) error
	DeleteUser(uid uuid.UUID) error
	DeleteUserByID(id string) error
	GetAllUsers() ([]datamodel.User, error)
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

// ListUser lists users
// Return error types
//   - codes.InvalidArgument
//   - codes.Internal
func (r *repository) ListUser(pageSize int, pageToken string) ([]datamodel.User, string, int64, error) {
	logger, _ := logger.GetZapLogger()
	totalSize := int64(0)
	if result := r.db.Model(&datamodel.User{}).Count(&totalSize); result.Error != nil {
		logger.Error(result.Error.Error())
		return nil, "", totalSize, status.Errorf(codes.Internal, "error %v", result.Error)
	}

	queryBuilder := r.db.Model(&datamodel.User{}).Order("create_time DESC, id DESC")

	if pageSize > 0 {
		queryBuilder = queryBuilder.Limit(pageSize)
	}

	if pageToken != "" {
		createTime, id, err := paginate.DecodeToken(pageToken)
		if err != nil {
			return nil, "", totalSize, status.Errorf(codes.InvalidArgument, "Invalid page token: %s", err.Error())
		}
		queryBuilder = queryBuilder.Where("(create_time,id) < (?::timestamp, ?)", createTime, id)
	}

	var users []datamodel.User
	var createTime time.Time

	rows, err := queryBuilder.Rows()
	if err != nil {
		logger.Error(err.Error())
		return nil, "", totalSize, status.Errorf(codes.Internal, "error %v", err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var item datamodel.User
		if err = r.db.ScanRows(rows, &item); err != nil {
			logger.Error(err.Error())
			return nil, "", totalSize, status.Errorf(codes.Internal, "error %v", err.Error())
		}
		createTime = item.CreateTime
		users = append(users, item)
	}

	if len(users) > 0 {
		// Last page
		if (len(users) < pageSize) || (len(users) == pageSize && int64(len(users)) == totalSize) {
			return users, "", totalSize, nil
		}
		// Not last page
		nextPageToken := paginate.EncodeToken(createTime, (users)[len(users)-1].UID.String())
		return users, nextPageToken, totalSize, nil
	}

	return users, "", totalSize, nil
}

// CreateUser creates a new user
// Return error types
//   - codes.Internal
func (r *repository) CreateUser(user *datamodel.User) error {
	logger, _ := logger.GetZapLogger()
	if result := r.db.Model(&datamodel.User{}).Create(user); result.Error != nil {
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

// GetUser gets a user by UUID
// Return error types
//   - codes.NotFound
func (r *repository) GetUser(uid uuid.UUID) (*datamodel.User, error) {
	var user datamodel.User
	if result := r.db.First(&user, "uid = ?", uid.String()); result.Error != nil {
		return nil, status.Error(codes.NotFound, "the user is not found")
	}
	return &user, nil
}

// GetAllUsers gets all users in the database
// Return error types
//   - codes.Internal
func (r *repository) GetAllUsers() ([]datamodel.User, error) {
	logger, _ := logger.GetZapLogger()
	var users []datamodel.User
	if result := r.db.Find(&users); result.Error != nil {
		logger.Error(result.Error.Error())
		return users, status.Errorf(codes.Internal, "error %v", result.Error)
	}
	return users, nil
}

// GetUserByID gets a user by ID
// Return error types
//   - codes.NotFound
func (r *repository) GetUserByID(id string) (*datamodel.User, error) {
	var user datamodel.User
	if result := r.db.Model(&datamodel.User{}).Where("id = ?", id).First(&user); result.Error != nil {
		return nil, status.Error(codes.NotFound, "the user is not found")
	}
	return &user, nil
}

// UpdateUser updates a user by UUID
// Return error types
//   - codes.Internal
func (r *repository) UpdateUser(uid uuid.UUID, user *datamodel.User) error {
	logger, _ := logger.GetZapLogger()
	if result := r.db.Select("*").Omit("UID").Model(&datamodel.User{}).Where("uid = ?", uid.String()).Updates(user); result.Error != nil {
		logger.Error(result.Error.Error())
		return status.Errorf(codes.Internal, "error %v", result.Error)
	}
	return nil
}

// DeleteUser deletes a user by UUID
// Return error types
//   - codes.NotFound
//   - codes.Internal
func (r *repository) DeleteUser(uid uuid.UUID) error {
	logger, _ := logger.GetZapLogger()
	result := r.db.Model(&datamodel.User{}).Where("uid = ?", uid.String()).Delete(&datamodel.User{})

	if result.Error != nil {
		logger.Error(result.Error.Error())
		return status.Errorf(codes.Internal, "error %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return status.Error(codes.NotFound, "the user is not found")
	}

	return nil
}

// DeleteUserByID deletes a user by ID
// Return error types
//   - codes.NotFound
//   - codes.Internal
func (r *repository) DeleteUserByID(id string) error {
	logger, _ := logger.GetZapLogger()
	result := r.db.Model(&datamodel.User{}).
		Where("id = ?", id).
		Delete(&datamodel.User{})

	if result.Error != nil {
		logger.Error(result.Error.Error())
		return status.Errorf(codes.Internal, "error %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return status.Errorf(codes.NotFound, "the user with id %s is not found", id)
	}

	return nil
}
