package middleware

import (
	"context"
	"errors"
	"net/http"

	"bitbucket.org/teachingstrategies/go-svc-bootstrap/utils"
	newrelic "github.com/newrelic/go-agent"
	"github.com/rs/zerolog"
)

// AddProfiling automatically adds timing, and recovery error notices to
// wrapped routes
func AddProfiling(apm newrelic.Application) func(next http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {

			// Begin our transaction.  As of this point, we are tracking the execution time
			// of the request.  The transaction is injected in to the context so that
			// functions within the processing flow may segment the transaction as warranted
			txn := apm.StartTransaction(r.URL.Path, w, r)
			augmentedContext := context.WithValue(r.Context(), utils.ContextKey("txn"), txn)

			defer func(txn newrelic.Transaction) {

				// If this was a panic, add telemetry.
				if e := recover(); e != nil {

					txn.NoticeError(e.(error))
					txn.End()

					// propagate the panic
					panic(e)

				}
				txn.End()
			}(txn)

			next.ServeHTTP(w, r.WithContext(augmentedContext))
		}
		return http.HandlerFunc(fn)
	}
}

// NoticeErrorHook allows us to publish errors to our APM of choice
// In this case NewRelic
type NoticeErrorHook struct {
	Txn newrelic.Transaction
}

// Run processes a log event, and publishes an error when level=error
func (h NoticeErrorHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level == zerolog.ErrorLevel {
		h.Txn.NoticeError(errors.New(msg))
	}
}
