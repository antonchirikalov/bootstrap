package middleware

import (
	"context"
	"net/http"

	"bitbucket.org/teachingstrategies/go-svc-bootstrap/response"
	"github.com/rs/zerolog"
)

// originTagKey is the context key used to retrieve the request's origin
const originTagKey = bootstrapContextKey("origin")

// OriginFromContext returns the origin value from the given context
func OriginFromContext(ctx context.Context) string {
	return ctx.Value(originTagKey).(string)
}

// ValidateOriginMiddleware checks if the incoming request is from an allowed Origin.
// If the origin is allowed, it will be added to the request context as "origin"
func ValidateOriginMiddleware(allowedOrigins []string, logger *zerolog.Logger) func(next http.Handler) http.Handler {
	originsMap := make(map[string]struct{}, len(allowedOrigins))

	// convert array to map to speed up lookups
	for _, o := range allowedOrigins {
		originsMap[o] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			originIn := r.Header.Get("Origin")

			denyResp := response.New()
			denyResp.StatusCode = http.StatusBadRequest
			denyResp.Message = "Origin is not permitted"

			if originIn == "" {
				logger.Warn().Msg("Request has no Origin header")

				denyResp.Message = "Origin is required"

				response.Send(w, denyResp)

				return
			}

			if _, ok := originsMap[originIn]; ok {
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), originTagKey, originIn)))
			} else {
				logger.Warn().Str("originIn", originIn).Msg("Origin is not permitted")

				response.Send(w, denyResp)
			}
		})
	}
}
