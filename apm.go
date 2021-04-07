package apm

import (
	"github.com/savsgio/atreugo/v11"
	"go.elastic.co/apm"
)

func New(options ...Option) atreugo.Middleware {
	m := new(middleware)

	for i := range options {
		options[i](m)
	}

	if m.tracer == nil {
		m.tracer = apm.DefaultTracer
	}

	if m.requestName == nil {
		m.requestName = serverRequestName
	}

	if m.requestIgnorer == nil {
		m.requestIgnorer = newDynamicServerRequestIgnorer(m.tracer)
	}

	return m.handler
}

func (m *middleware) handler(ctx *atreugo.RequestCtx) error {
	if !m.tracer.Recording() || m.requestIgnorer(ctx) {
		return ctx.Next() // nolint:wrapcheck
	}

	tx, err := startTransaction(m.tracer, m.requestName(ctx), ctx)
	if err != nil {
		return err
	}

	ctx.SetUserValue(apmTxKey, tx)

	return ctx.Next() // nolint:wrapcheck
}

func NewRecoveryView() atreugo.PanicView {
	return func(ctx *atreugo.RequestCtx, err interface{}) {
		tx := GetTransaction(ctx)

		e := tx.tracer.Recovered(err)
		e.SetTransaction(tx.tx)

		_ = setResponseContext(tx)

		e.Send()
	}
}

// WithTracer returns a Option which sets t as the tracer
// to use for tracing server requests.
func WithTracer(t *apm.Tracer) Option {
	if t == nil {
		panic("t == nil")
	}

	return func(m *middleware) {
		m.tracer = t
	}
}

// WithServerRequestName returns a Option which sets fn as the function
// to use to obtain the transaction name for the given server request.
func WithServerRequestName(fn RequestNameFunc) Option {
	if fn == nil {
		panic("fn == nil")
	}

	return func(m *middleware) {
		m.requestName = fn
	}
}

// WithServerRequestIgnorer returns a Option which sets fn as the
// function to use to determine whether or not a server request should
// be ignored. If request ignorer is nil, all requests will be reported.
func WithServerRequestIgnorer(fn RequestIgnorerFunc) Option {
	if fn == nil {
		panic("fn == nil")
	}

	return func(m *middleware) {
		m.requestIgnorer = fn
	}
}
