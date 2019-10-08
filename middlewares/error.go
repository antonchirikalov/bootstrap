package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/middleware"
	newrelic "github.com/newrelic/go-agent"
	"github.com/rs/zerolog"
)

// ObservedHandler is an http handler that supports an error return
// for instrumentation
type ObservedHandler func(w http.ResponseWriter, r *http.Request) error

// ErrorObserver leverages the error response of an observed handler to
// provide visibility
func ErrorObserver(h ObservedHandler, logger *zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := h(w, r)

		if err != nil {

			http.Error(w, err.Error(), http.StatusInternalServerError)

			logger.Error().Timestamp().Bytes("debug_stack", debug.Stack()).Str("rid", middleware.GetReqID(r.Context())).Msg(err.Error())

			if txn, ok := w.(newrelic.Transaction); ok {
				txn.NoticeError(err)
			}

		}
	}
}
