package apm

import (
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
)

func newTxCloser(ctx *fasthttp.RequestCtx, tx *apm.Transaction, bc *apm.BodyCapturer) *txCloser {
	return &txCloser{
		ctx: ctx,
		tx:  tx,
		bc:  bc,
	}
}

func (c *txCloser) Tx() *apm.Transaction {
	return c.tx
}

func (c *txCloser) BodyCapturer() *apm.BodyCapturer {
	return c.bc
}

func (c *txCloser) Close() error {
	if err := setResponseContext(c.ctx, c.tx, c.bc); err != nil {
		return err
	}

	c.tx.End()

	return nil
}
