package middleware

import (
	"context"
	"net/http"

	"bitbucket.org/teachingstrategies/go-svc-bootstrap/utils"
)

type bootstrapContextKey string

func (b bootstrapContextKey) String() string {
	return "middleware context key: " + string(b)
}

func AddContext(fields map[string]interface{}) func(next http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {

			augmentedContext := r.Context()

			for k, v := range fields {
				augmentedContext = context.WithValue(augmentedContext, utils.ContextKey(k), v)
			}

			next.ServeHTTP(w, r.WithContext(augmentedContext))
		}
		return http.HandlerFunc(fn)
	}
}
