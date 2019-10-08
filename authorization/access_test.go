package authorization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rs/zerolog"
)

func TestLoadAccess(t *testing.T) {
	testCases := []struct {
		name      string
		serverURI func(*httptest.Server) string
		handler   func(rw http.ResponseWriter, req *http.Request)
		result    *Access
		resultErr string
		userID    string
	}{
		{
			name:      "unable to parse url",
			serverURI: func(*httptest.Server) string { return "_http://failme.co" },
			handler:   func(w http.ResponseWriter, r *http.Request) {},
			resultErr: "unable to parse authorization service url: parse _http://failme.co: first path segment in URL cannot contain colon",
			userID:    "42",
		},
		{
			name:      "request failed",
			serverURI: func(*httptest.Server) string { return "http://fakeuri" },
			handler:   func(w http.ResponseWriter, r *http.Request) {},
			resultErr: "Get http://fakeuri/access/42: dial tcp: lookup fakeuri: no such host",
			userID:    "42",
		},
		{
			name:      "status code <> 200",
			serverURI: func(s *httptest.Server) string { return s.URL },
			handler:   func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(404) },
			resultErr: "unable to retrieve access data '404 Not Found', status 404",
			userID:    "42",
		},
		{
			name:      "read response json error",
			serverURI: func(s *httptest.Server) string { return s.URL },
			handler:   func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) },
			resultErr: "unable to unmarshal result from authorization service: unexpected end of JSON input",
			userID:    "42",
		},
		{
			name:      "not success flow (status <> success)",
			serverURI: func(s *httptest.Server) string { return s.URL },
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				fmt.Fprint(w, `{"Message":"fail message","Status":"fail"}`)
			},
			resultErr: "fail message",
			userID:    "42",
		},
		{
			name:      "not success flow (status <> success)",
			serverURI: func(s *httptest.Server) string { return s.URL },
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				fmt.Fprint(w, `{"Message":"","Status":"success", "Data":{"SuperUser":false}}`)
			},
			result: &Access{},
			userID: "42",
		},
	}

	for _, testCase := range testCases {
		server := httptest.NewServer(http.HandlerFunc(testCase.handler))
		defer server.Close()

		var buf bytes.Buffer
		logger := zerolog.New(&buf)
		access, err := LoadAccess(testCase.serverURI(server), testCase.userID, "test-token", &logger)

		if testCase.result != nil {
			expected, err := json.Marshal(testCase.result)
			assert.NoError(t, err)
			actual, err := json.Marshal(access)
			assert.NoError(t, err)
			assert.Equal(t, string(expected), string(actual), testCase.name)
		} else {
			assert.Nil(t, access)
		}
		if testCase.resultErr == "" {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, testCase.resultErr, err.Error(), testCase.name)
		}
	}

}
