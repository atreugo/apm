package apm

import (
	"net/http"

	"github.com/savsgio/atreugo/v11"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
)

type middleware struct {
	tracer         *apm.Tracer
	requestName    RequestNameFunc
	requestIgnorer RequestIgnorerFunc
}

// Option sets options for tracing server requests.
type Option func(*middleware)

// RequestNameFunc is the type of a function for use in
// WithServerRequestName.
type RequestNameFunc func(*atreugo.RequestCtx) string

// RequestIgnorerFunc is the type of a function for use in
// WithServerRequestIgnorer.
type RequestIgnorerFunc func(*atreugo.RequestCtx) bool

type Transaction struct {
	tracer    *apm.Tracer
	tx        *apm.Transaction
	req       http.Request
	httpCtx   *fasthttp.RequestCtx
	manualEnd bool
}
