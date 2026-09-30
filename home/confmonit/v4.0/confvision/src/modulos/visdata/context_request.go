package visdata

import (
	"context"
	"net/http"
)

type httpRequestKey struct{}

func ContextWithHTTPRequest(ctx context.Context, r *http.Request) context.Context {
	if r == nil {
		return ctx
	}
	return context.WithValue(ctx, httpRequestKey{}, r)
}

func HTTPRequestFromContext(ctx context.Context) *http.Request {
	if ctx == nil {
		return nil
	}
	r, _ := ctx.Value(httpRequestKey{}).(*http.Request)
	return r
}
