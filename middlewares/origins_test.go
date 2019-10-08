package middleware

import (
	"bufio"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	"github.com/go-chi/chi"
)

func TestOriginMiddleware(t *testing.T) {
	tests := []struct {
		name             string
		allowedOrigins   []string
		requestingOrigin string
		wantStatus       int
	}{
		{
			name:             "valid origin",
			allowedOrigins:   []string{"http://abc.com", "http://123.com"},
			requestingOrigin: "http://abc.com",
			wantStatus:       http.StatusOK,
		},
		{
			name:             "disallowed origin",
			allowedOrigins:   []string{"http://abc.com", "http://123.com"},
			requestingOrigin: "http://notallowed.com",
			wantStatus:       http.StatusBadRequest,
		},
		{
			name:             "origin header missing",
			allowedOrigins:   []string{"http://abc.com", "http://123.com"},
			requestingOrigin: "",
			wantStatus:       http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		if tt.requestingOrigin != "" {
			r.Header.Set("Origin", tt.requestingOrigin)
		}

		rctx := chi.NewRouteContext()
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

		var b bytes.Buffer
		logger := zerolog.New(bufio.NewWriter(&b))

		router := chi.NewRouter()
		router.Use(ValidateOriginMiddleware(tt.allowedOrigins, &logger))
		router.Route("/test", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)

				origin := OriginFromContext(r.Context())

				if origin != tt.requestingOrigin {
					t.Fatal("Origin context does not match provided: want", tt.requestingOrigin, "got", origin)
				}
			})
		})
		router.ServeHTTP(w, r)

		if w.Code != tt.wantStatus {
			t.Fatal(tt.name, ": Wrong result", w.Code, "want", tt.wantStatus)
		}
	}
}
