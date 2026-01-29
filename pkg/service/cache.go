package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"google.golang.org/protobuf/encoding/protojson"

	mgmtpb "github.com/instill-ai/protogen-go/mgmt/v1beta"
)

const CacheTargetUser = "user"
const CacheTargetToken = "api_token"
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
func (s *service) getUserFromCacheByUID(ctx context.Context, uid uuid.UUID) *mgmtpb.User {
	i := s.getFromCacheByUID(ctx, CacheTargetUser, uid)
	if i != nil {
		return i.(*mgmtpb.User)
	} else {
		return nil
	}
}

func (s *service) setToCache(ctx context.Context, target string, src interface{}) error {
	var b []byte
	var id string
	var err error

	switch src := src.(type) {
	case *mgmtpb.User:
		b, err = protojson.Marshal(src)
		if err != nil {
			return err
		}
		id = src.Id
	}

	// Cache only by ID (UID is no longer in the protobuf)
	setCmd := s.redisClient.Set(ctx, fmt.Sprintf("%s:%s", target, id), b, 5*time.Minute)
	if setCmd.Err() != nil {
		return setCmd.Err()
	}
	return nil
}
func (s *service) setUserToCache(ctx context.Context, user *mgmtpb.User) error {
	return s.setToCache(ctx, CacheTargetUser, user)
}

func (s *service) setUserToCacheWithUID(ctx context.Context, user *mgmtpb.User, uid uuid.UUID) error {
	b, err := protojson.Marshal(user)
	if err != nil {
		return err
	}

	// Cache by ID
	if err := s.redisClient.Set(ctx, fmt.Sprintf("%s:%s", CacheTargetUser, user.Id), b, 5*time.Minute).Err(); err != nil {
		return err
	}

	// Also cache by UID for lookups by UID
	if err := s.redisClient.Set(ctx, fmt.Sprintf("%s:%s", CacheTargetUser, uid.String()), b, 5*time.Minute).Err(); err != nil {
		return err
	}

	return nil
}

func (s *service) deleteFromCacheByID(ctx context.Context, target string, id string) error {
	// Delete only by ID (UID is no longer in the protobuf)
	setCmd := s.redisClient.Del(ctx, fmt.Sprintf("%s:%s", target, id))
	if setCmd.Err() != nil {
		return setCmd.Err()
	}
	return nil
}

func (s *service) deleteUserFromCacheByID(ctx context.Context, id string) error {
	return s.deleteFromCacheByID(ctx, CacheTargetUser, id)
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
