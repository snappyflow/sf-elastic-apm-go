package apmgoji

import (
	"fmt"
	"net/http"

	"github.com/zenazn/goji/web"

	"go.elastic.co/apm/module/apmhttp/v2"
	"go.elastic.co/apm/v2"
	"go.elastic.co/apm/v2/stacktrace"
)

func init() {
	stacktrace.RegisterLibraryPackage(
		"github.com/zenazn",
	)
}

// ServeConfig specifies the tracing configuration when using TraceAndServe.
type ServeConfig struct {
	Resource string
}

// Middleware returns a goji middleware function that will trace incoming requests.
func Middleware() func(*web.C, http.Handler) http.Handler {
	m := &middleware{
		tracer: apm.DefaultTracer(),
	}
	return func(c *web.C, h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resource := r.Method
			p := web.GetMatch(*c).RawPattern()
			if p != nil {
				resource += fmt.Sprintf(" %s", p)
			} else {
				p = r.URL.Path
				resource += fmt.Sprintf(" %s", p)
			}
			m.TraceAndServe(h, w, r, &ServeConfig{
				Resource: resource,
			})
		})
	}
}

type middleware struct {
	tracer *apm.Tracer
}

// TraceAndServe serves the handler h using the given ResponseWriter and Request, applying tracing
// according to the specified config.
func (m *middleware) TraceAndServe(h http.Handler, w http.ResponseWriter, r *http.Request, cfg *ServeConfig) {
	if cfg == nil {
		cfg = new(ServeConfig)
	}

	tx, body, req := apmhttp.StartTransactionWithBody(m.tracer, cfg.Resource, r)
	defer tx.End()
	rw, resp := apmhttp.WrapResponseWriter(w)
	defer func() {
		panicked := false
		if v := recover(); v != nil {
			w.WriteHeader(http.StatusInternalServerError)
			e := m.tracer.Recovered(v)
			e.SetTransaction(tx)
			setContext(&e.Context, req, http.StatusInternalServerError, body)
			e.Send()
			panicked = true
		}
		if panicked {
			resp.StatusCode = http.StatusInternalServerError
			apmhttp.SetTransactionContext(tx, req, resp, body)
		} else {
			apmhttp.SetTransactionContext(tx, req, resp, body)
		}
		body.Discard()
	}()
	h.ServeHTTP(rw, req)
	if resp.StatusCode == 0 {
		resp.StatusCode = http.StatusOK
	}
}

func setContext(ctx *apm.Context, req *http.Request, status int, body *apm.BodyCapturer) {
	ctx.SetFramework("goji", "")
	ctx.SetHTTPRequest(req)
	ctx.SetHTTPRequestBody(body)
	ctx.SetHTTPStatusCode(status)
	ctx.SetHTTPResponseHeaders(req.Header)
}
