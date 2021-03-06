package apmatreugo // import apmatreugo "github.com/atreugo/apm"

import (
	"github.com/savsgio/atreugo/v11"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm/module/apmfasthttp/v2"
	"go.elastic.co/apm/v2"
)

// New returns a factory instance.
func New(options ...Option) *Factory {
	f := new(Factory)

	for i := range options {
		options[i](f)
	}

	if f.tracer == nil {
		f.tracer = apm.DefaultTracer()
	}

	if f.requestName == nil {
		f.requestName = func(ctx *atreugo.RequestCtx) string {
			return apmfasthttp.ServerRequestName(ctx.RequestCtx)
		}
	}

	if f.requestIgnorer == nil {
		requestIgnorer := apmfasthttp.NewDynamicServerRequestIgnorer(f.tracer)
		f.requestIgnorer = func(ctx *atreugo.RequestCtx) bool {
			return requestIgnorer(ctx.RequestCtx)
		}
	}

	if f.recovery == nil {
		recovery := apmfasthttp.NewTraceRecovery(f.tracer)
		f.recovery = func(ctx *atreugo.RequestCtx, tx *apm.Transaction, bc *apm.BodyCapturer, recovered interface{}) {
			recovery(ctx.RequestCtx, tx, bc, recovered)
		}
	}

	return f
}

// Middleware returns a middleware.
func (f *Factory) Middleware() atreugo.Middleware {
	return func(ctx *atreugo.RequestCtx) error {
		if !f.tracer.Recording() || f.requestIgnorer(ctx) {
			return ctx.Next()
		}

		tx, _, err := apmfasthttp.StartTransactionWithBody(ctx.RequestCtx, f.tracer, f.requestName(ctx))
		if err != nil {
			return ctx.ErrorResponse(err, fasthttp.StatusInternalServerError)
		}

		tx.Context.SetFramework("atreugo", "v11")

		return ctx.Next()
	}
}

// PanicView returns a panic view.
func (f *Factory) PanicView() atreugo.PanicView {
	return func(ctx *atreugo.RequestCtx, err interface{}) {
		if f.panicPropagation {
			defer panic(err)
		}

		// 500 status code will be set only for APM transaction
		// to allow other middleware to choose a different response code
		if ctx.Response.Header.StatusCode() == fasthttp.StatusOK {
			ctx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
		}

		tx := apm.TransactionFromContext(ctx)
		bc := apm.BodyCapturerFromContext(ctx)

		f.recovery(ctx, tx, bc, err)
	}
}

// WithTracer returns a Option which sets t as the tracer
// to use for tracing server requests.
func WithTracer(t *apm.Tracer) Option {
	if t == nil {
		panic("t == nil")
	}

	return func(f *Factory) {
		f.tracer = t
	}
}

// WithServerRequestName returns a Option which sets fn as the function
// to use to obtain the transaction name for the given server request.
func WithServerRequestName(fn RequestNameFunc) Option {
	if fn == nil {
		panic("fn == nil")
	}

	return func(f *Factory) {
		f.requestName = fn
	}
}

// WithServerRequestIgnorer returns a Option which sets fn as the
// function to use to determine whether or not a server request should
// be ignored. If request ignorer is nil, all requests will be reported.
func WithServerRequestIgnorer(fn RequestIgnorerFunc) Option {
	if fn == nil {
		panic("fn == nil")
	}

	return func(f *Factory) {
		f.requestIgnorer = fn
	}
}

// WithRecovery returns a Option which sets r as the recovery
// function to use for tracing server requests.
func WithRecovery(r RecoveryFunc) Option {
	if r == nil {
		panic("r == nil")
	}

	return func(f *Factory) {
		f.recovery = r
	}
}

// WithPanicPropagation returns a Option which enable panic propagation.
// Any panic will be recovered and recorded as an error in a transaction, then
// panic will be caused again.
func WithPanicPropagation() Option {
	return func(f *Factory) {
		f.panicPropagation = true
	}
}
