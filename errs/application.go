package errs

import "errors"

type ErrorCode int

const (
	InternalError ErrorCode = iota
	DomainError
	NotFoundError
	DatabaseError
	ExternalError
	ValidationError
	AlreadyExistsError
	UnauthenticatedError
	PermissionDeniedError
)

const (
	msgInternalError      = "Internal"
	msgResourceNotFound   = "ResourceNotFound"
	msgValidationResource = "ValidationResource"
	msgDuplicatedResource = "DuplicatedResource"
)

type Err struct {
	Err     error
	Message string
	Code    ErrorCode
	Context map[string]any
}

func (e Err) Error() string {
	return e.Message
}

func (e Err) Unwrap() error {
	return e.Err
}

func (e Err) Is(err error) bool {
	other, ok := err.(Err)
	if !ok {
		return false
	}

	if e.Err == nil || other.Err == nil {
		return other.Code == e.Code
	}

	return e.Code == other.Code && errors.Is(e.Err, other.Err)
}

func (e *Err) WithContext(ctx map[string]any) {
	e.Context = ctx
}

func WrapUnauthenticatedError(msg string, err error) *Err {
	return &Err{
		Code:    UnauthenticatedError,
		Err:     err,
		Message: msg,
	}
}

func WrapPermissionDeniedError(msg string, err error) *Err {
	return &Err{
		Code:    PermissionDeniedError,
		Err:     err,
		Message: msg,
	}
}

func WrapDomainError(msg string, err error) *Err {
	return &Err{
		Code:    DomainError,
		Err:     err,
		Message: msg,
	}
}

func WrapError(err error) *Err {
	return &Err{
		Code:    InternalError,
		Err:     err,
		Message: msgInternalError,
	}
}

func WrapDuplicatedResourceError(resource, field string, err error) *Err {
	return &Err{
		Code:    AlreadyExistsError,
		Message: msgDuplicatedResource + resource,
		Err:     err,
		Context: map[string]any{
			"Field": field,
		},
	}
}

func WrapResourceNotFoundError(err error) *Err {
	return &Err{
		Code:    NotFoundError,
		Err:     err,
		Message: msgResourceNotFound,
	}
}

func WrapValidationError(resource string, err error) *Err {
	return &Err{
		Code:    ValidationError,
		Err:     err,
		Message: msgValidationResource + resource,
	}
}

func WrapExternalError(msg string, err error) *Err {
	return &Err{
		Err:     err,
		Code:    ExternalError,
		Message: msg,
	}
}
