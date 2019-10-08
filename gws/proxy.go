package gws

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/rs/zerolog"
)

// Proxy provides an interface for interacting with GWS
type Proxy interface {
	Post(signedIam *string, endpoint string, payload []byte) (int, []byte, error)
	Get(signedIam *string, endpoint string, queryString *url.Values) (int, []byte, error)
}

type proxy struct {
	config gwsConfig
	logger *zerolog.Logger
}

type gwsConfig struct {
	baseURI      *url.URL
	authToken    string
	sharedSecret []byte
}

// StandardResponse represents a standard GWS response
type StandardResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

var (
	ErrProxyRequestFailed  = errors.New("invalid status from gws")
	ErrProxyMisconfigured  = errors.New("proxy misconfigured")
	ErrProxyInvalidRequest = errors.New("invalid request")
)

// NewProxy creates a proxy configured and ready to use
func NewProxy(baseURI string, authToken string, sharedSecret []byte, logger *zerolog.Logger) (Proxy, error) {

	if baseURI == "" || authToken == "" || len(sharedSecret) == 0 || logger == nil {
		return nil, ErrProxyMisconfigured
	}

	uri, err := url.Parse(baseURI)
	if err != nil {
		return nil, err
	}

	return &proxy{
		config: gwsConfig{
			baseURI:      uri,
			authToken:    authToken,
			sharedSecret: sharedSecret,
		},
		logger: logger,
	}, nil

}

// Post issues an HTTP Post to GWS
func (p *proxy) Post(signedIam *string, endpoint string, payload []byte) (status int, jsn []byte, err error) {
	return p.makeRequest(signedIam, http.MethodPost, endpoint, bytes.NewReader(payload), &url.Values{})
}

// Get issues an HTTP Get to GWS
func (p *proxy) Get(signedIam *string, endpoint string, queryString *url.Values) (status int, jsn []byte, err error) {
	return p.makeRequest(signedIam, http.MethodGet, endpoint, nil, queryString)
}

func (p *proxy) makeRequest(signedIam *string, method string, endpoint string, body io.Reader, queryString *url.Values) (status int, jsn []byte, err error) {

	status = http.StatusInternalServerError

	if p.config.baseURI == nil {
		p.logger.Error().Err(err).Msg("p baseURI is nil")
		return status, jsn, ErrProxyMisconfigured
	}

	uri := buildURI(p.config.baseURI, endpoint, queryString)

	request, err := http.NewRequest(method, uri, body)
	if err != nil {
		p.logger.Error().Err(err).Str("uri", uri).Msg("could not create request")
		return status, jsn, ErrProxyInvalidRequest
	}

	addHeaders(request, signedIam, p.config.authToken)

	status, resBody, err := send(request, p.logger)
	if err != nil {
		p.logger.Error().Err(err).Str("uri", uri).Msg("could not send request")
		return status, jsn, err
	}

	return status, resBody, err

}

func buildURI(baseURI *url.URL, endpoint string, queryString *url.Values) string {
	var builder strings.Builder
	builder.WriteString(baseURI.String())
	builder.WriteString(endpoint)
	if queryString != nil {
		qs := queryString.Encode()
		if qs != "" {
			builder.WriteString("?")
			builder.WriteString(qs)
		}
	}
	return builder.String()
}

func send(request *http.Request, logr *zerolog.Logger) (statusCode int, body []byte, err error) {

	res, err := http.DefaultClient.Do(request)
	if err != nil || res == nil {
		return 0, body, err
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			logr.Error().Err(err).Str("uri", request.URL.String()).Str("verb", request.Method).Msg("could not close res")
		}
	}()

	if res.StatusCode == http.StatusInternalServerError {
		return http.StatusInternalServerError, nil, ErrProxyRequestFailed
	}

	if body, err = ioutil.ReadAll(res.Body); err != nil {
		return res.StatusCode, body, err
	}

	return res.StatusCode, body, nil

}

func addHeaders(request *http.Request, signedIam *string, authToken string) {

	request.Header.Add("Authorization", authToken)
	request.Header.Add("Content-Type", "application/json")

	if signedIam != nil {
		request.Header.Add("Iam", *signedIam)
	}

}
