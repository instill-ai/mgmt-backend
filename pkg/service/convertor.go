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
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofrs/uuid"
	"golang.org/x/image/draw"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/logger"

	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
)

// maps for user owner type
var (
	PBUserType2DBUserType = map[mgmtPB.OwnerType]string{
		mgmtPB.OwnerType_OWNER_TYPE_UNSPECIFIED:  "unspecified",
		mgmtPB.OwnerType_OWNER_TYPE_USER:         "user",
		mgmtPB.OwnerType_OWNER_TYPE_ORGANIZATION: "organization",
	}
	DBUserType2PBUserType = map[string]mgmtPB.OwnerType{
		"unspecified":  mgmtPB.OwnerType_OWNER_TYPE_UNSPECIFIED,
		"user":         mgmtPB.OwnerType_OWNER_TYPE_USER,
		"organization": mgmtPB.OwnerType_OWNER_TYPE_ORGANIZATION,
	}
)

func (s *service) compressAvatar(profileAvatar string) (string, error) {
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
func (s *service) DBUser2PBUser(ctx context.Context, dbUser *datamodel.Owner) (*mgmtPB.User, error) {
	if dbUser == nil {
		return nil, status.Error(codes.Internal, "can't convert a nil user")
	}

	id := dbUser.ID
	uid := dbUser.Base.UID.String()

	socialProfileLinks := map[string]string{}
	if dbUser.SocialProfileLinks != nil {
		b, _ := dbUser.SocialProfileLinks.MarshalJSON()
		_ = json.Unmarshal(b, &socialProfileLinks)
	}

	return &mgmtPB.User{
		Name:       fmt.Sprintf("users/%s", id),
		Uid:        &uid,
		Id:         id,
		CreateTime: timestamppb.New(dbUser.Base.CreateTime),
		UpdateTime: timestamppb.New(dbUser.Base.UpdateTime),
		Profile: &mgmtPB.UserProfile{
			DisplayName:        &dbUser.DisplayName.String,
			CompanyName:        &dbUser.CompanyName.String,
			PublicEmail:        &dbUser.PublicEmail.String,
			Avatar:             &dbUser.ProfileAvatar.String,
			Bio:                &dbUser.Bio.String,
			SocialProfileLinks: socialProfileLinks,
		},
	}, nil
}

// DBUser2PBAuthenticatedUser converts a database user instance to proto authenticated user
func (s *service) DBUser2PBAuthenticatedUser(ctx context.Context, dbUser *datamodel.Owner) (*mgmtPB.AuthenticatedUser, error) {
	if dbUser == nil {
		return nil, status.Error(codes.Internal, "can't convert a nil user")
	}

	id := dbUser.ID
	uid := dbUser.Base.UID.String()
	socialProfileLinks := map[string]string{}
	if dbUser.SocialProfileLinks != nil {
		b, _ := dbUser.SocialProfileLinks.MarshalJSON()
		_ = json.Unmarshal(b, &socialProfileLinks)
	}

	return &mgmtPB.AuthenticatedUser{
		Name:                   fmt.Sprintf("users/%s", id),
		Uid:                    &uid,
		Id:                     id,
		CreateTime:             timestamppb.New(dbUser.Base.CreateTime),
		UpdateTime:             timestamppb.New(dbUser.Base.UpdateTime),
		Email:                  dbUser.Email,
		CustomerId:             dbUser.CustomerID,
		Role:                   &dbUser.Role.String,
		CookieToken:            &dbUser.CookieToken.String,
		NewsletterSubscription: dbUser.NewsletterSubscription,
		Profile: &mgmtPB.UserProfile{
			DisplayName:        &dbUser.DisplayName.String,
			CompanyName:        &dbUser.CompanyName.String,
			PublicEmail:        &dbUser.PublicEmail.String,
			Avatar:             &dbUser.ProfileAvatar.String,
			Bio:                &dbUser.Bio.String,
			SocialProfileLinks: socialProfileLinks,
		},
		OnboardingStatus: mgmtPB.OnboardingStatus(dbUser.OnboardingStatus),
	}, nil
}

// PBAuthenticatedUser2DBUser converts a proto user instance to database user
func (s *service) PBAuthenticatedUser2DBUser(ctx context.Context, pbUser *mgmtPB.AuthenticatedUser) (*datamodel.Owner, error) {
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

func (s *service) DBUsers2PBUsers(ctx context.Context, dbUsers []*datamodel.Owner) ([]*mgmtPB.User, error) {
	var err error
	pbUsers := make([]*mgmtPB.User, len(dbUsers))
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

func (s *service) DBUsers2PBAuthenticatedUsers(ctx context.Context, dbUsers []*datamodel.Owner) ([]*mgmtPB.AuthenticatedUser, error) {
	var err error
	pbUsers := make([]*mgmtPB.AuthenticatedUser, len(dbUsers))
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
func (s *service) DBOrg2PBOrg(ctx context.Context, dbOrg *datamodel.Owner) (*mgmtPB.Organization, error) {
	if dbOrg == nil {
		return nil, status.Error(codes.Internal, "can't convert a nil organization")
	}

	id := dbOrg.ID
	uid := dbOrg.Base.UID.String()

	relations, err := s.aclClient.GetOrganizationUsers(ctx, dbOrg.Base.UID)
	if err != nil {
		return nil, err
	}

	var owner *mgmtPB.User
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

	return &mgmtPB.Organization{
		Name:       fmt.Sprintf("organizations/%s", id),
		Uid:        uid,
		Id:         id,
		CreateTime: timestamppb.New(dbOrg.Base.CreateTime),
		UpdateTime: timestamppb.New(dbOrg.Base.UpdateTime),
		Profile: &mgmtPB.OrganizationProfile{
			DisplayName:        &dbOrg.DisplayName.String,
			PublicEmail:        &dbOrg.PublicEmail.String,
			Avatar:             &dbOrg.ProfileAvatar.String,
			Bio:                &dbOrg.Bio.String,
			SocialProfileLinks: socialProfileLinks,
		},
		Owner: owner,
	}, nil
}

// PBOrg2DBOrg converts a proto user instance to database user
func (s *service) PBOrg2DBOrg(ctx context.Context, pbOrg *mgmtPB.Organization) (*datamodel.Owner, error) {
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

func (s *service) DBOrgs2PBOrgs(ctx context.Context, dbOrgs []*datamodel.Owner) ([]*mgmtPB.Organization, error) {
	var err error
	pbOrgs := make([]*mgmtPB.Organization, len(dbOrgs))
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
func (s *service) DBToken2PBToken(ctx context.Context, dbToken *datamodel.Token) (*mgmtPB.ApiToken, error) {
	id := dbToken.ID
	state := mgmtPB.ApiToken_State(dbToken.State)
	if dbToken.ExpireTime.Before(time.Now()) {
		state = mgmtPB.ApiToken_State(mgmtPB.ApiToken_STATE_EXPIRED)
	}

	return &mgmtPB.ApiToken{
		Name:        fmt.Sprintf("tokens/%s", id),
		Uid:         dbToken.Base.UID.String(),
		Id:          id,
		State:       state,
		AccessToken: dbToken.AccessToken,
		TokenType:   dbToken.TokenType,
		Expiration:  &mgmtPB.ApiToken_ExpireTime{ExpireTime: timestamppb.New(dbToken.ExpireTime)},
		CreateTime:  timestamppb.New(dbToken.Base.CreateTime),
		UpdateTime:  timestamppb.New(dbToken.Base.UpdateTime),
	}, nil
}

// PBToken2DBToken converts a proto user instance to database user
func (s *service) PBToken2DBToken(ctx context.Context, pbToken *mgmtPB.ApiToken) (*datamodel.Token, error) {

	logger, _ := logger.GetZapLogger(ctx)

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

func (s *service) DBTokens2PBTokens(ctx context.Context, dbTokens []*datamodel.Token) ([]*mgmtPB.ApiToken, error) {
	var err error
	pbUsers := make([]*mgmtPB.ApiToken, len(dbTokens))
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
