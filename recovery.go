package apm

import (
	"github.com/savsgio/atreugo/v11"
	"go.elastic.co/apm"
)

func newTraceRecovery(t *apm.Tracer) RecoveryFunc {
	if t == nil {
		t = apm.DefaultTracer
	}

	return func(ctx *atreugo.RequestCtx, tx *apm.Transaction, bc *apm.BodyCapturer, recovered interface{}) {
		_ = setResponseContext(ctx.RequestCtx, tx, bc)

		e := t.Recovered(recovered)
		e.SetTransaction(tx)
		e.Send()
	}
}
