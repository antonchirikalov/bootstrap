package response

import (
	"net/http"
	"strings"

	"gopkg.in/go-playground/validator.v9"
)

const (
	// CorsAllowedHeaders is the set of allowed headers for CORS requests
	CorsAllowedHeaders = "authorization, origin, content-type"
	// CorsAllowedMethods is the set of allowed HTTP methods for CORS requests
	CorsAllowedMethods = "GET, POST, PUT, DELETE, PATCH, OPTIONS"
)

// StandardBody is the standard format for REST/JSON response bodies
type StandardBody struct {
	Message string `json:"message"`
	Status  string `json:"status"`

	Data        interface{}          `json:"data,omitempty"`
	Pagination  *Pagination          `json:"pagination,omitempty"`
	Validations *ValidationResponses `json:"validations,omitempty"`
}

// Pagination contains parameters useful for paging through a subset
type Pagination struct {
	Offset       int `json:"offset"`
	Limit        int `json:"limit"`
	TotalRecords int `json:"totalRecords"`
}

// ValidationResponses represents a collection of validation failure responses
type ValidationResponses []ValidationFailureResponse

// ValidationFailureResponse contains individual validation failure metadata
type ValidationFailureResponse struct {
	Field   string      `json:"field"`
	Tag     string      `json:"tag"`
	Value   interface{} `json:"value"`
	Allowed interface{} `json:"allowed"`
}

// ValidationFailure provides an interface for validation failures based upon gopkg.in/go-playground/validator.v9
type ValidationFailure interface {
	Field() string
	Tag() string
	Value() interface{}
	Param() string
}

// StandardResponse provides utility for creating consistent headers, payloads, pagination params and validation
type StandardResponse struct {
	StatusCode int
	Message    string
	Data       interface{}

	pagination         *Pagination
	validationFailures []ValidationFailure
	headers            map[string][]string
}

// New creates a standard response with defaults
func New() *StandardResponse {
	res := &StandardResponse{}
	res.AddHeader("Content-Type", "application/json")
	return res
}

// NewWithCors creates a standard response with defaults and cors headers
func NewWithCors(origin string) *StandardResponse {
	res := New()
	res.AddCorsHeaders(origin)
	return res
}

// AddCorsHeaders adds common CORS HTTP headers
func (r *StandardResponse) AddCorsHeaders(origin string) {
	r.AddHeader("Access-Control-Allow-Origin", origin)
	r.AddHeader("Access-Control-Allow-Methods", CorsAllowedMethods)
	r.AddHeader("Access-Control-Allow-Headers", CorsAllowedHeaders)
	r.AddHeader("Access-Control-Allow-Credentials", "true")
}

// AddHeader adds HTTP header using canonical header case
func (r *StandardResponse) AddHeader(key string, value string) {
	cleanKey := http.CanonicalHeaderKey(key)
	if r.headers == nil {
		r.headers = map[string][]string{}
	}
	if _, ok := r.headers[cleanKey]; !ok {
		r.headers[cleanKey] = []string{}
	}
	r.headers[cleanKey] = append(r.headers[cleanKey], value)
}

// GetHeaders exposes private field, headers
func (r *StandardResponse) GetHeaders() map[string][]string {
	return r.headers
}

// GetFlattenedHeaders returns multi-value headers flattened by key
func (r *StandardResponse) GetFlattenedHeaders() map[string]string {
	flatHeaders := make(map[string]string, len(r.headers))

	for k, v := range r.headers {
		flatHeaders[k] = strings.Join(v, ",")
	}

	return flatHeaders
}

// AddCookie adds a Set-Cookie header
func (r *StandardResponse) AddCookie(cookie *http.Cookie) {
	r.AddHeader("Set-Cookie", cookie.String())
}

// SetValidationFailures sets/replaces validation failures
func (r *StandardResponse) SetValidationFailures(f []ValidationFailure) {
	r.validationFailures = f
}

// ProcessValidatorV9Failures converts validator v9 errors into []ValidationFailure and sets it internally
func (r *StandardResponse) ProcessValidatorV9Failures(f validator.ValidationErrors) {
	var failures []ValidationFailure
	for _, e := range f {
		failures = append(failures, e)
	}
	r.SetValidationFailures(failures)
}

// AddValidationFailure adds validation failure
func (r *StandardResponse) AddValidationFailure(f ValidationFailure) {
	r.validationFailures = append(r.validationFailures, f)
}

// SetPagination stores pagination parameters
func (r *StandardResponse) SetPagination(p *Pagination) {
	r.pagination = p
}

// BuildBody builds a standard response body in a consistent format
func (r *StandardResponse) BuildBody() *StandardBody {
	body := &StandardBody{
		Status:  getStatus(r.StatusCode),
		Message: r.Message,
		Data:    r.Data,
	}
	validations := prepareValidationResponses(r.validationFailures)
	if len(*validations) > 0 {
		body.Validations = validations
	}
	if r.pagination != nil {
		body.Pagination = r.pagination
	}
	return body
}

func prepareValidationResponses(errs []ValidationFailure) *ValidationResponses {
	responses := make(ValidationResponses, len(errs))
	for i, err := range errs {
		responses[i] = ValidationFailureResponse{Field: err.Field(), Tag: err.Tag(), Value: err.Value(), Allowed: err.Param()}
	}

	return &responses
}

func getStatus(code int) string {
	switch {
	case code < 1:
		return ""
	case code > 0 && code < 200:
		return "info"
	case code > 199 && code < 300:
		return "success"
	case code > 299 && code < 400:
		return "redirect"
	case code > 399 && code < 500:
		return "client error"
	}
	return "server error"
}
