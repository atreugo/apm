package apm

import (
	"fmt"
	"strings"
	"sync"

	"github.com/savsgio/atreugo/v11"
	"github.com/savsgio/gotils/bytes"
	"github.com/savsgio/gotils/strconv"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmhttp"
)

var (
	apmTxKey = fmt.Sprintf("__apmTxKey::%s__", bytes.Rand(make([]byte, 5)))

	transactionPool = sync.Pool{
		New: func() interface{} {
			return new(Transaction)
		},
	}
)

func acquireTransaction() *Transaction {
	return transactionPool.Get().(*Transaction)
}

func releaseTransaction(tx *Transaction) {
	tx.reset()
	transactionPool.Put(tx)
}

func (tx *Transaction) reset() {
	tx.tracer = nil
	tx.tx = nil
	resetHTTPRequest(&tx.req)
	tx.httpCtx = nil
	tx.manualEnd = false
}

func (tx *Transaction) AutoEnd(v bool) {
	tx.manualEnd = !v
}

func (tx *Transaction) Tx() *apm.Transaction {
	return tx.tx
}

func (tx *Transaction) End() {
	tx.tx.End()
	releaseTransaction(tx)
}

func (tx *Transaction) Close() error {
	if err := setResponseContext(tx); err != nil {
		return err
	}

	if !tx.manualEnd {
		tx.End()
	}

	return nil
}

func startTransaction(tracer *apm.Tracer, name string, ctx *atreugo.RequestCtx) (*Transaction, error) {
	traceContext, ok := getRequestTraceparent(ctx, apmhttp.ElasticTraceparentHeader)
	if !ok {
		traceContext, ok = getRequestTraceparent(ctx, apmhttp.W3CTraceparentHeader)
	}

	if ok {
		tracestateHeader := strconv.B2S(ctx.Request.Header.Peek(apmhttp.TracestateHeader))
		traceContext.State, _ = apmhttp.ParseTracestateHeader(strings.Split(tracestateHeader, ",")...)
	}

	tx := acquireTransaction()
	tx.tracer = tracer
	tx.tx = tracer.StartTransactionOptions(name, "request", apm.TransactionOptions{TraceContext: traceContext})
	tx.httpCtx = ctx.RequestCtx

	if err := setRequestContext(tx, ctx, nil); err != nil {
		tx.Close()

		return nil, err
	}

	return tx, nil
}

func GetTransaction(ctx *atreugo.RequestCtx) *Transaction {
	if tx, ok := ctx.UserValue(apmTxKey).(*Transaction); ok {
		return tx
	}

	return nil
}
