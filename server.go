package apm

import (
	"github.com/savsgio/atreugo/v11"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
)

func New(options ...Option) *APM {
	a := new(APM)

	for i := range options {
		options[i](a)
	}

	if a.tracer == nil {
		a.tracer = apm.DefaultTracer
	}

	if a.requestName == nil {
		a.requestName = serverRequestName
	}

	if a.requestIgnorer == nil {
		a.requestIgnorer = newDynamicServerRequestIgnorer(a.tracer)
	}

	if a.recovery == nil {
		a.recovery = newTraceRecovery(a.tracer)
	}

	return a
}

func (a *APM) Middleware() atreugo.Middleware {
	return func(ctx *atreugo.RequestCtx) error {
		if !a.tracer.Recording() || a.requestIgnorer(ctx) {
			return ctx.Next()
		}

		tx, bc, err := startTransactionWithBody(a.tracer, a.requestName(ctx), ctx)
		if err != nil {
			return ctx.ErrorResponse(err, fasthttp.StatusInternalServerError)
		}

		ctx.SetUserValue(txKey, newTxCloser(ctx.RequestCtx, tx, bc))

		return ctx.Next()
	}
}

func (a *APM) PanicView() atreugo.PanicView {
	return func(ctx *atreugo.RequestCtx, err interface{}) {
		if a.panicPropagation {
			defer panic(err)
		}

		// 500 status code will be set only for APM transaction
		// to allow other middleware to choose a different response code
		if ctx.Response.Header.StatusCode() == fasthttp.StatusOK {
			ctx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
		}

		if tx, ok := ctx.UserValue(txKey).(*txCloser); ok {
			a.recovery(ctx, tx.Tx(), tx.BodyCapturer(), err)
		}
	}
}

// WithTracer returns a Option which sets t as the tracer
// to use for tracing server requests.
func WithTracer(t *apm.Tracer) Option {
	if t == nil {
		panic("t == nil")
	}

	return func(a *APM) {
		a.tracer = t
	}
}

// WithServerRequestName returns a Option which sets fn as the function
// to use to obtain the transaction name for the given server request.
func WithServerRequestName(fn RequestNameFunc) Option {
	if fn == nil {
		panic("fn == nil")
	}

	return func(a *APM) {
		a.requestName = fn
	}
}

// WithServerRequestIgnorer returns a Option which sets fn as the
// function to use to determine whether or not a server request should
// be ignored. If request ignorer is nil, all requests will be reported.
func WithServerRequestIgnorer(fn RequestIgnorerFunc) Option {
	if fn == nil {
		panic("fn == nil")
	}

	return func(a *APM) {
		a.requestIgnorer = fn
	}
}

// WithRecovery returns a Option which sets r as the recovery
// function to use for tracing server requests.
func WithRecovery(r RecoveryFunc) Option {
	if r == nil {
		panic("r == nil")
	}

	return func(a *APM) {
		a.recovery = r
	}
}

// WithPanicPropagation returns a Option which enable panic propagation.
// Any panic will be recovered and recorded as an error in a transaction, then
// panic will be caused again.
func WithPanicPropagation() Option {
	return func(a *APM) {
		a.panicPropagation = true
	}
}
