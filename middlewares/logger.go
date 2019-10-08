package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	tsjwt "bitbucket.org/teachingstrategies/go-svc-bootstrap/tsjwt"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
)

// AddLogging provides access logging for all requests
// It will attempt to add TeachingStrategies specific fields where possible
//    uid, vid
// It implements the responsewriter interface to provide status logging as well
// if panicStdout is true, the panic stacktrace will print to stdout and will not be logged, useful for local development
func AddLogging(logger *zerolog.Logger, panicStdout bool) func(next http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {

			rid := middleware.GetReqID(r.Context())

			// All log entries will contain a request ID
			ctxLogger := logger.With().Str("rid", rid).Logger()

			rec := statusRecorder{w, 200}

			reqStartTime := time.Now()

			defer func() {

				reqEndTime := time.Now()

				accessLog := map[string]interface{}{
					"remote_ip":  r.RemoteAddr,
					"host":       r.Host,
					"url":        r.URL.Path,
					"proto":      r.Proto,
					"method":     r.Method,
					"user_agent": r.Header.Get("User-Agent"),
					"dur_ms":     reqEndTime.Sub(reqStartTime).Nanoseconds() / 1000000,
					"status":     rec.status,
				}

				// If there are contextual data points present about the user
				// that we would like to include, like the user id, and visitor id,
				// we'll include them in the access log
				_, claims, _ := tsjwt.FromContext(r.Context())

				accessLog["uid"] = claims.UserID
				accessLog["vid"] = claims.VisitorID

				if e := recover(); e != nil {
					accessLog["status"] = 500
					ctxLogger.Info().Timestamp().Fields(accessLog).Msg("")
					stackBytes := debug.Stack()

					baseLog := ctxLogger.Error().Timestamp().Interface("recover_info", e)
					if panicStdout {
						fmt.Println(string(stackBytes))
						baseLog.Msg("panic_on_request")
					} else {
						baseLog.Bytes("debug_stack", stackBytes).Msg("panic_on_request")
					}

					// propagate the panic
					panic(e)
				}

				ctxLogger.Info().Timestamp().Fields(accessLog).Msg("")

			}()

			next.ServeHTTP(&rec, r)
		}
		return http.HandlerFunc(fn)
	}
}

// statusRecorder implements the response writer interface
// it tracks the status code for logging later
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}
