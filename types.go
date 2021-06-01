package apm

import (
	"github.com/savsgio/atreugo/v11"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
)

type APM struct {
	tracer           *apm.Tracer
	requestName      RequestNameFunc
	requestIgnorer   RequestIgnorerFunc
	recovery         RecoveryFunc
	panicPropagation bool
}

type txCloser struct {
	ctx *fasthttp.RequestCtx
	tx  *apm.Transaction
	bc  *apm.BodyCapturer
}

// Option sets options for tracing server requests.
type Option func(*APM)

// RequestNameFunc is the type of a function for use in
// WithServerRequestName.
type RequestNameFunc func(*atreugo.RequestCtx) string

// RequestIgnorerFunc is the type of a function for use in
// WithServerRequestIgnorer.
type RequestIgnorerFunc func(*atreugo.RequestCtx) bool

// RecoveryFunc is the type of a function for use in WithRecovery.
type RecoveryFunc func(ctx *atreugo.RequestCtx, tx *apm.Transaction, bc *apm.BodyCapturer, recovered interface{})
