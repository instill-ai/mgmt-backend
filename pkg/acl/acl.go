package acl

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/redis/go-redis/v9"

	openfga "github.com/openfga/go-sdk"
	openfgaClient "github.com/openfga/go-sdk/client"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/internal/resource"
	"github.com/instill-ai/mgmt-backend/pkg/constant"

	errorsx "github.com/instill-ai/x/errors"
)

type ACLClient struct {
	writeClient *openfgaClient.OpenFgaClient
	readClient  *openfgaClient.OpenFgaClient
	redisClient *redis.Client
}

type Relation struct {
	UID      uuid.UUID
	Relation string
}

type Mode string

const (
	ReadMode  Mode = "read"
	WriteMode Mode = "write"
)

// NewACLClient creates a new ACL client.
func NewACLClient(wc *openfgaClient.OpenFgaClient, rc *openfgaClient.OpenFgaClient, redisClient *redis.Client) ACLClient {
	if rc == nil {
		rc = wc
	}

	return ACLClient{
		writeClient: wc,
		readClient:  rc,
		redisClient: redisClient,
	}
}

func (c *ACLClient) getClient(ctx context.Context, mode Mode) *openfgaClient.OpenFgaClient {
	userUID := resource.GetRequestSingleHeader(ctx, constant.HeaderUserUIDKey)

	if mode == WriteMode {
		// To solve the read-after-write inconsistency problem,
		// we will direct the user to read from the primary database for a certain time frame
		// to ensure that the data is synchronized from the primary DB to the replica DB.
		_ = c.redisClient.Set(ctx, fmt.Sprintf("db_pin_user:%s", userUID), time.Now(), time.Duration(config.Config.OpenFGA.Replica.ReplicationTimeFrame)*time.Second)
	}

	// If the user is pinned, we will use the primary database for querying.
	if !errors.Is(c.redisClient.Get(ctx, fmt.Sprintf("db_pin_user:%s", userUID)).Err(), redis.Nil) {
		return c.writeClient
	}
	if mode == ReadMode {
		return c.readClient
	}
	return c.writeClient
}

// SetOrganizationUserMembership sets the membership of a user in an organization.
func (c *ACLClient) SetOrganizationUserMembership(ctx context.Context, orgUID uuid.UUID, userUID uuid.UUID, role string) error {
	var err error

	_ = c.DeleteOrganizationUserMembership(ctx, orgUID, userUID)

	body := openfgaClient.ClientWriteRequest{
		Writes: []openfgaClient.ClientTupleKey{
			{
				User:     fmt.Sprintf("user:%s", userUID.String()),
				Relation: role,
				Object:   fmt.Sprintf("organization:%s", orgUID.String()),
			}},
	}

	_, err = c.getClient(ctx, WriteMode).Write(ctx).Body(body).Execute()
	if err != nil {
		return err
	}
	return nil
}

// DeleteOrganizationUserMembership deletes the membership of a user in an organization.
func (c *ACLClient) DeleteOrganizationUserMembership(ctx context.Context, orgUID uuid.UUID, userUID uuid.UUID) error {

	for _, role := range []string{"owner", "admin", "member", "pending_owner", "pending_admin", "pending_member"} {
		body := openfgaClient.ClientWriteRequest{
			Deletes: []openfgaClient.ClientTupleKeyWithoutCondition{
				{
					User:     fmt.Sprintf("user:%s", userUID.String()),
					Relation: role,
					Object:   fmt.Sprintf("organization:%s", orgUID.String()),
				}}}
		_, _ = c.getClient(ctx, WriteMode).Write(ctx).Body(body).Execute()

	}

	return nil
}

// CheckOrganizationUserMembership checks if a user has a specific role in an organization.
func (c *ACLClient) CheckOrganizationUserMembership(ctx context.Context, orgUID uuid.UUID, userUID uuid.UUID, role string) (bool, error) {
	body := openfgaClient.ClientCheckRequest{
		User:     fmt.Sprintf("user:%s", userUID.String()),
		Relation: role,
		Object:   fmt.Sprintf("organization:%s", orgUID.String()),
	}
	data, err := c.getClient(ctx, ReadMode).Check(ctx).Body(body).Execute()
	if err != nil {
		return false, err
	}
	return *data.Allowed, nil

}

// GetOrganizationUserMembership gets the membership of a user in an organization.
func (c *ACLClient) GetOrganizationUserMembership(ctx context.Context, orgUID uuid.UUID, userUID uuid.UUID) (string, error) {
	options := openfgaClient.ClientReadOptions{
		PageSize: openfga.PtrInt32(1),
	}
	body := openfgaClient.ClientReadRequest{
		User:   openfga.PtrString(fmt.Sprintf("user:%s", userUID.String())),
		Object: openfga.PtrString(fmt.Sprintf("organization:%s", orgUID.String())),
	}
	data, err := c.getClient(ctx, ReadMode).Read(ctx).Body(body).Options(options).Execute()
	if err != nil {
		return "", err
	}

	for _, tuple := range data.Tuples {
		return tuple.Key.Relation, nil
	}
	return "", errorsx.ErrMembershipNotFound
}

// GetOrganizationUsers gets the users of an organization.
func (c *ACLClient) GetOrganizationUsers(ctx context.Context, orgUID uuid.UUID) ([]*Relation, error) {
	options := openfgaClient.ClientReadOptions{
		PageSize: openfga.PtrInt32(1),
	}
	// Find all relationship tuples where any user has a relationship as any relation with a particular document
	body := openfgaClient.ClientReadRequest{
		Object: openfga.PtrString(fmt.Sprintf("organization:%s", orgUID.String())),
		// Relation: openfga.PtrString("member"),
	}

	relations := []*Relation{}
	for {
		data, err := c.getClient(ctx, ReadMode).Read(ctx).Body(body).Options(options).Execute()
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
func (c *ACLClient) GetUserOrganizations(ctx context.Context, userUID uuid.UUID) ([]*Relation, error) {
	options := openfgaClient.ClientReadOptions{
		PageSize: openfga.PtrInt32(1),
	}
	// Find all relationship tuples where any user has a relationship as any relation with a particular document
	body := openfgaClient.ClientReadRequest{
		User:   openfga.PtrString(fmt.Sprintf("user:%s", userUID.String())),
		Object: openfga.PtrString("organization:"),
	}

	relations := []*Relation{}
	for {
		data, err := c.getClient(ctx, ReadMode).Read(ctx).Body(body).Options(options).Execute()
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

// CheckPermission checks if a user has a specific permission on an object.
func (c *ACLClient) CheckPermission(ctx context.Context, objectType string, objectUID uuid.UUID, userType string, userUID uuid.UUID, code string, role string) (bool, error) {

	body := openfgaClient.ClientCheckRequest{
		User:     fmt.Sprintf("%s:%s", userType, userUID.String()),
		Relation: role,
		Object:   fmt.Sprintf("%s:%s", objectType, objectUID.String()),
	}
	data, err := c.getClient(ctx, ReadMode).Check(ctx).Body(body).Execute()
	if err != nil {
		return false, err
	}
	if *data.Allowed {
		return *data.Allowed, nil
	}

	if code == "" {
		return false, nil
	}
	body = openfgaClient.ClientCheckRequest{
		User:     fmt.Sprintf("code:%s", code),
		Relation: role,
		Object:   fmt.Sprintf("%s:%s", objectType, objectUID.String()),
	}
	data, err = c.getClient(ctx, ReadMode).Check(ctx).Body(body).Execute()

	if err != nil {
		return false, err
	}
	return *data.Allowed, nil
}
