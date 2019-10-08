package router

import (
	"time"

	mw "bitbucket.org/teachingstrategies/go-svc-bootstrap/middlewares"
	"bitbucket.org/teachingstrategies/go-svc-bootstrap/tsjwt"
	newrelic "github.com/newrelic/go-agent"
	"github.com/rs/zerolog"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

// New creates and returns our base service router.
// This is our universal Middleware stack,
// which will ALWAYS run,
// regardless of subrouted path on the router.
// Middlewares added with this function:
//   * Heartbeat (/ping)
//   * Recoverer
//   * StripSlashes
//   * Timeout (600 seconds)
//   * AddContext (adds env key to context)
//   * AddProfiling (NewRelic profiling)
//   * RequestID
//   * AddLogging
//   * SetContentType (application/json)
func New(zlog *zerolog.Logger, apm newrelic.Application, env string, isLocal bool) *chi.Mux {

	r := chi.NewRouter()

	// Router Middleware

	// Uptime monitor endpoint
	// https://godoc.org/github.com/go-chi/chi/middleware#Heartbeat
	r.Use(middleware.Heartbeat("/ping"))
	// This recovers from panics, ensuring that logging/metric collection is not lost
	r.Use(mw.Recoverer)
	// This standardizes request paths
	r.Use(middleware.StripSlashes)
	// This gives us a base timeout for requests
	r.Use(middleware.Timeout(time.Second * time.Duration(600)))

	var ctxs = make(map[string]interface{})
	ctxs["env"] = env

	r.Use(mw.AddContext(ctxs))
	// Metrics via NewRelic Integration
	r.Use(mw.AddProfiling(apm))
	// Adds a request id to our context so that we
	// can piece together requests
	r.Use(middleware.RequestID)
	// Log the access
	r.Use(mw.AddLogging(zlog, isLocal))
	// This is a JSON API, thus set that content type for everything
	r.Use(render.SetContentType(render.ContentTypeJSON))

	return r
}

// AddJWTAuth creates a JWT Authenticator and if validates, add it to the context
func AddJWTAuth(r *chi.Mux, jwtVerifyKey string) *chi.Mux {
	// Create a JWT Authenticator based on our signature
	tokenAuth := tsjwt.New("HS256", []byte(jwtVerifyKey), nil)

	// JWT parser - we want this to always be attempted so that we
	// can enrich our logging with user info
	r.Use(tsjwt.Verifier(tokenAuth))

	return r
}
