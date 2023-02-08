package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/instill-ai/mgmt-backend/pkg/logger"
)

func HttpResponseModifier(ctx context.Context, w http.ResponseWriter, p proto.Message) error {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return nil
	}

	// set http status code
	if vals := md.HeaderMD.Get("x-http-code"); len(vals) > 0 {
		code, err := strconv.Atoi(vals[0])
		if err != nil {
			return err
		}
		// delete the headers to not expose any grpc-metadata in http response
		delete(md.HeaderMD, "x-http-code")
		delete(w.Header(), "Grpc-Metadata-X-Http-Code")
		delete(w.Header(), "Grpc-Metadata-Content-Type")
		delete(w.Header(), "Grpc-Metadata-Trailer")

		w.WriteHeader(code)
	}

	return nil
}

func ErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	logger, _ := logger.GetZapLogger()

	// return Internal when Marshal failed
	const fallback = `{"code": 13, "message": "failed to marshal error message"}`

	s := status.Convert(err)
	pb := s.Proto()

	w.Header().Del("Trailer")
	w.Header().Del("Transfer-Encoding")

	contentType := marshaler.ContentType(pb)
	if contentType == "application/json" {
		w.Header().Set("Content-Type", "application/problem+json")
	} else {
		w.Header().Set("Content-Type", contentType)
	}

	if s.Code() == codes.Unauthenticated {
		w.Header().Set("WWW-Authenticate", s.Message())
	}

	buf, err := marshaler.Marshal(pb)
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to marshal error message %q: %v", s, err))
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(w, fallback); err != nil {
			logger.Info(fmt.Sprintf("Failed to write response: %v", err))
		}
		return
	}

	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		logger.Info("Failed to extract ServerMetadata from context")
	}

	for k, vs := range md.HeaderMD {
		if h, ok := func(key string) (string, bool) {
			return fmt.Sprintf("%s%s", runtime.MetadataHeaderPrefix, key), true
		}(k); ok {
			for _, v := range vs {
				w.Header().Add(h, v)
			}
		}
	}

	// RFC 7230 https://tools.ietf.org/html/rfc7230#section-4.1.2
	// Unless the request includes a TE header field indicating "trailers"
	// is acceptable, as described in Section 4.3, a server SHOULD NOT
	// generate trailer fields that it believes are necessary for the user
	// agent to receive.
	doForwardTrailers := strings.Contains(strings.ToLower(r.Header.Get("TE")), "trailers")

	if doForwardTrailers {
		for k := range md.TrailerMD {
			tKey := textproto.CanonicalMIMEHeaderKey(fmt.Sprintf("%s%s", runtime.MetadataTrailerPrefix, k))
			w.Header().Add("Trailer", tKey)
		}
		w.Header().Set("Transfer-Encoding", "chunked")
	}

	var st int
	switch {
	case s.Code() == codes.FailedPrecondition && strings.Contains(s.Message(), "[DELETE]"):
		st = http.StatusUnprocessableEntity
	default:
		st = runtime.HTTPStatusFromCode(s.Code())
	}

	w.WriteHeader(st)
	if _, err := w.Write(buf); err != nil {
		logger.Info(fmt.Sprintf("Failed to write response: %v", err))
	}

	if doForwardTrailers {
		for k, vs := range md.TrailerMD {
			tKey := fmt.Sprintf("%s%s", runtime.MetadataTrailerPrefix, k)
			for _, v := range vs {
				w.Header().Add(tKey, v)
			}
		}
	}

}
