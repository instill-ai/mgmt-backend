package service

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofrs/uuid"
	"golang.org/x/image/draw"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/instill-ai/mgmt-backend/pkg/datamodel"

	mgmtpb "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
	logx "github.com/instill-ai/x/log"
)

// maps for user owner type
var (
	PBUserType2DBUserType = map[mgmtpb.OwnerType]string{
		mgmtpb.OwnerType_OWNER_TYPE_UNSPECIFIED:  "unspecified",
		mgmtpb.OwnerType_OWNER_TYPE_USER:         "user",
		mgmtpb.OwnerType_OWNER_TYPE_ORGANIZATION: "organization",
	}
	DBUserType2PBUserType = map[string]mgmtpb.OwnerType{
		"unspecified":  mgmtpb.OwnerType_OWNER_TYPE_UNSPECIFIED,
		"user":         mgmtpb.OwnerType_OWNER_TYPE_USER,
		"organization": mgmtpb.OwnerType_OWNER_TYPE_ORGANIZATION,
	}
)

func (s *service) compressAvatar(profileAvatar string) (string, error) {

	if strings.HasPrefix(profileAvatar, "http") {
		response, err := http.Get(profileAvatar)
		if err == nil {
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err == nil {
				mimeType := strings.Split(mimetype.Detect(body).String(), ";")[0]
				profileAvatar = fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(body))
			} else {
				return "", nil
			}
		} else {
			return "", nil
		}
	}
	// Due to the local env, we don't set the `InstillCoreHost` config, the avatar path is not working.
	// As a workaround, if the profileAvatar is not a base64 string, we ignore the avatar.
	if !strings.HasPrefix(profileAvatar, "data:") {
		return "", nil
	}

	profileAvatarStrs := strings.Split(profileAvatar, ",")
	b, err := base64.StdEncoding.DecodeString(profileAvatarStrs[len(profileAvatarStrs)-1])
	if err != nil {
		return "", err
	}
	if len(b) > 200*1024 {
		mimeType := strings.Split(mimetype.Detect(b).String(), ";")[0]

		var src image.Image
		switch mimeType {
		case "image/png":
			src, _ = png.Decode(bytes.NewReader(b))
		case "image/jpeg":
			src, _ = jpeg.Decode(bytes.NewReader(b))
		default:
			return "", status.Errorf(codes.InvalidArgument, "only support avatar image in jpeg and png formats")
		}

		// Set the expected size that you want:
		dst := image.NewRGBA(image.Rect(0, 0, 256, 256*src.Bounds().Max.Y/src.Bounds().Max.X))

		// Resize:
		draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

		var buf bytes.Buffer
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		err = encoder.Encode(bufio.NewWriter(&buf), dst)
		if err != nil {
			return "", status.Errorf(codes.InvalidArgument, "avatar image error")
		}
		profileAvatar = fmt.Sprintf("data:%s;base64,%s", "image/png", base64.StdEncoding.EncodeToString(buf.Bytes()))
	}
	return profileAvatar, nil
}

// DBUser2PBUser converts a database user instance to proto user
func (s *service) DBUser2PBUser(ctx context.Context, dbUser *datamodel.Owner) (*mgmtpb.User, error) {
	if dbUser == nil {
		return nil, status.Error(codes.Internal, "can't convert a nil user")
	}

	id := dbUser.ID
	uid := dbUser.UID.String()

	socialProfileLinks := map[string]string{}
	if dbUser.SocialProfileLinks != nil {
		b, _ := dbUser.SocialProfileLinks.MarshalJSON()
		_ = json.Unmarshal(b, &socialProfileLinks)
	}

	avatar := fmt.Sprintf("%s/v1beta/users/%s/avatar", s.instillCoreHost, dbUser.ID)
	return &mgmtpb.User{
		Name:       fmt.Sprintf("users/%s", id),
		Uid:        &uid,
		Id:         id,
		CreateTime: timestamppb.New(dbUser.CreateTime),
		UpdateTime: timestamppb.New(dbUser.UpdateTime),
		Profile: &mgmtpb.UserProfile{
			DisplayName:        &dbUser.DisplayName.String,
			CompanyName:        &dbUser.CompanyName.String,
			PublicEmail:        &dbUser.PublicEmail.String,
			Avatar:             &avatar,
			Bio:                &dbUser.Bio.String,
			SocialProfileLinks: socialProfileLinks,
		},
		Email: dbUser.Email,
	}, nil
}

// DBUser2PBAuthenticatedUser converts a database user instance to proto authenticated user
func (s *service) DBUser2PBAuthenticatedUser(ctx context.Context, dbUser *datamodel.Owner) (*mgmtpb.AuthenticatedUser, error) {
	if dbUser == nil {
		return nil, status.Error(codes.Internal, "can't convert a nil user")
	}

	id := dbUser.ID
	uid := dbUser.UID.String()
	socialProfileLinks := map[string]string{}
	if dbUser.SocialProfileLinks != nil {
		b, _ := dbUser.SocialProfileLinks.MarshalJSON()
		_ = json.Unmarshal(b, &socialProfileLinks)
	}

	avatar := fmt.Sprintf("%s/v1beta/users/%s/avatar", s.instillCoreHost, dbUser.ID)
	return &mgmtpb.AuthenticatedUser{
		Name:                   fmt.Sprintf("users/%s", id),
		Uid:                    &uid,
		Id:                     id,
		CreateTime:             timestamppb.New(dbUser.CreateTime),
		UpdateTime:             timestamppb.New(dbUser.UpdateTime),
		Email:                  dbUser.Email,
		CustomerId:             dbUser.CustomerID,
		Role:                   &dbUser.Role.String,
		CookieToken:            &dbUser.CookieToken.String,
		NewsletterSubscription: dbUser.NewsletterSubscription,
		Profile: &mgmtpb.UserProfile{
			DisplayName:        &dbUser.DisplayName.String,
			CompanyName:        &dbUser.CompanyName.String,
			PublicEmail:        &dbUser.PublicEmail.String,
			Avatar:             &avatar,
			Bio:                &dbUser.Bio.String,
			SocialProfileLinks: socialProfileLinks,
		},
		OnboardingStatus: mgmtpb.OnboardingStatus(dbUser.OnboardingStatus),
	}, nil
}

// PBAuthenticatedUser2DBUser converts a proto user instance to database user
func (s *service) PBAuthenticatedUser2DBUser(ctx context.Context, pbUser *mgmtpb.AuthenticatedUser) (*datamodel.Owner, error) {
	if pbUser == nil {
		return nil, status.Error(codes.Internal, "can't convert a nil user")
	}

	uid, err := uuid.FromString(pbUser.GetUid())
	if err != nil {
		return nil, err
	}

	userType := "user"
	email := pbUser.GetEmail()

	profileAvatar, err := s.compressAvatar(pbUser.GetProfile().GetAvatar())
	if err != nil {
		return nil, err
	}

	return &datamodel.Owner{
		Base: datamodel.Base{
			UID: uid,
		},
		ID: pbUser.GetId(),
		OwnerType: sql.NullString{
			String: userType,
			Valid:  len(userType) > 0,
		},
		Email:      email,
		CustomerID: pbUser.GetCustomerId(),
		DisplayName: sql.NullString{
			String: pbUser.GetProfile().GetDisplayName(),
			Valid:  len(pbUser.GetProfile().GetDisplayName()) > 0,
		},
		CompanyName: sql.NullString{
			String: pbUser.GetProfile().GetCompanyName(),
			Valid:  len(pbUser.GetProfile().GetCompanyName()) > 0,
		},
		PublicEmail: sql.NullString{
			String: pbUser.GetProfile().GetPublicEmail(),
			Valid:  len(pbUser.GetProfile().GetPublicEmail()) > 0,
		},
		Bio: sql.NullString{
			String: pbUser.GetProfile().GetBio(),
			Valid:  len(pbUser.GetProfile().GetBio()) > 0,
		},
		Role: sql.NullString{
			String: pbUser.GetRole(),
			Valid:  len(pbUser.GetRole()) > 0,
		},
		CookieToken: sql.NullString{
			String: pbUser.GetCookieToken(),
			Valid:  len(pbUser.GetCookieToken()) > 0,
		},
		NewsletterSubscription: pbUser.GetNewsletterSubscription(),
		ProfileAvatar: sql.NullString{
			String: profileAvatar,
			Valid:  len(profileAvatar) > 0,
		},
		SocialProfileLinks: func() []byte {
			if pbUser.GetProfile().GetSocialProfileLinks() != nil {
				b, err := json.Marshal(pbUser.GetProfile().GetSocialProfileLinks())
				if err != nil {
					return nil
				}
				return b
			}
			return []byte{}
		}(),
		OnboardingStatus: datamodel.OnboardingStatus(pbUser.OnboardingStatus),
	}, nil
}

func (s *service) DBUsers2PBUsers(ctx context.Context, dbUsers []*datamodel.Owner) ([]*mgmtpb.User, error) {
	var err error
	pbUsers := make([]*mgmtpb.User, len(dbUsers))
	for idx := range dbUsers {
		pbUsers[idx], err = s.DBUser2PBUser(
			ctx,
			dbUsers[idx],
		)
		if err != nil {
			return nil, err
		}

	}
	return pbUsers, nil
}

func (s *service) DBUsers2PBAuthenticatedUsers(ctx context.Context, dbUsers []*datamodel.Owner) ([]*mgmtpb.AuthenticatedUser, error) {
	var err error
	pbUsers := make([]*mgmtpb.AuthenticatedUser, len(dbUsers))
	for idx := range dbUsers {
		pbUsers[idx], err = s.DBUser2PBAuthenticatedUser(
			ctx,
			dbUsers[idx],
		)
		if err != nil {
			return nil, err
		}

	}
	return pbUsers, nil
}

// DBUser2PBUser converts a database user instance to proto user
func (s *service) DBOrg2PBOrg(ctx context.Context, dbOrg *datamodel.Owner) (*mgmtpb.Organization, error) {
	if dbOrg == nil {
		return nil, status.Error(codes.Internal, "can't convert a nil organization")
	}

	id := dbOrg.ID
	uid := dbOrg.UID.String()

	relations, err := s.aclClient.GetOrganizationUsers(ctx, dbOrg.UID)
	if err != nil {
		return nil, err
	}

	var owner *mgmtpb.User
	for _, relation := range relations {
		if relation.Relation == "owner" {
			owner, err = s.GetUserByUIDAdmin(ctx, relation.UID)
			if err != nil {
				return nil, err
			}
			break
		}
	}

	socialProfileLinks := map[string]string{}
	if dbOrg.SocialProfileLinks != nil {
		b, _ := dbOrg.SocialProfileLinks.MarshalJSON()
		_ = json.Unmarshal(b, &socialProfileLinks)
	}

	avatar := fmt.Sprintf("%s/v1beta/organizations/%s/avatar", s.instillCoreHost, dbOrg.ID)
	canUpdateOrganization := false
	ctxUserUID, err := s.ExtractCtxUser(ctx, true)
	if err == nil {
		canUpdateOrganization, err = s.aclClient.CheckOrganizationUserMembership(ctx, uuid.FromStringOrNil(uid), ctxUserUID, "can_update_organization")
		if err != nil {
			return nil, err
		}
	}
	users, err := s.aclClient.GetOrganizationUsers(ctx, dbOrg.UID)
	if err != nil {
		return nil, err
	}

	return &mgmtpb.Organization{
		Name:       fmt.Sprintf("organizations/%s", id),
		Uid:        uid,
		Id:         id,
		CreateTime: timestamppb.New(dbOrg.CreateTime),
		UpdateTime: timestamppb.New(dbOrg.UpdateTime),
		Profile: &mgmtpb.OrganizationProfile{
			DisplayName:        &dbOrg.DisplayName.String,
			PublicEmail:        &dbOrg.PublicEmail.String,
			Avatar:             &avatar,
			Bio:                &dbOrg.Bio.String,
			SocialProfileLinks: socialProfileLinks,
		},
		Owner: owner,
		Permission: &mgmtpb.Permission{
			CanEdit: canUpdateOrganization,
		},
		Stats: &mgmtpb.Organization_Stats{
			UserCount: int32(len(users)),
		},
	}, nil
}

// PBOrg2DBOrg converts a proto user instance to database user
func (s *service) PBOrg2DBOrg(ctx context.Context, pbOrg *mgmtpb.Organization) (*datamodel.Owner, error) {
	if pbOrg == nil {
		return nil, status.Error(codes.Internal, "can't convert a nil organization")
	}

	uid, err := uuid.FromString(pbOrg.GetUid())
	if err != nil {
		return nil, err
	}

	userType := "organization"
	profileAvatar, err := s.compressAvatar(pbOrg.GetProfile().GetAvatar())
	if err != nil {
		return nil, err
	}

	return &datamodel.Owner{
		Base: datamodel.Base{
			UID: uid,
		},
		ID: pbOrg.GetId(),
		OwnerType: sql.NullString{
			String: userType,
			Valid:  len(userType) > 0,
		},
		DisplayName: sql.NullString{
			String: pbOrg.GetProfile().GetDisplayName(),
			Valid:  len(pbOrg.GetProfile().GetDisplayName()) > 0,
		},

		PublicEmail: sql.NullString{
			String: pbOrg.GetProfile().GetPublicEmail(),
			Valid:  len(pbOrg.GetProfile().GetPublicEmail()) > 0,
		},
		Bio: sql.NullString{
			String: pbOrg.GetProfile().GetBio(),
			Valid:  len(pbOrg.GetProfile().GetBio()) > 0,
		},
		ProfileAvatar: sql.NullString{
			String: profileAvatar,
			Valid:  len(profileAvatar) > 0,
		},
		SocialProfileLinks: func() []byte {
			if pbOrg.GetProfile().GetSocialProfileLinks() != nil {
				b, err := json.Marshal(pbOrg.GetProfile().GetSocialProfileLinks())
				if err != nil {
					return nil
				}
				return b
			}
			return []byte{}
		}(),
	}, nil
}

func (s *service) DBOrgs2PBOrgs(ctx context.Context, dbOrgs []*datamodel.Owner) ([]*mgmtpb.Organization, error) {
	var err error
	pbOrgs := make([]*mgmtpb.Organization, len(dbOrgs))
	for idx := range dbOrgs {
		pbOrgs[idx], err = s.DBOrg2PBOrg(
			ctx,
			dbOrgs[idx],
		)
		if err != nil {
			return nil, err
		}

	}
	return pbOrgs, nil
}

// DBToken2PBToken converts a database user instance to proto user
func (s *service) DBToken2PBToken(ctx context.Context, dbToken *datamodel.Token) (*mgmtpb.ApiToken, error) {
	id := dbToken.ID
	state := mgmtpb.ApiToken_State(dbToken.State)
	if dbToken.ExpireTime.Before(time.Now()) {
		state = mgmtpb.ApiToken_State(mgmtpb.ApiToken_STATE_EXPIRED)
	}

	return &mgmtpb.ApiToken{
		Name:        fmt.Sprintf("tokens/%s", id),
		Uid:         dbToken.UID.String(),
		Id:          id,
		State:       state,
		AccessToken: dbToken.AccessToken,
		TokenType:   dbToken.TokenType,
		Expiration:  &mgmtpb.ApiToken_ExpireTime{ExpireTime: timestamppb.New(dbToken.ExpireTime)},
		CreateTime:  timestamppb.New(dbToken.CreateTime),
		UpdateTime:  timestamppb.New(dbToken.UpdateTime),
		LastUseTime: func() *timestamppb.Timestamp {
			if dbToken.LastUseTime.IsZero() {
				return nil
			}
			return timestamppb.New(dbToken.LastUseTime)
		}(),
	}, nil
}

// PBToken2DBToken converts a proto user instance to database user
func (s *service) PBToken2DBToken(ctx context.Context, pbToken *mgmtpb.ApiToken) (*datamodel.Token, error) {

	logger, _ := logx.GetZapLogger(ctx)

	r := &datamodel.Token{
		Base: datamodel.Base{
			UID: func() uuid.UUID {
				if pbToken.GetUid() == "" {
					return uuid.UUID{}
				}
				id, err := uuid.FromString(pbToken.GetUid())
				if err != nil {
					logger.Error(err.Error())
				}
				return id
			}(),

			CreateTime: func() time.Time {
				if pbToken.GetCreateTime() != nil {
					return pbToken.GetCreateTime().AsTime()
				}
				return time.Time{}
			}(),

			UpdateTime: func() time.Time {
				if pbToken.GetUpdateTime() != nil {
					return pbToken.GetUpdateTime().AsTime()
				}
				return time.Time{}
			}(),
		},
		ID:          pbToken.GetId(),
		State:       datamodel.TokenState(pbToken.GetState()),
		AccessToken: pbToken.AccessToken,
		TokenType:   pbToken.TokenType,
		ExpireTime:  pbToken.GetExpireTime().AsTime(),
	}
	return r, nil

}

func (s *service) DBTokens2PBTokens(ctx context.Context, dbTokens []*datamodel.Token) ([]*mgmtpb.ApiToken, error) {
	var err error
	pbUsers := make([]*mgmtpb.ApiToken, len(dbTokens))
	for idx := range dbTokens {
		pbUsers[idx], err = s.DBToken2PBToken(
			ctx,
			dbTokens[idx],
		)
		if err != nil {
			return nil, err
		}

	}
	return pbUsers, nil
}
