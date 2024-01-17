package acl

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	openfga "github.com/openfga/go-sdk"
	openfgaClient "github.com/openfga/go-sdk/client"
)

type ACLClient struct {
	client               *openfgaClient.OpenFgaClient
	authorizationModelID *string
}

type Relation struct {
	UID      uuid.UUID
	Relation string
}

func NewACLClient(c *openfgaClient.OpenFgaClient, a *string) ACLClient {
	return ACLClient{
		client:               c,
		authorizationModelID: a,
	}
}

func (c *ACLClient) SetOrganizationUserMembership(orgUID uuid.UUID, userUID uuid.UUID, role string) error {
	var err error
	options := openfgaClient.ClientWriteOptions{
		AuthorizationModelId: c.authorizationModelID,
	}

	_ = c.DeleteOrganizationUserMembership(orgUID, userUID)

	body := openfgaClient.ClientWriteRequest{
		Writes: &[]openfgaClient.ClientTupleKey{
			{
				User:     fmt.Sprintf("user:%s", userUID.String()),
				Relation: role,
				Object:   fmt.Sprintf("organization:%s", orgUID.String()),
			}},
	}

	_, err = c.client.Write(context.Background()).Body(body).Options(options).Execute()
	if err != nil {
		return err
	}
	return nil
}

func (c *ACLClient) DeleteOrganizationUserMembership(orgUID uuid.UUID, userUID uuid.UUID) error {
	// var err error
	options := openfgaClient.ClientWriteOptions{
		AuthorizationModelId: c.authorizationModelID,
	}

	for _, role := range []string{"owner", "member", "pending_owner", "pending_member"} {
		body := openfgaClient.ClientWriteRequest{
			Deletes: &[]openfgaClient.ClientTupleKey{
				{
					User:     fmt.Sprintf("user:%s", userUID.String()),
					Relation: role,
					Object:   fmt.Sprintf("organization:%s", orgUID.String()),
				}}}
		_, _ = c.client.Write(context.Background()).Body(body).Options(options).Execute()

	}

	return nil
}

func (c *ACLClient) CheckOrganizationUserMembership(orgUID uuid.UUID, userUID uuid.UUID, role string) (bool, error) {
	options := openfgaClient.ClientCheckOptions{
		AuthorizationModelId: c.authorizationModelID,
	}
	body := openfgaClient.ClientCheckRequest{
		User:     fmt.Sprintf("user:%s", userUID.String()),
		Relation: role,
		Object:   fmt.Sprintf("organization:%s", orgUID.String()),
	}
	data, err := c.client.Check(context.Background()).Body(body).Options(options).Execute()
	if err != nil {
		return false, err
	}
	return *data.Allowed, nil

}

func (c *ACLClient) GetOrganizationUserMembership(orgUID uuid.UUID, userUID uuid.UUID) (string, error) {
	options := openfgaClient.ClientReadOptions{
		PageSize: openfga.PtrInt32(1),
	}
	body := openfgaClient.ClientReadRequest{
		User:   openfga.PtrString(fmt.Sprintf("user:%s", userUID.String())),
		Object: openfga.PtrString(fmt.Sprintf("organization:%s", orgUID.String())),
	}
	data, err := c.client.Read(context.Background()).Body(body).Options(options).Execute()
	if err != nil {
		return "", err
	}

	for _, tuple := range *data.Tuples {
		return *tuple.Key.Relation, nil
	}
	return "", ErrMembershipNotFound
}

func (c *ACLClient) GetOrganizationUsers(orgUID uuid.UUID) ([]*Relation, error) {
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
		data, err := c.client.Read(context.Background()).Body(body).Options(options).Execute()
		if err != nil {
			return nil, err
		}

		for _, tuple := range *data.Tuples {
			relations = append(relations, &Relation{
				UID:      uuid.FromStringOrNil(strings.Split(*tuple.Key.User, ":")[1]),
				Relation: *tuple.Key.Relation,
			})
		}
		if *data.ContinuationToken == "" {
			break
		}
		options.ContinuationToken = data.ContinuationToken
	}

	return relations, nil
}

func (c *ACLClient) GetUserOrganizations(userUID uuid.UUID) ([]*Relation, error) {
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
		data, err := c.client.Read(context.Background()).Body(body).Options(options).Execute()
		if err != nil {
			return nil, err
		}

		for _, tuple := range *data.Tuples {
			relations = append(relations, &Relation{
				UID:      uuid.FromStringOrNil(strings.Split(*tuple.Key.Object, ":")[1]),
				Relation: *tuple.Key.Relation,
			})
		}
		if *data.ContinuationToken == "" {
			break
		}
		options.ContinuationToken = data.ContinuationToken
	}

	return relations, nil
}

func (c *ACLClient) CheckPermission(objectType string, objectUID uuid.UUID, userType string, userUID uuid.UUID, code string, role string) (bool, error) {

	options := openfgaClient.ClientCheckOptions{
		AuthorizationModelId: c.authorizationModelID,
	}
	body := openfgaClient.ClientCheckRequest{
		User:     fmt.Sprintf("%s:%s", userType, userUID.String()),
		Relation: role,
		Object:   fmt.Sprintf("%s:%s", objectType, objectUID.String()),
	}
	data, err := c.client.Check(context.Background()).Body(body).Options(options).Execute()
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
	data, err = c.client.Check(context.Background()).Body(body).Options(options).Execute()

	if err != nil {
		return false, err
	}
	return *data.Allowed, nil
}
