package middleware

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/instill-ai/mgmt-backend/pkg/repository"
)

type fn func(*runtime.ServeMux, repository.Repository, http.ResponseWriter, *http.Request, map[string]string)

// AppendCustomHeaderMiddleware appends custom headers
func AppendCustomHeaderMiddleware(mux *runtime.ServeMux, repository repository.Repository, next fn) runtime.HandlerFunc {
	return runtime.HandlerFunc(func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		next(mux, repository, w, r, pathParams)
	})
}

func HandleAvatar(mux *runtime.ServeMux, repository repository.Repository, w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
	ctx, cancel := context.WithTimeout(req.Context(), 10*time.Second)
	defer cancel()
	if v, ok := pathParams["name"]; !ok || len(strings.Split(v, "/")) < 2 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var avatarBase64 string
	nameParts := strings.Split(pathParams["name"], "/")
	switch nameParts[0] {
	case "users":
		user, err := repository.GetUser(ctx, nameParts[1], true)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// Check if avatar is NULL or empty
		if !user.ProfileAvatar.Valid || user.ProfileAvatar.String == "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		avatarBase64 = user.ProfileAvatar.String
	case "organizations":
		org, err := repository.GetOrganization(ctx, nameParts[1], true)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// Check if avatar is NULL or empty
		if !org.ProfileAvatar.Valid || org.ProfileAvatar.String == "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		avatarBase64 = org.ProfileAvatar.String
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Avatar data should be in format "data:image/...;base64,<data>"
	parts := strings.Split(avatarBase64, ",")
	if len(parts) < 2 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	b, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = w.Write(b)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}
