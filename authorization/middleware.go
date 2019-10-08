package authorization

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog"

	"github.com/dgrijalva/jwt-go"

	"bitbucket.org/teachingstrategies/go-svc-bootstrap/tokens"

	"github.com/gamegos/jsend"
)

var (
	ErrNoTokenFound = errors.New("jwtauth: no token found")
	ErrTokenInvalid = errors.New("jwtauth: token is not valid")
	ErrSubNotFound  = errors.New("jwtauth: sub is wrong")
	ErrSubInvalid   = errors.New("jwtauth: sub doesn't match requested user")
)

// VisitorRequestContext defines a visitor context to retrieve its content from request context
const VisitorRequestContext = RequestContext("VisitorRequestContextData")

// Visitor defines a context object populated by `MakeAuthValidator`
type Visitor struct {
	UserID int
	SignedIam  string
	Access *Access
}

//  RequestContext used to hold authorization related keys in the request context
type RequestContext string

// ClaimsValidator is a function type to validate clames of parsed jwt token
type ClaimsValidator func(jwt.MapClaims, *http.Request) (*http.Request, error)

// finds IAM and adds it into context
func FindTokenMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// functions which find `iam` token
		findTokenFns := []findTokenFn{iamTokenFromHeader("Authorization"), iamTokenFromQuery,
			iamTokenFromCookie, iamTokenFromHeader("iam")}
		fn := func(w http.ResponseWriter, r *http.Request) {
			var token string
			for _, fn := range findTokenFns {
				token = fn(r)
				if token != "" {
					break
				}
			}
			if token == "" {
				jsend.Wrap(w).Message(ErrNoTokenFound.Error()).Status(http.StatusUnauthorized).Send()
			} else {
				ctx := context.WithValue(r.Context(), RequestContext("token"), token)
				next.ServeHTTP(w, r.WithContext(ctx))
			}
		}
		return http.HandlerFunc(fn)
	}
}

// Verify RSA token middleware
func VerifyTokenMiddleware(keysServerURL string, validator ClaimsValidator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			tokenString, ok := r.Context().Value(RequestContext("token")).(string)
			if !ok {
				jsend.Wrap(w).Message("Internal Server Error: token not found").Status(http.StatusInternalServerError).Send()
				return
			}
			token, err := jwt.Parse(tokenString, makeVerificationRSAKeyFn(keysServerURL))
			if err != nil {
				jsend.Wrap(w).Message(err.Error()).Status(http.StatusUnauthorized).Send()
				return
			}
			var claims jwt.MapClaims

			if claims, ok = token.Claims.(jwt.MapClaims); !ok || !token.Valid {
				jsend.Wrap(w).Message(ErrTokenInvalid.Error()).Status(http.StatusUnauthorized).Send()
				return
			}
			if r, err = validator(claims, r); err != nil {
				jsend.Wrap(w).Message(err.Error()).Status(http.StatusUnauthorized).Send()
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// type findTokenFn declares find token function
type findTokenFn func(r *http.Request) string

func iamTokenFromQuery(r *http.Request) string { return r.URL.Query().Get("iam") }
func iamTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("iam")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func iamTokenFromHeader(header string) func(r *http.Request) string {
	return func(r *http.Request) string {
		bearer := r.Header.Get(header)
		if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
			return bearer[7:]
		}
		return bearer
	}
}

func makeVerificationRSAKeyFn(keysServerURL string) func(tk *jwt.Token) (interface{}, error) {
	return func(tk *jwt.Token) (interface{}, error) {
		kid, err := tokens.RetrieveKID(tk.Header)
		if err != nil {
			return nil, err
		}

		return tokens.RetrievePublicKey(keysServerURL, kid)
	}
}

// Creates auth validator which loads and assigns user's access data into request context
// for further usage with key "access"
// To extract those value from context of request next snipped of code can be used:
// access, ok := r.Context().Value(authorization.RequestContext("access")).(*authorization.Access)
// with nil check as well
func MakeAuthValidator(authorizationServiceURL string, logger *zerolog.Logger) ClaimsValidator {
	return ClaimsValidator(func(claims jwt.MapClaims, r *http.Request) (*http.Request, error) {
		var subscription, token string
		var ok bool
		if sub, ok := claims["sub"]; !ok || sub == nil {
			return r, ErrSubNotFound
		}
		if subscription, ok = claims["sub"].(string); !ok {
			return r, ErrSubNotFound
		}

		t := r.Context().Value(RequestContext("token"))
		if token, ok = t.(string); !ok {
			return r, ErrNoTokenFound
		}
		access, err := LoadAccess(authorizationServiceURL, subscription, token, logger)
		if err != nil {
			return r, err
		}
		var userID int
		if userID, err = strconv.Atoi(subscription); err != nil {
			return r, err
		}
		ctx := context.WithValue(r.Context(), VisitorRequestContext, &Visitor{UserID: userID, SignedIam: token, Access: access})
		return r.WithContext(ctx), nil
	})
}
