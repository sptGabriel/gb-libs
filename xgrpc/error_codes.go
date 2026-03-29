package xgrpc

import (
	"github.com/sptGabriel/gb-libs/errs"
	"google.golang.org/grpc/codes"
)

func GetCodeAndReasonFromErr(errCode errs.ErrorCode) (codes.Code, string) {
	switch errCode {
	case errs.DomainError, errs.ValidationError:
		return codes.InvalidArgument, "VALIDATION_ERROR"
	case errs.NotFoundError:
		return codes.NotFound, "NOT_FOUND_ERROR"
	case errs.AlreadyExistsError:
		return codes.AlreadyExists, "ALREADY_EXISTS_ERROR"
	case errs.PermissionDeniedError:
		return codes.PermissionDenied, ""
	case errs.ExternalError, errs.DatabaseError, errs.InternalError:
		return codes.Internal, "INTERNAL_SERVER_ERROR"
	default:
		return codes.Internal, "INTERNAL_SERVER_ERROR"
	}
}
