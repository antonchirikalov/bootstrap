package response

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSend(t *testing.T) {
	tests := []struct {
		name    string
		sr      *StandardResponse
		cookies []*http.Cookie
		wantErr bool
	}{
		{
			name: "sets headers",
			sr: &StandardResponse{
				StatusCode: 200,
				headers: map[string][]string{
					"Header-A": {"value-A"},
					"Header-B": {"value-B1", "value-B2"},
				},
			},
		},
		{
			name: "sets cookies",
			sr: &StandardResponse{
				StatusCode: 200,
			},
			cookies: []*http.Cookie{
				{
					Name:    "cookie1",
					Value:   "1",
					Expires: time.Now().Add(time.Hour),
				},
				{
					Name:    "cookie2",
					Value:   "2",
					Expires: time.Now().Add(time.Hour),
				},
			},
		},
	}
	for _, tt := range tests {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if tt.cookies != nil {
				for _, c := range tt.cookies {
					tt.sr.AddCookie(c)
				}
			}

			err := Send(w, tt.sr)
			if err != nil && !tt.wantErr {
				assert.Fail(t, err.Error())
			}
		}))

		cli := server.Client()
		resp, err := cli.Get(server.URL)
		if err != nil {
			assert.Error(t, err, "get failed")
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			assert.Error(t, err, "decode failed")
		}
		resp.Body.Close()

		var received StandardBody
		err = json.Unmarshal(b, &received)
		if err != nil {
			assert.Error(t, err, "unmarshal failed")
		}

		headers := resp.Header

		if len(tt.sr.headers) != 0 {
			for k, v := range tt.sr.headers {
				assert.Equal(t, v, headers[http.CanonicalHeaderKey(k)], "header missing/invalid")
			}
		}

		if len(tt.cookies) != 0 {
			cookies := resp.Cookies()

			assert.Equal(t, len(tt.cookies), len(cookies))
			for _, wantCookie := range tt.cookies {
				found := false

				for _, gotCookie := range cookies {
					if wantCookie.Name == gotCookie.Name &&
						wantCookie.Value == gotCookie.Value {
						found = true
						break
					}
				}

				if !found {
					assert.Failf(t, "cookie not found in response", wantCookie.Name)
				}
			}
		}

		server.Close()
	}
}
