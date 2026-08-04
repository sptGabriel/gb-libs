package xhttp

import "net/http"

type Response struct {
	Status int
	Error  error
	Body   any
	header http.Header
}

func NewResponse(status int) Response {
	return Response{
		Status: status,
		header: http.Header{},
	}
}

func (r Response) WithBody(body any) Response {
	r.Body = body
	return r
}

func (r Response) WithError(err error) Response {
	r.Error = err
	return r
}

func (r Response) AddHeader(name, value string) Response {
	if r.header == nil {
		r.header = http.Header{}
	}
	r.header.Add(name, value)
	return r
}

func (r Response) SetHeader(name, value string) Response {
	if r.header == nil {
		r.header = http.Header{}
	}
	r.header.Set(name, value)
	return r
}

func (r Response) Header() http.Header {
	return r.header
}

func OK(body any) Response {
	return NewResponse(http.StatusOK).WithBody(body)
}

func Created(body any) Response {
	return NewResponse(http.StatusCreated).WithBody(body)
}

func NoContent() Response {
	return NewResponse(http.StatusNoContent)
}

func BadRequest(err error, body any) Response {
	return NewResponse(http.StatusBadRequest).WithError(err).WithBody(body)
}

func Unauthorized(err error, body any) Response {
	return NewResponse(http.StatusUnauthorized).WithError(err).WithBody(body)
}

func Forbidden(err error, body any) Response {
	return NewResponse(http.StatusForbidden).WithError(err).WithBody(body)
}

func NotFound(err error, body any) Response {
	return NewResponse(http.StatusNotFound).WithError(err).WithBody(body)
}

func Conflict(err error, body any) Response {
	return NewResponse(http.StatusConflict).WithError(err).WithBody(body)
}

func UnprocessableEntity(err error, body any) Response {
	return NewResponse(http.StatusUnprocessableEntity).WithError(err).WithBody(body)
}

func InternalServerError(err error) Response {
	return NewResponse(http.StatusInternalServerError).
		WithError(err).
		WithBody(InternalServerErrorPayload)
}
