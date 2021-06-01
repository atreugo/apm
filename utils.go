package apm

import (
	"net/url"

	"github.com/savsgio/atreugo/v11"
	"github.com/valyala/bytebufferpool"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmhttp"
)

func getRequestTraceparent(ctx *atreugo.RequestCtx, header string) (apm.TraceContext, bool) {
	if value := ctx.Request.Header.Peek(header); len(value) > 0 {
		if c, err := apmhttp.ParseTraceparentHeader(string(value)); err == nil {
			return c, true
		}
	}

	return apm.TraceContext{}, false
}

func newDynamicServerRequestIgnorer(t *apm.Tracer) RequestIgnorerFunc {
	return func(ctx *atreugo.RequestCtx) bool {
		uri := string(ctx.Request.URI().RequestURI())

		u, err := url.ParseRequestURI(uri)
		if err != nil {
			return true
		}

		return t.IgnoredTransactionURL(u)
	}
}

func serverRequestName(ctx *atreugo.RequestCtx) string {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)

	b.Write(ctx.Request.Header.Method()) // nolint:errcheck
	b.WriteByte(' ')                     // nolint:errcheck
	b.Write(ctx.Request.URI().Path())    // nolint:errcheck

	return b.String()
}
