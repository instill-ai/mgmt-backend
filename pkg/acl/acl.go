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
	authorizationModelId *string
}

func NewACLClient(c *openfgaClient.OpenFgaClient, a *string) ACLClient {
	return ACLClient{
		client:               c,
		authorizationModelId: a,
	}
}

func (c *ACLClient) SetOrganizationUserMembership(orgUID uuid.UUID, userUID uuid.UUID) error {
	exist, err := c.GetOrganizationUserMembership(orgUID, userUID)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}
	options := openfgaClient.ClientWriteOptions{
		AuthorizationModelId: c.authorizationModelId,
	}
	body := openfgaClient.ClientWriteRequest{
		Writes: &[]openfgaClient.ClientTupleKey{
			{
				User:     fmt.Sprintf("user:%s", userUID.String()),
				Relation: "member",
				Object:   fmt.Sprintf("organization:%s", orgUID.String()),
			}}}
	_, err = c.client.Write(context.Background()).Body(body).Options(options).Execute()
	if err != nil {
		return err
	}
	return nil
}

func (c *ACLClient) DeleteOrganizationUserMembership(orgUID uuid.UUID, userUID uuid.UUID) error {
	exist, err := c.GetOrganizationUserMembership(orgUID, userUID)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}
	options := openfgaClient.ClientWriteOptions{
		AuthorizationModelId: c.authorizationModelId,
	}
	body := openfgaClient.ClientWriteRequest{
		Deletes: &[]openfgaClient.ClientTupleKey{
			{
				User:     fmt.Sprintf("user:%s", userUID.String()),
				Relation: "member",
				Object:   fmt.Sprintf("organization:%s", orgUID.String()),
			}}}
	_, err = c.client.Write(context.Background()).Body(body).Options(options).Execute()
	if err != nil {
		return err
	}
	return nil
}

func (c *ACLClient) GetOrganizationUserMembership(orgUID uuid.UUID, userUID uuid.UUID) (bool, error) {
	options := openfgaClient.ClientCheckOptions{
		AuthorizationModelId: c.authorizationModelId,
	}
	body := openfgaClient.ClientCheckRequest{
		User:     fmt.Sprintf("user:%s", userUID.String()),
		Relation: "member",
		Object:   fmt.Sprintf("organization:%s", orgUID.String()),
	}
	data, err := c.client.Check(context.Background()).Body(body).Options(options).Execute()
	if err != nil {
		return false, err
	}

	return *data.Allowed, nil
}

func (c *ACLClient) GetOrganizationUsers(orgUID uuid.UUID) ([]uuid.UUID, error) {
	options := openfgaClient.ClientReadOptions{
		PageSize: openfga.PtrInt32(1),
	}
	// Find all relationship tuples where any user has a relationship as any relation with a particular document
	body := openfgaClient.ClientReadRequest{
		Object:   openfga.PtrString(fmt.Sprintf("organization:%s", orgUID.String())),
		Relation: openfga.PtrString("member"),
	}

	users := []uuid.UUID{}
	for {
		data, err := c.client.Read(context.Background()).Body(body).Options(options).Execute()
		if err != nil {
			return nil, err
		}

		for _, tuple := range *data.Tuples {
			userUIDStr := strings.Split(*tuple.Key.User, ":")[1]
			userUID, err := uuid.FromString(userUIDStr)
			if err != nil {
				return nil, err
			}
			users = append(users, userUID)
		}
		if *data.ContinuationToken == "" {
			break
		}
		options.ContinuationToken = data.ContinuationToken
	}

	return users, nil
}

func (c *ACLClient) GetUserOrganizations(userUID uuid.UUID) ([]uuid.UUID, error) {
	options := openfgaClient.ClientReadOptions{
		PageSize: openfga.PtrInt32(1),
	}
	// Find all relationship tuples where any user has a relationship as any relation with a particular document
	body := openfgaClient.ClientReadRequest{
		User:     openfga.PtrString(fmt.Sprintf("user:%s", userUID.String())),
		Relation: openfga.PtrString("member"),
		Object:   openfga.PtrString("organization:"),
	}

	orgs := []uuid.UUID{}
	for {
		data, err := c.client.Read(context.Background()).Body(body).Options(options).Execute()
		if err != nil {
			return nil, err
		}

		for _, tuple := range *data.Tuples {
			orgUIDStr := strings.Split(*tuple.Key.Object, ":")[1]
			orgUID, err := uuid.FromString(orgUIDStr)
			if err != nil {
				return nil, err
			}
			orgs = append(orgs, orgUID)
		}
		if *data.ContinuationToken == "" {
			break
		}
		options.ContinuationToken = data.ContinuationToken
	}

	return orgs, nil
}
