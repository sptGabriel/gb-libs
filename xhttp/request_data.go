package xhttp

import "context"

type requestDataCtxKey struct{}

type RequestInfo struct {
	IP        string
	Port      string
	RequestID string
}

// WithRequestData attaches request metadata to a context
func WithRequestData(ctx context.Context, data RequestInfo) context.Context {
	return context.WithValue(ctx, requestDataCtxKey{}, data)
}

// RequestDataFromContext returns the request metadata previously attached
// by the RequestData middleware
func RequestDataFromContext(ctx context.Context) (RequestInfo, bool) {
	data, ok := ctx.Value(requestDataCtxKey{}).(RequestInfo)
	return data, ok
}
