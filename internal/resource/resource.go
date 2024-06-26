package resource

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc/metadata"
)

// ExtractFromMetadata extracts context metadata given a key
func ExtractFromMetadata(ctx context.Context, key string) ([]string, bool) {
	data, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return []string{}, false
	}
	return data[strings.ToLower(key)], true
}

// GetRequestSingleHeader get a request header, the header has to be single-value HTTP header
func GetRequestSingleHeader(ctx context.Context, header string) string {
	metaHeader := metadata.ValueFromIncomingContext(ctx, strings.ToLower(header))
	if len(metaHeader) != 1 {
		return ""
	}
	return metaHeader[0]
}

// GetRscNameID returns the resource ID given a resource name
func GetRscNameID(path string) (string, error) {
	id := path[strings.LastIndex(path, "/")+1:]
	if id == "" {
		return "", fmt.Errorf("error when extract resource id from resource name '%s'", path)
	}
	return id, nil
}

// GetRscPermalinkUID returns the resource UID given a resource permalink
func GetRscPermalinkUID(path string) (uuid.UUID, error) {

	splits := strings.Split(path, "/")
	if len(splits) < 2 {
		return uuid.Nil, fmt.Errorf("error when extract resource id from resource permalink '%s'", path)
	}

	return uuid.FromStringOrNil(splits[1]), nil
}

func UserUIDToUserPermalink(userUID uuid.UUID) string {
	return fmt.Sprintf("users/%s", userUID.String())
}
