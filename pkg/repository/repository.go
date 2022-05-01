package repository

import (
	"time"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/instill-ai/mgmt-backend/internal/paginate"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
)

type Repository interface {
	ListUser(pageSize int, pageToken string) ([]datamodel.User, string, int, error)
	CreateUser(user *datamodel.User) error
	GetUser(id uuid.UUID) (*datamodel.User, error)
	GetUserByLogin(login string) (*datamodel.User, error)
	UpdateUser(id uuid.UUID, user *datamodel.User) error
	DeleteUser(id uuid.UUID) error
	DeleteUserByLogin(login string) error
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
func (r *repository) ListUser(pageSize int, pageToken string) ([]datamodel.User, string, int, error) {
	totalSize := int64(0)
	if result := r.db.Model(&datamodel.User{}).Count(&totalSize); result.Error != nil {
		return nil, "", int(totalSize), status.Errorf(codes.Internal, "Error %v", result.Error)
	}

	queryBuilder := r.db.Model(&datamodel.User{}).Order("created_at DESC, id DESC")

	if pageSize > 0 {
		queryBuilder = queryBuilder.Limit(pageSize)
	}

	if pageToken != "" {
		createdAt, id, err := paginate.DecodeToken(pageToken)
		if err != nil {
			return nil, "", int(totalSize), status.Errorf(codes.InvalidArgument, "Invalid page token: %s", err.Error())
		}
		queryBuilder = queryBuilder.Where("(created_at,id) < (?::timestamp, ?)", createdAt, id)
	}

	var users []datamodel.User
	var createdAt time.Time

	rows, err := queryBuilder.Rows()
	if err != nil {
		return nil, "", int(totalSize), err
	}
	defer rows.Close()
	for rows.Next() {
		var item datamodel.User
		if err = r.db.ScanRows(rows, &item); err != nil {
			return nil, "", int(totalSize), status.Errorf(codes.Internal, "Error %v", err.Error())
		}
		createdAt = item.CreatedAt
		users = append(users, item)
	}

	if len(users) > 0 {
		nextPageToken := paginate.EncodeToken(createdAt, (users)[len(users)-1].Id.String())
		return users, nextPageToken, int(totalSize), nil
	}

	return nil, "", int(totalSize), nil
}

// CreateUser creates a new user
func (r *repository) CreateUser(user *datamodel.User) error {
	if result := r.db.Model(&datamodel.User{}).Create(user); result.Error != nil {
		return status.Errorf(codes.Internal, "Error %v", result.Error)
	}
	return nil
}

// GetUser gets a user by uuid Id
func (r *repository) GetUser(id uuid.UUID) (*datamodel.User, error) {
	var user datamodel.User
	if result := r.db.First(&user, "id = ?", id.String()); result.Error != nil {

		return nil, status.Error(codes.NotFound, "The user is not found")
	}
	return &user, nil
}

// GetUserByLogin gets a user by login
func (r *repository) GetUserByLogin(login string) (*datamodel.User, error) {
	var user datamodel.User
	if result := r.db.Model(&datamodel.User{}).Where("login = ?", login).First(&user); result.Error != nil {
		return nil, status.Errorf(codes.NotFound, "The user with login `%s` specified is not found", login)
	}
	return &user, nil
}

// UpdateUser updates a user by uuid Id
func (r *repository) UpdateUser(id uuid.UUID, user *datamodel.User) error {
	if result := r.db.Select("*").Omit("Id").Model(&datamodel.User{}).Where("id = ?", id.String()).Updates(user); result.Error != nil {
		return status.Errorf(codes.Internal, "Error %v", result.Error)
	}
	return nil
}

// DeleteUser deletes a user by uuid Id
func (r *repository) DeleteUser(id uuid.UUID) error {
	result := r.db.Model(&datamodel.User{}).Where("id = ?", id.String()).Delete(&datamodel.User{})

	if result.Error != nil {
		return status.Errorf(codes.Internal, "Error %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return status.Error(codes.NotFound, "The user is not found")
	}

	return nil
}

// DeleteUserByLogin deletes a user by login
func (r *repository) DeleteUserByLogin(login string) error {
	result := r.db.Model(&datamodel.User{}).Where("login = ?", login).Delete(&datamodel.User{})

	if result.Error != nil {
		return status.Errorf(codes.Internal, "Error %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return status.Errorf(codes.NotFound, "The user with login `%s` specified is not found", login)
	}

	return nil
}
