package gws

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// This package makes an external call to GWS.  In order to decouple this from GWS,
// these tests	will spin up a local HTTP server that echos back most of what you
// send it. In order to add setup/teardown of the server, these tests are being run
// as part of a testify suite.

type TestSuite struct {
	suite.Suite
	srv  *httptest.Server
	url  *url.URL
	logr *zerolog.Logger
}

// TestRunSuite runs the test suite
func TestRunSuite(t *testing.T) {

	suite.Run(t, new(TestSuite))

}

// SetupSuite runs before this suite to start or create dependencies that are
// difficult to mock
func (suite *TestSuite) SetupSuite() {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path != "/valid" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		qs := r.URL.Query()

		if qs.Get("queryString") == "true" {
			w.WriteHeader(http.StatusTeapot)
			return
		}

		if qs.Get("multi") == "true" && qs.Get("multi2") == "false" {
			w.WriteHeader(http.StatusUnavailableForLegalReasons)
			return
		}

		if qs.Get("error") == "true" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		resp := struct {
			Headers http.Header `json:"headers"`
			Body    string      `json:"res"`
			Cookies []*http.Cookie
		}{
			Headers: r.Header,
			Body:    string(body),
			Cookies: r.Cookies(),
		}

		out, err := json.Marshal(resp)
		if err != nil {
			panic(err)
		}

		_, err = io.WriteString(w, string(out))
		if err != nil {
			panic(err)
		}

	}))

	fmt.Printf("Test Server started on: %+v \n", srv.URL)

	suite.srv = srv
	uri, err := url.Parse(fmt.Sprintf("%s/valid", suite.srv.URL))
	if err != nil {
		panic(err)
	}
	suite.url = uri

	testLogger := zerolog.New(ioutil.Discard)
	suite.logr = &testLogger

}

// TearDownSuite shuts downs and/or deletes dependencies create before suite started
func (suite *TestSuite) TearDownSuite() {

	fmt.Printf("Test Server stopping on: %+v \n", suite.srv.URL)

	suite.srv.Close()

}

func (suite *TestSuite) TestNewProxy() {

	type args struct {
		baseURI      string
		authToken    string
		sharedSecret []byte
		logger       *zerolog.Logger
	}
	tests := []struct {
		name string
		args args
		want interface{}
		err  error
	}{
		{
			name: "Creates Proxy interface",
			args: args{
				baseURI:      "http://unittest.teachingstrategies.com",
				authToken:    "1234",
				sharedSecret: []byte("123"),
				logger:       suite.logr,
			},
			want: &proxy{
				config: gwsConfig{
					baseURI:      makeTestURL("http://unittest.teachingstrategies.com"),
					authToken:    "1234",
					sharedSecret: []byte("123"),
				},
				logger: suite.logr,
			},
		},
		{
			name: "Fails to create proxy when baseURI is empty string",
			args: args{
				baseURI:      "",
				authToken:    "1234",
				sharedSecret: []byte("123"),
				logger:       suite.logr,
			},
			want: nil,
			err:  ErrProxyMisconfigured,
		},
		{
			name: "Fails to create proxy when authToken is empty string",
			args: args{
				baseURI:      "1234",
				authToken:    "",
				sharedSecret: []byte("123"),
				logger:       suite.logr,
			},
			want: nil,
			err:  ErrProxyMisconfigured,
		},
		{
			name: "Fails to create proxy when sharedSecret is empty",
			args: args{
				baseURI:      "1234",
				authToken:    "1234",
				sharedSecret: []byte{},
				logger:       suite.logr,
			},
			want: nil,
			err:  ErrProxyMisconfigured,
		},
		{
			name: "Fails to create proxy when logger is nil",
			args: args{
				baseURI:      "1234",
				authToken:    "1234",
				sharedSecret: []byte("1234"),
				logger:       nil,
			},
			want: nil,
			err:  ErrProxyMisconfigured,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			p, err := NewProxy(tt.args.baseURI, tt.args.authToken, tt.args.sharedSecret, tt.args.logger)
			assert.Equal(suite.T(), tt.want, p)

			checkForError(suite.T(), tt.err, err)
		})
	}

}

func (suite *TestSuite) TestProxyPost() {

	iam := "123"

	type args struct {
		endpoint string
		payload  []byte
	}
	tests := []struct {
		name   string
		gws    *proxy
		args   args
		res    []byte
		status int
		err    error
	}{
		{
			name: "Sends request to local proxy with expected headers",
			gws: &proxy{
				config: gwsConfig{
					baseURI:      suite.url,
					authToken:    "1234",
					sharedSecret: []byte("1234"),
				},
				logger: suite.logr,
			},
			args:   args{},
			res:    []byte(`{"headers":{"Accept-Encoding":["gzip"],"Authorization":["1234"],"Content-Length":["0"],"Content-Type":["application/json"],"Iam":["123"],"User-Agent":["Go-http-client/1.1"]},"res":"","Cookies":[]}`),
			status: http.StatusOK,
		},
		{
			name:   "Returns error when request cannot by made due to misconfiguration",
			gws:    &proxy{config: gwsConfig{}, logger: suite.logr},
			args:   args{},
			err:    ErrProxyMisconfigured,
			status: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			gotStatus, gotJsn, err := tt.gws.Post(&iam, tt.args.endpoint, tt.args.payload)
			assert.Equal(suite.T(), tt.status, gotStatus)
			assert.Equal(suite.T(), string(tt.res), string(gotJsn))

			checkForError(suite.T(), tt.err, err)
		})
	}

}

func (suite *TestSuite) TestProxyGet() {

	iam := "123"

	type args struct {
		endpoint    string
		queryString *url.Values
	}
	tests := []struct {
		name   string
		gws    *proxy
		args   args
		res    []byte
		status int
		err    error
	}{
		{
			name: "Sends request to local proxy with expected headers",
			gws: &proxy{
				config: gwsConfig{
					baseURI:      suite.url,
					authToken:    "1234",
					sharedSecret: []byte("1234"),
				},
				logger: suite.logr,
			},
			args:   args{},
			res:    []byte(`{"headers":{"Accept-Encoding":["gzip"],"Authorization":["1234"],"Content-Type":["application/json"],"Iam":["123"],"User-Agent":["Go-http-client/1.1"]},"res":"","Cookies":[]}`),
			status: http.StatusOK,
		},
		{
			name: "Converts queryString into query string",
			gws: &proxy{
				config: gwsConfig{
					baseURI:      suite.url,
					authToken:    "1234",
					sharedSecret: []byte("1234"),
				},
				logger: suite.logr,
			},
			args:   args{queryString: &url.Values{"queryString": []string{"true"}}},
			status: http.StatusTeapot,
		},
		{
			name: "Converts multiple values in queryString into query string",
			gws: &proxy{
				config: gwsConfig{
					baseURI:      suite.url,
					authToken:    "1234",
					sharedSecret: []byte("1234"),
				},
				logger: suite.logr,
			},
			args:   args{queryString: &url.Values{"multi": []string{"true"}, "multi2": []string{"false"}}},
			status: http.StatusUnavailableForLegalReasons,
		},
		{
			name:   "Returns error when request cannot by made due to misconfiguration",
			gws:    &proxy{config: gwsConfig{}, logger: suite.logr},
			args:   args{},
			err:    ErrProxyMisconfigured,
			status: http.StatusInternalServerError,
		},
		{
			name: "Returns error when destination returns an error",
			gws: &proxy{
				config: gwsConfig{
					baseURI:      suite.url,
					authToken:    "1234",
					sharedSecret: []byte("1234"),
				},
				logger: suite.logr,
			},
			args:   args{queryString: &url.Values{"error": []string{"true"}}},
			err:    ErrProxyRequestFailed,
			status: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			gotStatus, gotJsn, err := tt.gws.Get(&iam, tt.args.endpoint, tt.args.queryString)

			assert.Equal(suite.T(), tt.status, gotStatus)
			assert.Equal(suite.T(), string(tt.res), string(gotJsn))

			checkForError(suite.T(), tt.err, err)
		})
	}

}

func (suite *TestSuite) TestSend() {

	tests := []struct {
		name       string
		request    *http.Request
		wantStatus int
		wantBody   []byte
		err        error
	}{
		{
			name: "Reads and returns res on success",
			request: func() *http.Request {
				r, _ := http.NewRequest("GET", suite.url.String(), strings.NewReader(`{"hello":"world"}`))
				return r
			}(),
			wantStatus: http.StatusOK,
			wantBody:   []byte(`{"headers":{"Accept-Encoding":["gzip"],"Content-Length":["17"],"User-Agent":["Go-http-client/1.1"]},"res":"{\"hello\":\"world\"}","Cookies":[]}`),
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			gotStatus, gotBody, err := send(tt.request, suite.logr)
			assert.Equal(suite.T(), tt.wantStatus, gotStatus)
			assert.Equal(suite.T(), string(tt.wantBody), string(gotBody))

			checkForError(suite.T(), tt.err, err)
		})
	}

}

func (suite *TestSuite) TestAddHeaders() {

	iam := "123"

	type args struct {
		iam       *string
		request   *http.Request
		authToken string
	}
	tests := []struct {
		name    string
		args    args
		headers []string
	}{
		{
			name: "Adds required headers for GWS with iam",
			args: args{
				iam:       &iam,
				request:   httptest.NewRequest("GET", suite.url.String(), strings.NewReader("")),
				authToken: "unittest",
			},
			headers: []string{"Authorization", "Content-Type", "Iam"},
		},
		{
			name: "Adds required headers for GWS without iam",
			args: args{
				iam:       nil,
				request:   httptest.NewRequest("GET", suite.url.String(), strings.NewReader("")),
				authToken: "unittest",
			},
			headers: []string{"Authorization", "Content-Type"},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			for _, h := range tt.headers {
				assert.Nil(suite.T(), tt.args.request.Header[h])
			}

			addHeaders(tt.args.request, tt.args.iam, tt.args.authToken)

			assert.Equal(suite.T(), len(tt.args.request.Header), len(tt.headers))
			assert.Equal(suite.T(), []string{tt.args.authToken}, tt.args.request.Header["Authorization"])
			assert.Equal(suite.T(), []string{"application/json"}, tt.args.request.Header["Content-Type"])
		})
	}

}

func makeTestURL(baseURI string) *url.URL {

	uri, err := url.Parse(baseURI)
	if err != nil {
		panic(err)
	}
	return uri

}

func checkForError(t *testing.T, want error, got error) {

	if want == nil {
		assert.Nil(t, got)
		return
	}

	if got == nil {
		assert.Failf(t, "expected error", `expected "%s"`, want.Error())
		return
	}

	assert.Equal(t, want.Error(), got.Error())

}
