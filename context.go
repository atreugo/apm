package apm

import (
	"net/http"
	"strings"

	"github.com/savsgio/atreugo/v11"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmhttp"
)

func TransactionFromRequestCtx(ctx *atreugo.RequestCtx) *apm.Transaction {
	if tx, ok := ctx.UserValue(txKey).(*txCloser); ok {
		return tx.Tx()
	}

	return nil
}

func setRequestContext(ctx *fasthttp.RequestCtx, tracer *apm.Tracer, tx *apm.Transaction) (*apm.BodyCapturer, error) {
	req := new(http.Request)
	if err := fasthttpadaptor.ConvertRequest(ctx, req, true); err != nil {
		return nil, err
	}

	bc := tracer.CaptureHTTPRequestBody(req)
	tx.Context.SetHTTPRequest(req)
	tx.Context.SetHTTPRequestBody(bc)

	return bc, nil
}

func setResponseContext(ctx *fasthttp.RequestCtx, tx *apm.Transaction, bc *apm.BodyCapturer) error {
	statusCode := ctx.Response.Header.StatusCode()

	tx.Result = apmhttp.StatusCodeResult(statusCode)
	if !tx.Sampled() {
		return nil
	}

	headers := make(http.Header)
	ctx.Response.Header.VisitAll(func(k, v []byte) {
		sk := string(k)
		sv := string(v)

		headers.Set(sk, sv)
	})

	tx.Context.SetHTTPResponseHeaders(headers)
	tx.Context.SetHTTPStatusCode(statusCode)

	if bc != nil {
		bc.Discard()
	}

	return nil
}

func startTransactionWithBody(
	tracer *apm.Tracer, name string, ctx *atreugo.RequestCtx,
) (*apm.Transaction, *apm.BodyCapturer, error) {
	traceContext, ok := getRequestTraceparent(ctx, apmhttp.W3CTraceparentHeader)
	if !ok {
		traceContext, ok = getRequestTraceparent(ctx, apmhttp.ElasticTraceparentHeader)
	}

	if ok {
		tracestateHeader := string(ctx.Request.Header.Peek(apmhttp.TracestateHeader))
		traceContext.State, _ = apmhttp.ParseTracestateHeader(strings.Split(tracestateHeader, ",")...)
	}

	tx := tracer.StartTransactionOptions(name, "request", apm.TransactionOptions{TraceContext: traceContext})

	bc, err := setRequestContext(ctx.RequestCtx, tracer, tx)
	if err != nil {
		tx.End()

		return nil, nil, err
	}

	return tx, bc, nil
}
