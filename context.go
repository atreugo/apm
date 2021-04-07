package apm

import (
	"github.com/savsgio/atreugo/v11"
	"github.com/savsgio/gotils/strconv"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmhttp"
)

func setRequestContext(tx *Transaction, ctx *atreugo.RequestCtx, _ *apm.BodyCapturer) error {
	req := &tx.req

	if err := ctxToRequest(ctx, req); err != nil {
		return err
	}

	tx.tx.Context.SetHTTPRequest(req)
	// tx.tx.Context.SetHTTPRequestBody(body)

	return nil
}

func setResponseContext(tx *Transaction) error { // nolint:unparam
	statusCode := tx.httpCtx.Response.Header.StatusCode()

	tx.tx.Result = apmhttp.StatusCodeResult(statusCode)
	if !tx.tx.Sampled() {
		return nil
	}

	headers := tx.req.Header
	resetHTTPMap(headers)

	tx.httpCtx.Response.Header.VisitAll(func(k, v []byte) {
		sk := strconv.B2S(k)
		sv := strconv.B2S(v)

		headers.Set(sk, sv)
	})

	tx.tx.Context.SetHTTPResponseHeaders(headers)
	tx.tx.Context.SetHTTPStatusCode(statusCode)

	return nil
}
