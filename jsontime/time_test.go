package jsontime

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDateTime_MarshalJSON(t *testing.T) {

	var (
		now      = time.Now().UTC()
		dateTime = DateTime(now)
		datePtr  = &dateTime
	)

	t.Run("Marshals time.Time in correct DateTime format", func(t *testing.T) {

		got, err := datePtr.MarshalJSON()
		assert.NoError(t, err)
		assert.Equal(t, `"`+now.Format(DateTimeFormat)+`"`, string(got))

	})

}

func TestDateTime_UnmarshalJSON(t *testing.T) {

	var (
		now      = time.Now().UTC()
		dateTime = DateTime(now)
		datePtr  = &dateTime
	)

	t.Run("Unmarshals time.Time in correct DateTime format", func(t *testing.T) {

		err := datePtr.UnmarshalJSON([]byte(`"` + now.Format(DateTimeFormat) + `"`))

		assert.NoError(t, err)
		assert.Equal(t, now.Format(DateTimeFormat), time.Time(*datePtr).Format(DateTimeFormat))

	})

	t.Run("Errors if value is not any sort of time value", func(t *testing.T) {

		err := datePtr.UnmarshalJSON([]byte(`"hello"`))

		assert.Error(t, err)

	})

	t.Run("json.Unmarshal converts to correct date format", func(t *testing.T) {

		type foo struct {
			Bar DateTime `json:"bar"`
		}

		f := foo{Bar: dateTime}

		err := json.Unmarshal([]byte(`{"bar":"`+now.Format(DateTimeFormat)+`"}`), &f)

		assert.NoError(t, err)
		assert.Equal(t, now.Format(DateTimeFormat), time.Time(f.Bar).Format(DateTimeFormat))

	})

	t.Run("json.Unmarshal converts pointer to correct date format", func(t *testing.T) {

		type foo struct {
			Bar *DateTime `json:"bar"`
		}

		f := foo{Bar: &dateTime}

		err := json.Unmarshal([]byte(`{"bar":"`+now.Format(DateTimeFormat)+`"}`), &f)

		assert.NoError(t, err)
		assert.Equal(t, now.Format(DateTimeFormat), time.Time(*f.Bar).Format(DateTimeFormat))

	})

}

func TestDateTime_String(t *testing.T) {

	var (
		now     = time.Now().UTC()
		date    = DateTime(now)
		datePtr = &date
	)

	t.Run("Converts time.Time in correct date format", func(t *testing.T) {

		got := datePtr.String()
		assert.Equal(t, now.Format(DateTimeFormat), got)

	})

}

func TestDateTime_Time(t *testing.T) {

	var (
		now     = time.Now()
		date    = DateTime(now)
		datePtr = &date
	)

	t.Run("Converts Date into time.Time", func(t *testing.T) {

		got := datePtr.Time()
		assert.Equal(t, now.Format(DateTimeFormat), time.Time(got).Format(DateTimeFormat))

	})

}
