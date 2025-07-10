package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"google.golang.org/protobuf/encoding/protojson"

	mgmtpb "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
)

const CacheTargetUser = "user"
const CacheTargetToken = "api_token"
const CacheTargetOrganization = "organization"
const CacheTargetUserPasswordHash = "user_password_hash"

func (s *service) getFromCacheByID(ctx context.Context, target string, id string) interface{} {
	getCmd := s.redisClient.Get(ctx, fmt.Sprintf("%s:%s", target, id))
	if getCmd.Err() == nil {
		b, err := getCmd.Bytes()
		if err == nil {
			if target == CacheTargetUser {
				pbUser := &mgmtpb.User{}
				if err := protojson.Unmarshal(b, pbUser); err == nil {
					return pbUser
				}
			} else {
				pbOrg := &mgmtpb.Organization{}
				if err := protojson.Unmarshal(b, pbOrg); err == nil {
					return pbOrg
				}
			}
		}
	}
	return nil
}
func (s *service) getFromCacheByUID(ctx context.Context, target string, uid uuid.UUID) interface{} {
	return s.getFromCacheByID(ctx, target, uid.String())
}
func (s *service) getUserFromCacheByID(ctx context.Context, id string) *mgmtpb.User {
	i := s.getFromCacheByID(ctx, CacheTargetUser, id)
	if i != nil {
		return i.(*mgmtpb.User)
	} else {
		return nil
	}
}
func (s *service) getOrganizationFromCacheByID(ctx context.Context, id string) *mgmtpb.Organization {
	i := s.getFromCacheByID(ctx, CacheTargetOrganization, id)
	if i != nil {
		return i.(*mgmtpb.Organization)
	} else {
		return nil
	}
}
func (s *service) getUserFromCacheByUID(ctx context.Context, uid uuid.UUID) *mgmtpb.User {
	i := s.getFromCacheByUID(ctx, CacheTargetUser, uid)
	if i != nil {
		return i.(*mgmtpb.User)
	} else {
		return nil
	}
}
func (s *service) getOrganizationFromCacheByUID(ctx context.Context, uid uuid.UUID) *mgmtpb.Organization {
	i := s.getFromCacheByUID(ctx, CacheTargetOrganization, uid)
	if i != nil {
		return i.(*mgmtpb.Organization)
	} else {
		return nil
	}
}

func (s *service) setToCache(ctx context.Context, target string, src interface{}) error {
	var b []byte
	var id string
	var uid string
	var err error

	switch src := src.(type) {
	case *mgmtpb.User:
		b, err = protojson.Marshal(src)
		if err != nil {
			return err
		}
		id = src.Id
		uid = *src.Uid
	case *mgmtpb.Organization:
		b, err = protojson.Marshal(src)
		if err != nil {
			return err
		}
		id = src.Id
		uid = src.Uid
	}

	setCmd := s.redisClient.Set(ctx, fmt.Sprintf("%s:%s", target, id), b, 5*time.Minute)
	if setCmd.Err() != nil {
		return setCmd.Err()
	}
	setCmd = s.redisClient.Set(ctx, fmt.Sprintf("%s:%s", target, uid), b, 5*time.Minute)
	if setCmd.Err() != nil {
		return setCmd.Err()
	}
	return nil
}
func (s *service) setUserToCache(ctx context.Context, user *mgmtpb.User) error {
	return s.setToCache(ctx, CacheTargetUser, user)
}
func (s *service) setOrganizationToCache(ctx context.Context, org *mgmtpb.Organization) error {
	return s.setToCache(ctx, CacheTargetOrganization, org)
}

func (s *service) deleteFromCacheByID(ctx context.Context, target string, id string) error {

	var uid string
	if target == CacheTargetUser {
		user := s.getFromCacheByID(ctx, target, id)
		if user != nil {
			uid = *user.(*mgmtpb.User).Uid
		}

	} else {

		org := s.getFromCacheByID(ctx, target, id)
		if org != nil {
			uid = org.(*mgmtpb.Organization).Uid
		}
	}

	setCmd := s.redisClient.Del(ctx, fmt.Sprintf("%s:%s", target, id))
	if setCmd.Err() != nil {
		return setCmd.Err()
	}
	if uid != "" {
		setCmd = s.redisClient.Del(ctx, fmt.Sprintf("%s:%s", target, uid))
		if setCmd.Err() != nil {
			return setCmd.Err()
		}
	}

	return nil
}

func (s *service) deleteUserFromCacheByID(ctx context.Context, id string) error {
	return s.deleteFromCacheByID(ctx, CacheTargetUser, id)
}
func (s *service) deleteOrganizationFromCacheByID(ctx context.Context, id string) error {
	return s.deleteFromCacheByID(ctx, CacheTargetOrganization, id)
}

func (s *service) getUserPasswordHashFromCache(ctx context.Context, uid uuid.UUID) string {
	getCmd := s.redisClient.Get(ctx, fmt.Sprintf("%s:%s", CacheTargetUserPasswordHash, uid))
	if getCmd.Err() == nil {
		return getCmd.Val()
	}
	return ""
}

func (s *service) setUserPasswordHashToCache(ctx context.Context, uid uuid.UUID, hash string) error {

	setCmd := s.redisClient.Set(ctx, fmt.Sprintf("%s:%s", CacheTargetUserPasswordHash, uid), hash, 5*time.Minute)
	if setCmd.Err() != nil {
		return setCmd.Err()
	}

	return nil
}

func (s *service) deleteUserPasswordHashFromCache(ctx context.Context, uid uuid.UUID) error {

	setCmd := s.redisClient.Del(ctx, fmt.Sprintf("%s:%s", CacheTargetUserPasswordHash, uid))
	if setCmd.Err() != nil {
		return setCmd.Err()
	}

	return nil
}

func (s *service) getAPITokenFromCache(ctx context.Context, token string) uuid.UUID {
	getCmd := s.redisClient.Get(ctx, fmt.Sprintf("%s:%s:user_uid", CacheTargetToken, token))
	if getCmd.Err() == nil {
		return uuid.FromStringOrNil(getCmd.Val())
	}
	return uuid.Nil
}

func (s *service) setAPITokenToCache(ctx context.Context, token string, userUID uuid.UUID, expire time.Time) error {

	setCmd := s.redisClient.Set(ctx, fmt.Sprintf("%s:%s:user_uid", CacheTargetToken, token), userUID.String(), 5*time.Minute)
	if setCmd.Err() != nil {
		return setCmd.Err()
	}

	return nil
}

func (s *service) deleteAPITokenFromCache(ctx context.Context, token string) error {

	setCmd := s.redisClient.Del(ctx, fmt.Sprintf("%s:%s:user_uid", CacheTargetToken, token))
	if setCmd.Err() != nil {
		return setCmd.Err()
	}

	return nil
}
