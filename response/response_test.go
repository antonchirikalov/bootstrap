package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/go-playground/validator.v9"
)

type validationFailure struct {
	id int
}

func (v validationFailure) Field() string {
	return fmt.Sprintf("Field: %v", v.id)
}
func (v validationFailure) Tag() string {
	return fmt.Sprintf("Tag: %v", v.id)
}
func (v validationFailure) Value() interface{} {
	return fmt.Sprintf("Value: %v", v.id)
}
func (v validationFailure) Param() string {
	return fmt.Sprintf("Param: %v", v.id)
}

func TestNew(t *testing.T) {
	response := New()

	//Sets default headers
	assert.Equalf(
		t,
		map[string][]string{"Content-Type": {"application/json"}},
		response.headers,
		"%v: Actual response should equal expectation",
		"sets default header(s)",
	)

}

func TestNewWithCors(t *testing.T) {

	response := NewWithCors("test")

	// Sets cors and default headers
	assert.Equalf(
		t,
		map[string][]string{
			"Content-Type":                     {"application/json"},
			"Access-Control-Allow-Origin":      {"test"},
			"Access-Control-Allow-Methods":     {CorsAllowedMethods},
			"Access-Control-Allow-Headers":     {CorsAllowedHeaders},
			"Access-Control-Allow-Credentials": {"true"},
		},
		response.headers,
		"%v: Actual response should equal expectation",
		"sets access control headers",
	)

}

func TestStandardResponseAddCorsHeaders(t *testing.T) {

	res := &StandardResponse{}
	res.AddCorsHeaders("test")

	assert.Equal(t, map[string][]string{
		"Access-Control-Allow-Origin":      {"test"},
		"Access-Control-Allow-Methods":     {CorsAllowedMethods},
		"Access-Control-Allow-Headers":     {CorsAllowedHeaders},
		"Access-Control-Allow-Credentials": {"true"},
	}, res.headers, "adds cors-related headers: Actual response should equal expectation")

}

func TestStandardResponseAddHeader(t *testing.T) {

	res := &StandardResponse{}
	res.AddHeader("foo-bar-foo", "bar")

	assert.Equal(t, map[string][]string{"Foo-Bar-Foo": {"bar"}}, res.headers, "canonicalizes header: Actual response should equal expectation")

}

func TestStandardResponseHeaders(t *testing.T) {

	res := &StandardResponse{}
	res.AddHeader("foo-bar-foo", "bar") // r.header is unexported, so we can't test res.GetHeaders() in isolation

	assert.Equal(t, map[string][]string{"Foo-Bar-Foo": {"bar"}}, res.GetHeaders(), "canonicalizes header: Actual response should equal expectation")

}

func TestProcessValidatorV9Failures(t *testing.T) {
	res := &StandardResponse{}
	// empty failures
	res.ProcessValidatorV9Failures(validator.ValidationErrors{})
	// a couple of failures
	type testStruct struct {
		Field1 int    `validate:"required,gte=1,lte=6"`
		Field2 string `validate:"required"`
	}
	err := validator.New().Struct(&testStruct{Field1: 10})
	assert.Error(t, err)
	res.ProcessValidatorV9Failures(err.(validator.ValidationErrors))
	assert.NotNil(t, res.validationFailures, "failures must be not empty")
	assert.Equal(t, 2, len(res.validationFailures), "must be 2 failures")
}

func TestStandardResponseAddCookie(t *testing.T) {

	simple := New()
	simple.AddCookie(&http.Cookie{Name: "foo", Value: "bar"})

	// Converts cookie into header
	assert.Equalf(
		t,
		[]string{"foo=bar"},
		simple.headers["Set-Cookie"],
		"%v: Actual response should equal expectation",
		"converts cookie into header",
	)

	now := time.Now()

	kitchenSink := New()
	kitchenSink.AddCookie(&http.Cookie{
		Name:       "foo",
		Value:      "bar",
		Path:       "baz",
		Domain:     "foobar",
		Expires:    now,
		RawExpires: now.String(),
		MaxAge:     1,
		Secure:     true,
		HttpOnly:   true,
		SameSite:   http.SameSite(1),
		Raw:        "test",
		Unparsed:   []string{"test2"},
	})

	cookie := fmt.Sprintf("foo=bar; Path=baz; Domain=foobar; Expires=%s; Max-Age=1; HttpOnly; Secure; SameSite", now.UTC().Format(http.TimeFormat))

	// Converts cookie into header
	assert.Equalf(
		t,
		[]string{cookie},
		kitchenSink.headers["Set-Cookie"],
		"%v: Actual response should equal expectation",
		"converts complex cookie into header",
	)

	tomorrow := now.AddDate(0, 0, 1)
	kitchenSink.AddCookie(&http.Cookie{
		Name:       "foo1",
		Value:      "bar1",
		Path:       "baz1",
		Domain:     "foobar1",
		Expires:    tomorrow,
		RawExpires: tomorrow.String(),
		MaxAge:     2,
		Secure:     false,
		HttpOnly:   false,
		SameSite:   http.SameSite(1),
		Raw:        "test1",
		Unparsed:   []string{"test3"},
	})

	cookie2 := fmt.Sprintf("foo1=bar1; Path=baz1; Domain=foobar1; Expires=%s; Max-Age=2; SameSite", tomorrow.UTC().Format(http.TimeFormat))

	// Converts cookie into header
	assert.Equalf(
		t,
		[]string{cookie, cookie2},
		kitchenSink.headers["Set-Cookie"],
		"%v: Actual response should equal expectation",
		"converts complex cookies into header",
	)

}

func TestStandardResponseBuildBody(t *testing.T) {

	type testCase struct {
		Name     string
		Actual   *StandardResponse
		Expected *StandardBody
	}

	cases := map[string]testCase{}

	cases["empty"] = testCase{
		"empty",
		&StandardResponse{},
		&StandardBody{
			Status:  "",
			Message: "",
		},
	}

	cases["simple"] = testCase{
		"simple",
		&StandardResponse{StatusCode: 200, Message: "It worked"},
		&StandardBody{
			Status:  "success",
			Message: "It worked",
		},
	}

	cases["nested"] = testCase{
		"nested",
		&StandardResponse{
			Data: map[string]map[int]string{
				"foo": {
					1: "text",
				},
				"bar": {
					89: "text2",
				},
			},
		},
		&StandardBody{
			Status:  "",
			Message: "",
			Data: map[string]map[int]string{
				"foo": {
					1: "text",
				},
				"bar": {
					89: "text2",
				},
			},
		},
	}

	cases["pagination"] = testCase{
		"nested",
		&StandardResponse{},
		&StandardBody{},
	}
	cases["pagination"].Actual.SetPagination(&Pagination{
		Limit:        1,
		Offset:       1,
		TotalRecords: 2,
	})
	cases["pagination"].Expected.Pagination = &Pagination{
		Limit:        1,
		Offset:       1,
		TotalRecords: 2,
	}

	cases["validations"] = testCase{
		"nested",
		&StandardResponse{},
		&StandardBody{},
	}
	cases["validations"].Actual.SetValidationFailures([]ValidationFailure{
		validationFailure{1},
		validationFailure{2},
	})
	cases["validations"].Expected.Validations = &ValidationResponses{
		ValidationFailureResponse{
			"Field: 1",
			"Tag: 1",
			"Value: 1",
			"Param: 1",
		},
		ValidationFailureResponse{
			"Field: 2",
			"Tag: 2",
			"Value: 2",
			"Param: 2",
		},
	}

	cases["validations"] = testCase{
		"nested",
		&StandardResponse{},
		&StandardBody{},
	}
	cases["validations"].Actual.AddValidationFailure(validationFailure{3})
	cases["validations"].Expected.Validations = &ValidationResponses{
		ValidationFailureResponse{
			"Field: 3",
			"Tag: 3",
			"Value: 3",
			"Param: 3",
		},
	}

	for _, v := range cases {
		assert.Equalf(
			t,
			v.Expected,
			v.Actual.BuildBody(),
			"%v: Actual response should match expectation",
			v.Name,
		)

		// Check serialization
		exp, err := json.Marshal(v.Expected)
		assert.Nilf(t, err, "%v: Errored during marshalling. %v", v.Name, err)
		act, err := json.Marshal(v.Actual.BuildBody())
		assert.Nilf(t, err, "%v: Errored during marshalling. %v", v.Name, err)

		assert.Equalf(
			t,
			string(exp),
			string(act),
			"%v: Actual json should equal expectation",
			v.Name,
		)
	}

}

func TestGetStatus(t *testing.T) {

	tests := []struct {
		name string
		code int
		want string
	}{
		{
			name: "less than 1",
			code: -1,
			want: "",
		},
		{
			name: "1 to 200",
			code: 108,
			want: "info",
		},
		{
			name: "199 to 300",
			code: 222,
			want: "success",
		},
		{
			name: "299 to 300",
			code: 333,
			want: "redirect",
		},
		{
			name: "399 to 500",
			code: 444,
			want: "client error",
		},
		{
			name: "more than 500",
			code: 666,
			want: "server error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStatus(tt.code); got != tt.want {
				t.Errorf("getStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStandardResponse_GetFlattenedHeaders(t *testing.T) {
	type fields struct {
		headers map[string][]string
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]string
	}{
		{
			name: "single value headers",
			fields: fields{map[string][]string{
				"a": {"a-value"},
				"b": {"b-value"},
			}},
			want: map[string]string{
				"a": "a-value",
				"b": "b-value",
			},
		},
		{
			name: "multi-value headers are collapsed",
			fields: fields{map[string][]string{
				"a": {"a-value1", "a-value2"},
				"b": {"b-value1", "b-value2", "b-value3"},
			}},
			want: map[string]string{
				"a": "a-value1,a-value2",
				"b": "b-value1,b-value2,b-value3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &StandardResponse{
				headers: tt.fields.headers,
			}
			if got := r.GetFlattenedHeaders(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StandardResponse.GetFlattenedHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}
