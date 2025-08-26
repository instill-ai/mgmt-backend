package acl

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	openfga "github.com/openfga/go-sdk"
	openfgaClient "github.com/openfga/go-sdk/client"

	errorsx "github.com/instill-ai/x/errors"
	openfgax "github.com/instill-ai/x/openfga"
)

// Management-specific object types
const (
	ObjectTypeOrganization openfgax.ObjectType = "organization"
	ObjectTypeUser         openfgax.ObjectType = "user"
)

// aclClient wraps the x/openfga Client with mgmt-backend specific operations
type aclClient struct {
	openfgax.Client
}

// ACLClient defines the interface for mgmt-backend ACL operations
type ACLClient interface {
	openfgax.Client

	SetOrganizationUserMembership(ctx context.Context, orgUID uuid.UUID, userUID uuid.UUID, role string) error
	DeleteOrganizationUserMembership(ctx context.Context, orgUID uuid.UUID, userUID uuid.UUID) error
	CheckOrganizationUserMembership(ctx context.Context, orgUID uuid.UUID, userUID uuid.UUID, role string) (bool, error)
	GetOrganizationUserMembership(ctx context.Context, orgUID uuid.UUID, userUID uuid.UUID) (string, error)
	GetOrganizationUsers(ctx context.Context, orgUID uuid.UUID) ([]*Relation, error)
	GetUserOrganizations(ctx context.Context, userUID uuid.UUID) ([]*Relation, error)
}

type Relation struct {
	UID      uuid.UUID
	Relation string
}

// NewFGAClient creates a new mgmt-backend specific FGA client
func NewFGAClient(client openfgax.Client) ACLClient {
	return &aclClient{Client: client}
}

// SetOrganizationUserMembership sets the membership of a user in an organization.
func (c *aclClient) SetOrganizationUserMembership(ctx context.Context, orgUID uuid.UUID, userUID uuid.UUID, role string) error {
	_ = c.DeleteOrganizationUserMembership(ctx, orgUID, userUID)

	body := openfgaClient.ClientWriteRequest{
		Writes: []openfgaClient.ClientTupleKey{
			{
				User:     fmt.Sprintf("%s:%s", openfgax.OwnerTypeUser, userUID.String()),
				Relation: role,
				Object:   fmt.Sprintf("%s:%s", ObjectTypeOrganization, orgUID.String()),
			}},
	}

	_, err := c.SDKClient().Write(ctx).Body(body).Execute()
	return err
}

// DeleteOrganizationUserMembership deletes the membership of a user in an organization.
func (c *aclClient) DeleteOrganizationUserMembership(ctx context.Context, orgUID uuid.UUID, userUID uuid.UUID) error {
	for _, role := range []string{"owner", "admin", "member", "pending_owner", "pending_admin", "pending_member"} {
		body := openfgaClient.ClientWriteRequest{
			Deletes: []openfgaClient.ClientTupleKeyWithoutCondition{
				{
					User:     fmt.Sprintf("%s:%s", openfgax.OwnerTypeUser, userUID.String()),
					Relation: role,
					Object:   fmt.Sprintf("%s:%s", ObjectTypeOrganization, orgUID.String()),
				}}}
		_, _ = c.SDKClient().Write(ctx).Body(body).Execute()
	}

	return nil
}

// CheckOrganizationUserMembership checks if a user has a specific role in an organization.
func (c *aclClient) CheckOrganizationUserMembership(ctx context.Context, orgUID uuid.UUID, userUID uuid.UUID, role string) (bool, error) {
	body := openfgaClient.ClientCheckRequest{
		User:     fmt.Sprintf("%s:%s", openfgax.OwnerTypeUser, userUID.String()),
		Relation: role,
		Object:   fmt.Sprintf("%s:%s", ObjectTypeOrganization, orgUID.String()),
	}
	data, err := c.SDKClient().Check(ctx).Body(body).Execute()
	if err != nil {
		return false, err
	}
	return *data.Allowed, nil
}

// GetOrganizationUserMembership gets the membership of a user in an organization.
func (c *aclClient) GetOrganizationUserMembership(ctx context.Context, orgUID uuid.UUID, userUID uuid.UUID) (string, error) {
	options := openfgaClient.ClientReadOptions{
		PageSize: openfga.PtrInt32(1),
	}
	body := openfgaClient.ClientReadRequest{
		User:   openfga.PtrString(fmt.Sprintf("%s:%s", openfgax.OwnerTypeUser, userUID.String())),
		Object: openfga.PtrString(fmt.Sprintf("%s:%s", ObjectTypeOrganization, orgUID.String())),
	}
	data, err := c.SDKClient().Read(ctx).Body(body).Options(options).Execute()
	if err != nil {
		return "", err
	}

	for _, tuple := range data.Tuples {
		return tuple.Key.Relation, nil
	}
	return "", errorsx.ErrMembershipNotFound
}

// GetOrganizationUsers gets the users of an organization.
func (c *aclClient) GetOrganizationUsers(ctx context.Context, orgUID uuid.UUID) ([]*Relation, error) {
	options := openfgaClient.ClientReadOptions{
		PageSize: openfga.PtrInt32(1),
	}
	// Find all relationship tuples where any user has a relationship as any relation with a particular document
	body := openfgaClient.ClientReadRequest{
		Object: openfga.PtrString(fmt.Sprintf("%s:%s", ObjectTypeOrganization, orgUID.String())),
	}

	relations := []*Relation{}
	for {
		data, err := c.SDKClient().Read(ctx).Body(body).Options(options).Execute()
		if err != nil {
			return nil, err
		}

		for _, tuple := range data.Tuples {
			relations = append(relations, &Relation{
				UID:      uuid.FromStringOrNil(strings.Split(tuple.Key.User, ":")[1]),
				Relation: tuple.Key.Relation,
			})
		}
		if data.ContinuationToken == "" {
			break
		}
		options.ContinuationToken = &data.ContinuationToken
	}

	return relations, nil
}

// GetUserOrganizations gets the organizations of a user.
func (c *aclClient) GetUserOrganizations(ctx context.Context, userUID uuid.UUID) ([]*Relation, error) {
	options := openfgaClient.ClientReadOptions{
		PageSize: openfga.PtrInt32(1),
	}
	// Find all relationship tuples where any user has a relationship as any relation with a particular document
	body := openfgaClient.ClientReadRequest{
		User:   openfga.PtrString(fmt.Sprintf("%s:%s", openfgax.OwnerTypeUser, userUID.String())),
		Object: openfga.PtrString(fmt.Sprintf("%s:", ObjectTypeOrganization)),
	}

	relations := []*Relation{}
	for {
		data, err := c.SDKClient().Read(ctx).Body(body).Options(options).Execute()
		if err != nil {
			return nil, err
		}

		for _, tuple := range data.Tuples {
			relations = append(relations, &Relation{
				UID:      uuid.FromStringOrNil(strings.Split(tuple.Key.Object, ":")[1]),
				Relation: tuple.Key.Relation,
			})
		}
		if data.ContinuationToken == "" {
			break
		}
		options.ContinuationToken = &data.ContinuationToken
	}

	return relations, nil
}
