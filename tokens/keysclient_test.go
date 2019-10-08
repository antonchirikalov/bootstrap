package tokens

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRetrievePublicKey(t *testing.T) {

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	pemKey, err := json.Marshal(convertPublicRSA(&key.PublicKey))
	if err != nil {
		panic(err)
	}

	pemEncodedKey := string(pemKey)

	testCases := []struct {
		name      string
		serverURI func(*httptest.Server) string
		handler   func(rw http.ResponseWriter, req *http.Request)
		result    *rsa.PublicKey
		resultErr string
		keyID     string
	}{
		{
			name:      "unable to parse url",
			serverURI: func(*httptest.Server) string { return "_http://failme.co" },
			handler:   func(w http.ResponseWriter, r *http.Request) {},
			resultErr: "unable to parse keys server url: parse _http://failme.co: first path segment in URL cannot contain colon",
			keyID:     "42",
		},
		{
			name:      "request failed",
			serverURI: func(*httptest.Server) string { return "http://fakeuri" },
			handler:   func(w http.ResponseWriter, r *http.Request) {},
			resultErr: "unable to fetch key from keys server: Get http://fakeuri/keys/42: dial tcp: lookup fakeuri: no such host",
			keyID:     "42",
		},
		{
			name:      "status code <> 200",
			serverURI: func(s *httptest.Server) string { return s.URL },
			handler:   func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(404) },
			resultErr: "unable to retrieve license '404 Not Found', status 404",
			keyID:     "42",
		},
		{
			name:      "read response json error",
			serverURI: func(s *httptest.Server) string { return s.URL },
			handler:   func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) },
			resultErr: "unable to unmarshal result from keys server: unexpected end of JSON input",
			keyID:     "42",
		},
		{
			name:      "unexpected response body",
			serverURI: func(s *httptest.Server) string { return s.URL },
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				fmt.Fprint(w, `{"key":"123"}`)
			},
			resultErr: `unexpected response body from keys server: {"key":"123"}`,
			keyID:     "42",
		},
		{
			name:      "invalid key",
			serverURI: func(s *httptest.Server) string { return s.URL },
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				fmt.Fprint(w, `{"message":"","status":"success","data":{"key": "123"}}`)
			},
			resultErr: `Invalid Key: Key must be PEM encoded PKCS1 or PKCS8 private key`,
			keyID:     "42",
		},
		{
			name:      "happy path",
			serverURI: func(s *httptest.Server) string { return s.URL },
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				res := fmt.Sprintf(`{"message":"","status":"success","data":{"key":%s}}`, pemEncodedKey)
				fmt.Fprint(w, res)
			},
			result: &key.PublicKey,
			keyID:  "42",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(testCase.handler))
			publicKey, err := RetrievePublicKey(testCase.serverURI(server), testCase.keyID)

			if testCase.result != nil {
				expected, err := json.Marshal(testCase.result)
				assert.NoError(t, err)
				actual, err := json.Marshal(publicKey)
				assert.NoError(t, err)
				assert.Equal(t, string(expected), string(actual), testCase.name)
			} else {
				assert.Nil(t, publicKey)
			}
			if testCase.resultErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, testCase.resultErr, err.Error(), testCase.name)
			}

			server.Close()
		})
	}

}

func TestUnmarshalJSON(t *testing.T) {
	// error
	time := &jsonTime{}
	err := time.UnmarshalJSON([]byte("ERROR-TIME"))
	assert.EqualError(t, err, "parsing time \"ERROR-TIME\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"ERROR-TIME\" as \"2006\"")
	// success
	err = time.UnmarshalJSON([]byte("2006-01-02T15:04:05Z"))
	assert.Nil(t, err)
}

// creates a PEM encoded public key
func convertPublicRSA(key *rsa.PublicKey) string {

	pubASN1, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		panic(err)
	}

	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})

	return string(pubBytes)
}
