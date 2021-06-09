package apmatreugo // import apmatreugo "github.com/atreugo/apm"

import (
	"github.com/savsgio/atreugo/v11"

	"go.elastic.co/apm"
)

// Factory is a factory to create the tracing middleware and panic view.
type Factory struct {
	tracer           *apm.Tracer
	requestName      RequestNameFunc
	requestIgnorer   RequestIgnorerFunc
	recovery         RecoveryFunc
	panicPropagation bool
}

// Option sets options for tracing requests.
type Option func(*Factory)

// RequestNameFunc is the type of a function for use in
// WithServerRequestName.
type RequestNameFunc func(*atreugo.RequestCtx) string

// RequestIgnorerFunc is the type of a function for use in
// WithServerRequestIgnorer.
type RequestIgnorerFunc func(*atreugo.RequestCtx) bool

// RecoveryFunc is the type of a function for use in WithRecovery.
type RecoveryFunc func(ctx *atreugo.RequestCtx, tx *apm.Transaction, bc *apm.BodyCapturer, recovered interface{})
