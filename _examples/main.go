package main

import (
	"bufio"
	"time"

	"github.com/atreugo/apm"
	"github.com/savsgio/atreugo/v11"
)

func main() {
	apmMiddleware := apm.New()

	s := atreugo.New(atreugo.Config{
		Addr:      ":8000",
		PanicView: apmMiddleware.PanicView(),
	})

	s.UseBefore(apmMiddleware.Middleware())

	s.GET("/", func(ctx *atreugo.RequestCtx) error {
		panic("holaa")

		ctx.SetBodyStreamWriter(func(w *bufio.Writer) {
			w.WriteString("Hello")
			w.WriteByte('\n')

			time.Sleep(5 * time.Second)

			w.WriteString("World")
			w.WriteByte('\n')
		})

		return nil
	})

	s.ListenAndServe()
}
