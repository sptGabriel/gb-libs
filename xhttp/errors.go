package xhttp

import (
	"errors"

	"github.com/sptGabriel/gb-libs/errs"
)

type Type string

const (
	TypeServerError         Type = "srn:error:server_error"
	TypeBadRequest          Type = "srn:error:invalid_params"
	TypeNotFound            Type = "srn:error:resource_not_found"
	TypeConflict            Type = "srn:error:conflict"
	TypeForbidden           Type = "srn:error:forbidden"
	TypeUnauthorized        Type = "srn:error:unauthorized"
	TypeNotImplemented      Type = "srn:error:not_implemented"
	TypeUnprocessableEntity Type = "srn:error:unprocessable_entity"
)

type errorPayloadDetail struct {
	Reason   string `json:"reason"`
	Value    any    `json:"value,omitempty"`
	Property string `json:"property,omitempty"`
}

type ErrorPayload struct {
	Type    string               `json:"type" example:"srn:error:some_error"`
	Title   string               `json:"title,omitempty" example:"Message for some error"`
	Details []errorPayloadDetail `json:"details,omitempty" swaggertype:"object"`
}

func (e ErrorPayload) Detail(reason, value, property string) ErrorPayload {
	e.Details = append(e.Details, errorPayloadDetail{Reason: reason, Value: value, Property: property})
	return e
}

func NewErrorPayload(_type Type, title string) ErrorPayload {
	return ErrorPayload{
		Type:  string(_type),
		Title: title,
	}
}

var InternalServerErrorPayload = NewErrorPayload(TypeServerError, "Internal Server Error")

// FromError translates errors coming from the application core into HTTP
// responses using the errs error codes. Codes not meant for clients
// (internal, database, external) get a generic 500 payload.
func FromError(err error) Response {
	var appErr *errs.Err
	if !errors.As(err, &appErr) {
		return InternalServerError(err)
	}

	switch appErr.Code {
	case errs.ValidationError:
		return BadRequest(err, NewErrorPayload(TypeBadRequest, detail(appErr)))
	case errs.DomainError:
		return UnprocessableEntity(err, NewErrorPayload(TypeUnprocessableEntity, detail(appErr)))
	case errs.NotFoundError:
		return NotFound(err, NewErrorPayload(TypeNotFound, detail(appErr)))
	case errs.AlreadyExistsError:
		return Conflict(err, NewErrorPayload(TypeConflict, detail(appErr)))
	case errs.UnauthenticatedError:
		return Unauthorized(err, NewErrorPayload(TypeUnauthorized, "authentication required"))
	case errs.PermissionDeniedError:
		return Forbidden(err, NewErrorPayload(TypeForbidden, "permission denied"))
	default:
		return InternalServerError(err)
	}
}

// detail exposes the underlying cause; only used for codes that are safe to
// return to clients
func detail(err *errs.Err) string {
	if err.Err != nil {
		return err.Err.Error()
	}
	return err.Message
}
