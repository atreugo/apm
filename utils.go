package apm

import (
	"net/http"
	"net/url"

	"github.com/savsgio/atreugo/v11"
	"github.com/savsgio/gotils/strconv"
	"github.com/valyala/bytebufferpool"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmhttp"
)

func resetHTTPRequest(req *http.Request) {
	req.Method = ""
	req.URL = nil
	req.Proto = ""
	req.ProtoMajor = 0
	req.ProtoMinor = 0
	resetHTTPMap(req.Header)
	req.Body = nil
	req.GetBody = nil
	req.ContentLength = 0
	req.TransferEncoding = req.TransferEncoding[:0]
	req.Close = false
	req.Host = ""
	resetHTTPMap(req.Form)
	resetHTTPMap(req.PostForm)
	req.MultipartForm = nil
	resetHTTPMap(req.Trailer)
	req.RemoteAddr = ""
	req.RequestURI = ""
	req.TLS = nil
	// req.Cancel = nil
	req.Response = nil
}

func resetHTTPMap(m map[string][]string) {
	for k := range m {
		delete(m, k)
	}
}

func ctxToRequest(ctx *atreugo.RequestCtx, req *http.Request) error {
	body := ctx.PostBody()

	req.Method = strconv.B2S(ctx.Method())
	req.Proto = "HTTP/1.1"
	req.ProtoMajor = 1
	req.ProtoMinor = 1
	req.RequestURI = strconv.B2S(ctx.RequestURI())
	req.ContentLength = int64(len(body))
	req.Host = strconv.B2S(ctx.Host())
	req.RemoteAddr = ctx.RemoteAddr().String()
	req.TLS = ctx.TLSConnectionState()

	req.Header = make(http.Header)
	ctx.Request.Header.VisitAll(func(k, v []byte) {
		sk := strconv.B2S(k)
		sv := strconv.B2S(v)

		switch sk {
		case "Transfer-Encoding":
			req.TransferEncoding = append(req.TransferEncoding, sv)
		default:
			req.Header.Set(sk, sv)
		}
	})

	rURL, err := url.ParseRequestURI(req.RequestURI)
	if err != nil {
		return err
	}

	req.URL = rURL

	return nil
}

func newDynamicServerRequestIgnorer(t *apm.Tracer) RequestIgnorerFunc {
	return func(ctx *atreugo.RequestCtx) bool {
		uri := strconv.B2S(ctx.Request.URI().RequestURI())
		u, _ := url.ParseRequestURI(uri)

		return t.IgnoredTransactionURL(u)
	}
}

func serverRequestName(ctx *atreugo.RequestCtx) string {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)

	b.Write(ctx.Request.Header.Method()) // nolint:errcheck
	b.WriteByte(' ')                     // nolint:errcheck
	b.Write(ctx.URI().Path())            // nolint:errcheck

	return b.String()
}

func getRequestTraceparent(ctx *atreugo.RequestCtx, header string) (apm.TraceContext, bool) {
	if value := ctx.Request.Header.Peek(header); len(value) > 0 {
		if c, err := apmhttp.ParseTraceparentHeader(strconv.B2S(value)); err == nil {
			return c, true
		}
	}

	return apm.TraceContext{}, false
}
