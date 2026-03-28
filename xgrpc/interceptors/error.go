package interceptors

import (
	"context"
	"maps"

	"github.com/sptGabriel/defaults/config"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorHandlerFunc func(error) error

// Intercepts and delegates error handling to a function errorHandler.
func UnaryErrorHandlingInterceptor(errorHandler ErrorHandlerFunc) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		resp, err = handler(ctx, req)
		if err == nil {
			return
		}

		err = errorHandler(err)
		return
	}
}

// Annotates an gRPC error with details and metadata.
func GetErrorDetails(
	message string,
	code codes.Code,
	reason string,
	config config.AppConfig,
	metadata *map[string]string,
) error {
	st := status.New(code, message)
	meta := map[string]string{
		"service": config.Name,
		"version": config.Version,
	}

	if metadata != nil {
		maps.Copy(meta, *metadata)
	}

	details, err := st.WithDetails(&epb.ErrorInfo{
		Reason:   reason,
		Domain:   config.Name,
		Metadata: meta,
	})
	if err != nil {
		return st.Err()
	}

	return details.Err()
}
