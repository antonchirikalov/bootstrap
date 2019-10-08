package authorization

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"github.com/stretchr/testify/mock"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

type mockHandler struct{ mock.Mock }

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) { h.Called(w, r) }

const notRealURI = "http://notreal"

func TestFindTokenMiddleware(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		token  string
	}{
		{
			name:   "token not found",
			status: 401,
			body:   `{"data":null,"message":"jwtauth: no token found","status":"fail"}`,
			token:  "",
		},
		{
			name:   "token not found",
			status: 200,
			body:   "",
			token:  "1008",
		},
	}
	mockHandler := &mockHandler{}
	mockHandler.On("ServeHTTP", mock.Anything, mock.Anything).Return().Once()
	handler := FindTokenMiddleware()(mockHandler)
	for _, tt := range tests {
		req := httptest.NewRequest("GET", "/test", strings.NewReader(""))
		req.Header.Add("iam", tt.token)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, tt.status, w.Code, tt.name)
		assert.Equal(t, tt.body, w.Body.String(), tt.name)
	}
	mockHandler.AssertExpectations(t)
}

func TestVerifyTokenMiddleware(t *testing.T) {
	wrongToken := "eyJhbGciOiJSUzI1NiIsImtpZCI6ImMyYzUyMjk5LTc4NzMtNDViOS05NzYyLTUwMGM5OTRjOTNhMSIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiTXlUU09yZ0FkbWluMSBNeVRTT3JnQWRtaW4xIiwiZXhwIjoxNTU4MTEyMTEwLCJqdGkiOiIyNjQzYjFiZi01MTZiLTRjMmYtOGJlMy1mNjk4YjM2NTc0YjgiLCJpYXQiOjE1NTgwOTc3MTAsImlzcyI6ImFwaS50ZWFjaGluZ3N0cmF0ZWdpZXMuY29tIiwibmJmIjoxNTU4MDk3NzEwLCJzdWIiOiIxMTY0Njc4In0.ZHRMp_7yzdPGh8qTpnCWGRDpzulrOVum-57bh4lbrQLyMn7Yf-GZO5EGwkimLEBrPnqY1YJM8tP9eqIn1auLjMghguSrqLwrGpKlqE8Oo4hNl6XzyIbKs1PKRWEVWbSnMkv0on8yiC5n1To5UOGs1hYENGKCPE_YbBd1wStvwjyip08P9To9-_rvq4w7nl_lrSedplapCJ5cZJrwzm4FuE9hiMNQ65gZCptI9AM6vRSo-x7wN15yI9vCfio7mdqRYczBjMS_UpqZj5n0zGpFAhAkg_H0iJUkxMS6FJsY2mwGfh64NQGPvA_dTMMLCytFy7nUSNqJqp-j0Vr2S0h3HOChiOMQg2l0QOxArfA7XInAVj5ahvUfyE4tOoB1FLAjAv1g-PY_K0SG0qq-ekST49a_ip0Tg9nQn2yreOd9Zv8azUoHoWMd7A7IupBZt3uFR0cjdOcdVuoE_8K4nfDQEf4f5Op1HuKm41q1qkhPIFaReGpysW4OBDqjZSlBiMMB9l7Mlvy6Qd7c7ggpDR0o-NzGCsvsgRQcEr3h-E__32gEL8JEjkxZZrPZyO6y0R7Z-hH94Id6DYmp4mFDW8gGn2ds1rmxBz8Lv0sMzkf5qYaNWdVs2gqUTJ3lEi1pA-XzWEfxZ5ekCBuksFR2vDWIeZOCIBCkgLF9jYFQSCJutpQ"
	correctToken := "eyJhbGciOiJSUzI1NiIsImtpZCI6ImMyYzUyMjk5LTc4NzMtNDViOS05NzYyLTUwMGM5OTRjOTNhMSIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiTXlUU09yZ0FkbWluMSBNeVRTT3JnQWRtaW4xIiwiZXhwIjoxMDU1OTU2MTI1OCwianRpIjoiOTRlOGUzM2QtZmQzZC00ZDhkLWI1NmItNjc4YWM1MDQ0MzQ2IiwiaWF0IjoxNTU5NTYxMjU4LCJpc3MiOiJhcGkudGVhY2hpbmdzdHJhdGVnaWVzLmNvbSIsIm5iZiI6MTU1OTU2MTI1OCwic3ViIjoiMTE2NDY3OCJ9.VsLUsDhIZyc6Ryt5X84QmxdIP-wo7TH5dfhDTxiNv21v-YPIodwmnReatebfpvrhDxwoG_XQ3f_32HTMlqVaUAvi4mIUFFDaemYsWhUXPwmDBe3-WeI1uIuruD2xrEJ52rrHgUPuEfAktpccxCmxToqrMma7Pp-d_SVrKIPNq4paXwFn90ZyIpLTZGtpfD2zAkcndJV022_KX2BdpwmIMXDMiiRgSLk7NNrqoOhHkGJfxVVDjr0LT70HFcfOCqVSZUvNBpyURAN7jRCG8dsoQo-FPKjoV2nXm1jXD4TySEy2kRUj7QziMWSFl0FYM6i3NLu6E1E3wGjObEb0n5TtDLUFswJ4mTuPJAuNGeYi7abw7H369KDYmpO9k7Dz2P2NC9yTOWFiG5lc4CvSP2Z2iqTs7DJ8940mFS7fRtLne1ieQi0qX-3lH9GFxrEVnXKW9leYJ4O3JDY7c6F-oHLBZVIXgK-yQUA0QJJNHQA4swlcrDnN9NpdPveq0op5EuNMOEHXDJF2FJH7o0ogt4lYDhnLmT7cFYBk6Wj8Iik5dUvcL7xgPk-43cjnpYLVyaDuXmRZkQwGv-t3pB93zIjw5ZMZw4zL4VZ9dckiTY_ptoPYrxkJhJGprP_UanacVrfCiJFCd5DD3Un8O0iv0bEi76kwg-Bp14IZqJGMgHITnfk"
	publicKeyJSON := `{
		"data": {
			"keyID":"c2c52299-7873-45b9-9762-500c994c93a1",
			"createdOn":"2019-03-25T15:03:06Z",
			"sourceID":"31c2b2e3-7491-4e31-a9a3-32ddb85e737d",
			"key":"-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAt4fIro6MnV0A///8UyJV\nobtKtBBFiY+9JU6egxnwGoCzlo5HzAnmGoRapvXVRlBFFTD6QvcWCPmofWTetDR/\nj3+rnDy7QpiaIsPBV9aQ4RLPmUG1jxmps/LF0tbC7HjnQGWb97456KNCgGtbbKPX\nVgB+ETe71/dPtVicRpTBM1PHEvE4LS84/q2czBwcerXFADiYPYih52Q0qQfEToju\n0cz5wObzFLWi3c4oABoSOoFYA+0VI3Fypa/dgKYlyMBOaVYXLIH5MMy77olao7VW\nHO1LEotnTIfSZY7z6pk3jIKpPvhRaY3TxF72NQBy8JvJRK/ngmKwsqCr2h4LLvZ+\nArufQpfWFK/bpR+T7gFbaddvGonHaA8/wRlc5dbp26R1DumKj7JZCHDzhv+TmGod\n/rjbvegzGaAvdCL2/bdCd7TtpcOmVGu2WENAT2r3p9GQXEdpxPVNRJoaEvvtIGU1\nBeC3Gf0+/PVxzqzapYfrGjb1dcpw6IDPm6/SXT4+sEo9p2RGB++GzJPXjN4YDAs5\njHA9gKXhykbrk9sTEA5HseZxWdLr0iYB+eTRQOcgho/1f1Gq8iP3mJRycYr+84nm\nzuI/CRBngMRDBESg52nr71bONXCDb9Hz01HGlAGg0xJzovK+5zNlucMT5ps8t3ij\n4848FGewjc3pG6kxW9SQWL8CAwEAAQ==\n-----END PUBLIC KEY-----\n"
		}
	}`
	tests := []struct {
		name              string
		keysServerHandler func(http.ResponseWriter, *http.Request)
		keysServerURL     func(*httptest.Server) string
		claimsValidator   ClaimsValidator
		token             string
		status            int
		body              string
	}{
		{
			name:              "token not found",
			status:            500,
			body:              `{"message":"Internal Server Error: token not found","status":"error"}`,
			keysServerHandler: func(http.ResponseWriter, *http.Request) {},
			keysServerURL:     func(*httptest.Server) string { return notRealURI },
		},
		{
			name:              "jwt parse error",
			status:            401,
			token:             wrongToken,
			body:              `{"data":null,"message":"unable to retrieve license '404 Not Found', status 404","status":"fail"}`,
			keysServerHandler: func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(404) },
			keysServerURL:     func(s *httptest.Server) string { return s.URL },
		},
		{
			name:   "verify token failed",
			status: 401,
			token:  correctToken,
			body:   `{"data":null,"message":"unable to validate (fake err)","status":"fail"}`,
			keysServerHandler: func(w http.ResponseWriter, r *http.Request) {
				if _, err := w.Write([]byte(publicKeyJSON)); err != nil {
					assert.NoError(t, err)
				}
			},
			claimsValidator: func(claims jwt.MapClaims, r *http.Request) (*http.Request, error) {
				return r, fmt.Errorf("unable to validate (fake err)")
			},
			keysServerURL: func(s *httptest.Server) string { return s.URL },
		},
		{
			name:   "success case",
			status: 200,
			token:  correctToken,
			body:   ``,
			keysServerHandler: func(w http.ResponseWriter, r *http.Request) {
				if _, err := w.Write([]byte(publicKeyJSON)); err != nil {
					assert.NoError(t, err)
				}
			},
			claimsValidator: func(claims jwt.MapClaims, r *http.Request) (*http.Request, error) {
				return r, nil
			},
			keysServerURL: func(s *httptest.Server) string { return s.URL },
		},
	}
	mockHandler := &mockHandler{}
	mockHandler.On("ServeHTTP", mock.Anything, mock.Anything).Return().Once()

	for _, tt := range tests {
		keysServer := httptest.NewServer(http.HandlerFunc(tt.keysServerHandler))
		defer keysServer.Close()
		handler := VerifyTokenMiddleware(tt.keysServerURL(keysServer), tt.claimsValidator)(mockHandler)
		req := httptest.NewRequest("GET", "/test", strings.NewReader(""))
		if tt.token != "" {
			req = req.WithContext(context.WithValue(req.Context(), RequestContext("token"), tt.token))
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, tt.status, w.Code, tt.name)
		assert.Equal(t, tt.body, w.Body.String(), tt.name)
	}
}

func TestIamTokenFromQuery(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "empty", want: ""},
		{name: "non empty", want: "myiam"},
	}
	for _, tt := range tests {
		req := httptest.NewRequest("GET", "/test?iam="+tt.want, nil)
		assert.Equal(t, tt.want, iamTokenFromQuery(req), tt.name)
	}
}

func TestIamTokenFromCookie(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "empty", want: ""},
		{name: "non empty", want: "myiam"},
	}
	for _, tt := range tests {
		req := httptest.NewRequest("GET", "/test", nil)
		// req.Header.Set("Cookie", fmt.Sprintf("iam=%s", tt.want))
		// t.Log(req.Header)
		req.AddCookie(&http.Cookie{Name: "iam", Value: tt.want})
		assert.Equal(t, tt.want, iamTokenFromCookie(req), tt.name)
	}
}

func TestIamTokenFromHeader(t *testing.T) {
	fn := iamTokenFromHeader("kid")
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{
			name:   "no header",
			header: "",
			want:   "",
		},
		{
			name:   "without bearer",
			header: "108",
			want:   "108",
		},
		{
			name:   "with bearer",
			header: "BEARER 108",
			want:   "108",
		},
	}
	for _, tt := range tests {
		req := &http.Request{Header: http.Header{}}
		if tt.header != "" {
			req.Header.Add("kid", tt.header)
		}
		assert.Equal(t, tt.want, fn(req), tt.name)
	}
}

func TestMakeVerificationRSAKeyFn(t *testing.T) {
	tests := []struct {
		name              string
		keysServerHandler func(http.ResponseWriter, *http.Request)
		keysServerURL     func(*httptest.Server) string
		token             *jwt.Token
		err               string
	}{
		{
			name:              "no `kid` found",
			keysServerHandler: func(http.ResponseWriter, *http.Request) {},
			keysServerURL:     func(*httptest.Server) string { return notRealURI },
			err:               "key is invalid",
			token:             &jwt.Token{},
		},
		{
			name:              "unable to load public key",
			keysServerHandler: func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(404) },
			keysServerURL:     func(s *httptest.Server) string { return s.URL },
			err:               "unable to retrieve license '404 Not Found', status 404",
			token:             &jwt.Token{Header: map[string]interface{}{"kid": "42"}},
		},
	}
	for _, tt := range tests {
		keysServer := httptest.NewServer(http.HandlerFunc(tt.keysServerHandler))
		defer keysServer.Close()
		rsaFn := makeVerificationRSAKeyFn(tt.keysServerURL(keysServer))
		i, err := rsaFn(tt.token)
		assert.EqualError(t, err, tt.err, tt.name)
		_ = i
	}
}

func TestMakeAuthValidator(t *testing.T) {
	token := "eyJhbGciOiJSUzI1NiIsImtpZCI6ImMyYzUyMjk5LTc4NzMtNDViOS05NzYyLTUwMGM5OTRjOTNhMSIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiTXlUU09yZ0FkbWluMSBNeVRTT3JnQWRtaW4xIiwiZXhwIjoxNTU4MTEyMTEwLCJqdGkiOiIyNjQzYjFiZi01MTZiLTRjMmYtOGJlMy1mNjk4YjM2NTc0YjgiLCJpYXQiOjE1NTgwOTc3MTAsImlzcyI6ImFwaS50ZWFjaGluZ3N0cmF0ZWdpZXMuY29tIiwibmJmIjoxNTU4MDk3NzEwLCJzdWIiOiIxMTY0Njc4In0.ZHRMp_7yzdPGh8qTpnCWGRDpzulrOVum-57bh4lbrQLyMn7Yf-GZO5EGwkimLEBrPnqY1YJM8tP9eqIn1auLjMghguSrqLwrGpKlqE8Oo4hNl6XzyIbKs1PKRWEVWbSnMkv0on8yiC5n1To5UOGs1hYENGKCPE_YbBd1wStvwjyip08P9To9-_rvq4w7nl_lrSedplapCJ5cZJrwzm4FuE9hiMNQ65gZCptI9AM6vRSo-x7wN15yI9vCfio7mdqRYczBjMS_UpqZj5n0zGpFAhAkg_H0iJUkxMS6FJsY2mwGfh64NQGPvA_dTMMLCytFy7nUSNqJqp-j0Vr2S0h3HOChiOMQg2l0QOxArfA7XInAVj5ahvUfyE4tOoB1FLAjAv1g-PY_K0SG0qq-ekST49a_ip0Tg9nQn2yreOd9Zv8azUoHoWMd7A7IupBZt3uFR0cjdOcdVuoE_8K4nfDQEf4f5Op1HuKm41q1qkhPIFaReGpysW4OBDqjZSlBiMMB9l7Mlvy6Qd7c7ggpDR0o-NzGCsvsgRQcEr3h-E__32gEL8JEjkxZZrPZyO6y0R7Z-hH94Id6DYmp4mFDW8gGn2ds1rmxBz8Lv0sMzkf5qYaNWdVs2gqUTJ3lEi1pA-XzWEfxZ5ekCBuksFR2vDWIeZOCIBCkgLF9jYFQSCJutpQ"

	tests := []struct {
		name               string
		authServiceHandler func(http.ResponseWriter, *http.Request)
		authServiceURL     func(*httptest.Server) string
		err                string
		token              string
		body               string
		claims             jwt.MapClaims
	}{
		{
			name:               "sub is wrong (absent)",
			err:                "jwtauth: sub is wrong",
			claims:             jwt.MapClaims(map[string]interface{}{}),
			authServiceHandler: func(http.ResponseWriter, *http.Request) {},
			authServiceURL:     func(*httptest.Server) string { return notRealURI },
		},
		{
			name:               "sub is wrong (wrong type)",
			err:                "jwtauth: sub is wrong",
			claims:             jwt.MapClaims(map[string]interface{}{"sub": 42}),
			authServiceHandler: func(http.ResponseWriter, *http.Request) {},
			authServiceURL:     func(*httptest.Server) string { return notRealURI },
		},
		{
			name:               "token not found",
			err:                "jwtauth: no token found",
			claims:             jwt.MapClaims(map[string]interface{}{"sub": "42"}),
			authServiceHandler: func(http.ResponseWriter, *http.Request) {},
			authServiceURL:     func(*httptest.Server) string { return notRealURI },
		},
		{
			name:               "load access error",
			err:                "unable to retrieve access data '404 Not Found', status 404",
			token:              token,
			claims:             jwt.MapClaims(map[string]interface{}{"sub": "42"}),
			authServiceHandler: func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(404) },
			authServiceURL:     func(s *httptest.Server) string { return s.URL },
		},
		{
			name:   "user id is not a number",
			err:    "strconv.Atoi: parsing \"42nan\": invalid syntax",
			token:  token,
			claims: jwt.MapClaims(map[string]interface{}{"sub": "42nan"}),
			authServiceHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				fmt.Fprint(w, `{"Status":"success", "Data": {"SuperUser":true}, "Message": "OK"}`)
			},
			authServiceURL: func(s *httptest.Server) string { return s.URL },
		},
		{
			name:   "success",
			err:    "",
			token:  token,
			claims: jwt.MapClaims(map[string]interface{}{"sub": "42"}),
			authServiceHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				fmt.Fprint(w, `{"Status":"success", "Data": {"SuperUser":true}, "Message": "OK"}`)
			},
			authServiceURL: func(s *httptest.Server) string { return s.URL },
			body:           `{"Access":{"SuperUser":true}, "SignedIam":"` + token + `", "UserID": 42}`,
		},
	}
	mockHandler := &mockHandler{}
	mockHandler.On("ServeHTTP", mock.Anything, mock.Anything).Return().Once()

	for _, tt := range tests {
		authService := httptest.NewServer(http.HandlerFunc(tt.authServiceHandler))
		defer authService.Close()
		logger := zerolog.Nop()
		validator := MakeAuthValidator(tt.authServiceURL(authService), &logger)
		req := httptest.NewRequest("GET", "/test", strings.NewReader(""))
		if tt.token != "" {
			req = req.WithContext(context.WithValue(req.Context(), RequestContext("token"), token))
		}
		request, err := validator(tt.claims, req)
		if err != nil {
			assert.EqualError(t, err, tt.err, tt.name)
		}
		if tt.body != "" {
			assert.NotNil(t, request, tt.name)
			visitor := request.Context().Value(VisitorRequestContext)
			assert.NotNil(t, visitor, tt.name)
			visitorJSON, err := json.Marshal(visitor)
			assert.NoError(t, err, tt.name)
			assert.JSONEq(t, tt.body, string(visitorJSON), tt.name)
		}
	}
}
