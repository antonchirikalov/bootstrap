package middleware

import (
	"context"
	"net/http"
	"strings"

	"golang.org/x/text/language"

	"github.com/go-chi/chi"
)

const (
	// localeTagKey key
	localeTagKey = bootstrapContextKey("LocaleTag")
)

// LocaleTag stores language, region, variant
type LocaleTag struct {
	Language *language.Base
	Region   *language.Region
	Variant  string
}

// Original formats the value of locale tag in the way as it was consumed
func (lt *LocaleTag) Original() string {
	locale := lt.Language.String() + "-" + strings.ToLower(lt.Region.String())
	if lt.Variant == "" {
		return locale
	}
	return locale + "-" + lt.Variant
}

// BundleName formats the value of locale tag in the way of bundle name
func (lt *LocaleTag) BundleName() string {
	locale := lt.Language.String() + "_" + strings.ToUpper(lt.Region.String())
	if lt.Variant == "" {
		return locale
	}
	return locale + "_" + strings.ToUpper(lt.Variant)
}

// AddLocale method wraps and adds to the root path handling of
// locales in format `lang-region`, `lang-region-variant` or `` empty.
// After parsing these values are populated into request context
// with `LocaleContext` as key type LocaleTag as value.
// Absence of locale (empty) will use default `en-us`.
func AddLocale(routes func(r chi.Router)) func(r chi.Router) {
	return func(r chi.Router) {
		r = r.With(localeMiddleware)
		r.Route("/{locale:^[a-z]{2,3}\\-[a-z]{2,3}(\\-[a-z]{2,3})?$|^$}", routes)
		r.Route("/", routes)
	}
}

// LocaleTagFromContext returns the locale tag from the given context.
func LocaleTagFromContext(ctx context.Context) *LocaleTag {
	tag := ctx.Value(localeTagKey).(LocaleTag)

	return &tag
}

func localeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		localeStr := chi.URLParam(r, "locale")
		if localeStr == "" {
			localeStr = "en-us"
		}
		locale := strings.Split(localeStr, "-")

		base, baseErr := language.ParseBase(locale[0])
		region, regionErr := language.ParseRegion(locale[1])

		if baseErr != nil || regionErr != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		// populate context with language, region and variant
		localeTag := LocaleTag{Language: &base, Region: &region}
		if len(locale) > 2 {
			localeTag.Variant = locale[2]
		}

		ctx := context.WithValue(r.Context(), localeTagKey, localeTag)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
