package service

import (
	"strings"

	"github.com/gofrs/uuid"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Service interface
type Service interface {
	ListRole() []string
	ListUser(pageSize int, pageCursor string) ([]datamodel.User, string, error)
	CreateUser(user *datamodel.User) (*datamodel.User, error)
	GetUser(id uuid.UUID) (*datamodel.User, error)
	GetUserByLogin(login string) (*datamodel.User, error)
	UpdateUser(id uuid.UUID, user *datamodel.User) (*datamodel.User, error)
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
func (s *service) ListUser(pageSize int, pageCursor string) ([]datamodel.User, string, error) {
	return s.repository.ListUser(pageSize, pageCursor)
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

func (s *service) GetUserByLogin(login string) (*datamodel.User, error) {
	// Validation: Required field
	if login == "" {
		return nil, status.Error(codes.FailedPrecondition, "The required field `login` is not specified")
	}

	return s.repository.GetUserByLogin(login)
}

func (s *service) GetUser(id uuid.UUID) (*datamodel.User, error) {
	// Validation: Required field
	if id.IsNil() {
		return nil, status.Error(codes.FailedPrecondition, "The required field `id` is not specified")
	}
	return s.repository.GetUser(id)
}

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
