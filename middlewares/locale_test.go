package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
)

func TestLocaleMiddleware(t *testing.T) {
	valid := []string{"en-us", "es-us", "en-us-nc", "en-us-fac", ""}
	invalid := []string{"en_us", "en-US", "EN-us", "eN-us", "en", "EN-US", "xx-xx"}

	t.Log("Running valid paths...")
	for _, pattern := range valid {
		testMiddleware(pattern, true, t)
	}

	t.Log("Running invalid paths...")
	for _, pattern := range invalid {
		testMiddleware(pattern, false, t)
	}
}

func testMiddleware(pattern string, valid bool, t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/planning/"+pattern+"/test", nil)
	rctx := chi.NewRouteContext()
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	routes := func(r chi.Router) {
		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			localeTag := LocaleTagFromContext(r.Context())
			ptn := pattern
			if ptn == "" {
				ptn = "en-us"
			}
			if localeTag.Original() != ptn {
				t.Fatal("Wrong original for", ptn, "vs", localeTag.Original())
			}
			if strings.ToLower(localeTag.BundleName()) != strings.ReplaceAll(ptn, "-", "_") {
				t.Fatal("Wrong bundle name for", ptn)
			}
		})
	}

	router := chi.NewRouter()
	router.Route("/planning", AddLocale(routes))
	router.ServeHTTP(w, r)

	if (valid && w.Code != 200) || (!valid && w.Code != 404) {
		t.Fatal("Wrong result", w.Code, "for", pattern, ";valid=", valid)
	}
}
