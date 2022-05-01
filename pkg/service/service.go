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
	GetUser(id uuid.UUID) (*datamodel.User, error)
	GetUserByLogin(login string) (*datamodel.User, error)
	UpdateUser(id uuid.UUID, user *datamodel.User) (*datamodel.User, error)
	DeleteUser(id uuid.UUID) error
	DeleteUserByLogin(login string) error
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
func (s *service) ListUser(pageSize int, pageToken string) ([]datamodel.User, string, int, error) {
	return s.repository.ListUser(pageSize, pageToken)
}

// CreateUser creates an user instance
func (s *service) CreateUser(user *datamodel.User) (*datamodel.User, error) {
	//TODO: validate spec JSON schema

	//Validation: role field
	if user.Role.Valid {
		if r := Role(user.Role.String); !ValidateRole(r) {
			return nil, status.Errorf(codes.FailedPrecondition, "The field `role` %s in the body is not valid. Please choose from: [ %v ]", r.GetName(), strings.Join(s.ListRole(), ", "))
		}
	}

	if err := s.repository.CreateUser(user); err != nil {
		return nil, err
	}

	return s.repository.GetUserByLogin(user.Login)
}

// GetUserByLogin gets a user by login
func (s *service) GetUserByLogin(login string) (*datamodel.User, error) {
	// Validation: Required field
	if login == "" {
		return nil, status.Error(codes.FailedPrecondition, "The required field `login` is not specified")
	}

	return s.repository.GetUserByLogin(login)
}

// GetUser gets a user by uuid ID
func (s *service) GetUser(id uuid.UUID) (*datamodel.User, error) {
	// Validation: Required field
	if id.IsNil() {
		return nil, status.Error(codes.FailedPrecondition, "The required field `id` is not specified")
	}
	return s.repository.GetUser(id)
}

// UpdateUser updates a user by uuid ID
func (s *service) UpdateUser(id uuid.UUID, user *datamodel.User) (*datamodel.User, error) {
	// Validation: Required field
	if id.IsNil() {
		return nil, status.Error(codes.FailedPrecondition, "The required field `id` is not specified")
	}

	//Validation: role field
	if user.Role.Valid {
		if r := Role(user.Role.String); !ValidateRole(r) {
			return nil, status.Errorf(codes.FailedPrecondition, "The field `role` %s in the body is not valid. Please choose from: [ %v ]", r.GetName(), strings.Join(s.ListRole(), ", "))
		}
	}

	// // Get the user
	// if _, err := s.repository.GetUser(id); err != nil {
	// return nil, err
	// }

	// Update the user
	if err := s.repository.UpdateUser(id, user); err != nil {
		return nil, err
	}

	// Get the updated user
	return s.repository.GetUser(id)
}

// DeleteUser deletes a user by uuid ID
func (s *service) DeleteUser(id uuid.UUID) error {
	// Validation: Required field
	if id.IsNil() {
		return status.Error(codes.FailedPrecondition, "The required field `id` is not specified")
	}
	return s.repository.DeleteUser(id)
}

// DeleteUserByLogin deletes a user by login
func (s *service) DeleteUserByLogin(login string) error {
	// Validation: Required field
	if login == "" {
		return status.Error(codes.FailedPrecondition, "The required field `login` is not specified")
	}

	return s.repository.DeleteUserByLogin(login)
}
