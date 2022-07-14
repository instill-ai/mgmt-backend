package service

import (
	"strings"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/repository"
)

// Service interface
type Service interface {
	ListRole() []string
	ListUser(pageSize int, pageToken string) ([]datamodel.User, string, int, error)
	CreateUser(user *datamodel.User) (*datamodel.User, error)
	GetUser(uid uuid.UUID) (*datamodel.User, error)
	GetUserByID(id string) (*datamodel.User, error)
	UpdateUser(uid uuid.UUID, user *datamodel.User) (*datamodel.User, error)
	DeleteUser(uid uuid.UUID) error
	DeleteUserByID(id string) error
}

type service struct {
	repository repository.Repository
}

// NewService initiates a service instance
func NewService(r repository.Repository) Service {
	return &service{
		repository: r,
	}
}

// ListRole lists names of all roles
func (s *service) ListRole() []string {
	return ListAllowedRoleName()
}

// ListUser lists all users
// Return error types
//   - codes.InvalidArgument
//   - codes.Internal
func (s *service) ListUser(pageSize int, pageToken string) ([]datamodel.User, string, int, error) {
	return s.repository.ListUser(pageSize, pageToken)
}

// CreateUser creates an user instance
// Return error types
//   - codes.InvalidArgument
//   - codes.NotFound
//   - codes.Internal
func (s *service) CreateUser(user *datamodel.User) (*datamodel.User, error) {
	//TODO: validate spec JSON schema

	//Validation: role field
	if user.Role.Valid {
		if r := Role(user.Role.String); !ValidateRole(r) {
			return nil, status.Errorf(codes.InvalidArgument, "`role` %s in the body is not valid. Please choose from: [ %v ]", r.GetName(), strings.Join(s.ListRole(), ", "))
		}
	}

	if err := s.repository.CreateUser(user); err != nil {
		return nil, err
	}

	return s.repository.GetUserByID(user.ID)
}

// GetUserByID gets a user by ID
// Return error types
//   - codes.InvalidArgument
//   - codes.NotFound
func (s *service) GetUserByID(id string) (*datamodel.User, error) {
	// Validation: Required field
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "the required field `id` is not specified")
	}

	return s.repository.GetUserByID(id)
}

// GetUser gets a user by UUID
// Return error types
//   - codes.InvalidArgument
//   - codes.NotFound
func (s *service) GetUser(uid uuid.UUID) (*datamodel.User, error) {
	// Validation: Required field
	if uid.IsNil() {
		return nil, status.Error(codes.InvalidArgument, "the required field `uid` is not specified")
	}
	return s.repository.GetUser(uid)
}

// UpdateUser updates a user by UUID
// Return error types
//   - codes.InvalidArgument
//   - codes.NotFound
//   - codes.Internal
func (s *service) UpdateUser(uid uuid.UUID, user *datamodel.User) (*datamodel.User, error) {
	// Validation: Required field
	if uid.IsNil() {
		return nil, status.Error(codes.InvalidArgument, "the required field `uid` is not specified")
	}

	//Validation: role field
	if user.Role.Valid {
		if r := Role(user.Role.String); !ValidateRole(r) {
			return nil, status.Errorf(codes.InvalidArgument, "`role` %s in the body is not valid. Please choose from: [ %v ]", r.GetName(), strings.Join(s.ListRole(), ", "))
		}
	}

	// Check if the user exists
	if _, err := s.repository.GetUser(uid); err != nil {
		return nil, err
	}

	// Update the user
	if err := s.repository.UpdateUser(uid, user); err != nil {
		return nil, err
	}

	// Get the updated user
	return s.repository.GetUser(uid)
}

// DeleteUser deletes a user by UUID
// Return error types
//   - codes.InvalidArgument
//   - codes.NotFound
//   - codes.Internal
func (s *service) DeleteUser(uid uuid.UUID) error {
	// Validation: Required field
	if uid.IsNil() {
		return status.Error(codes.InvalidArgument, "the required field `uid` is not specified")
	}
	return s.repository.DeleteUser(uid)
}

// DeleteUserByID deletes a user by ID
// Return error types
//   - codes.InvalidArgument
//   - codes.NotFound
//   - codes.Internal
func (s *service) DeleteUserByID(id string) error {
	// Validation: Required field
	if id == "" {
		return status.Error(codes.InvalidArgument, "the required field `id` is not specified")
	}

	return s.repository.DeleteUserByID(id)
}
