package jsontime

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDate_MarshalJSON(t *testing.T) {

	var (
		now     = time.Now()
		date    = Date(now)
		datePtr = &date
	)

	t.Run("Marshals time.Time in correct date format", func(t *testing.T) {

		got, err := datePtr.MarshalJSON()
		assert.NoError(t, err)
		assert.Equal(t, `"`+now.Format(DateFormat)+`"`, string(got))

	})

}

func TestDate_UnmarshalJSON(t *testing.T) {

	var (
		now     = time.Now()
		date    = Date(now)
		datePtr = &date
	)

	t.Run("Unmarshals time.Time in correct date format", func(t *testing.T) {

		err := datePtr.UnmarshalJSON([]byte(`"` + now.Format(DateFormat) + `"`))

		assert.NoError(t, err)

		y1, m1, d1 := now.Date()
		y2, m2, d2 := time.Time(*datePtr).Date()

		assert.Equal(t, y1, y2)
		assert.Equal(t, m1, m2)
		assert.Equal(t, d1, d2)

	})

	t.Run("Errors if value is not any sort of time value", func(t *testing.T) {

		err := datePtr.UnmarshalJSON([]byte(`"hello"`))

		assert.Error(t, err)

	})

	t.Run("json.Unmarshal converts to correct date format", func(t *testing.T) {

		type foo struct {
			Bar Date `json:"bar"`
		}

		f := foo{Bar: date}

		err := json.Unmarshal([]byte(`{"bar":"`+now.Format(DateFormat)+`"}`), &f)

		assert.NoError(t, err)

		y1, m1, d1 := now.Date()
		y2, m2, d2 := time.Time(f.Bar).Date()

		assert.Equal(t, y1, y2)
		assert.Equal(t, m1, m2)
		assert.Equal(t, d1, d2)

	})

	t.Run("json.Unmarshal converts pointer to correct date format", func(t *testing.T) {

		type foo struct {
			Bar *Date `json:"bar"`
		}

		f := foo{Bar: &date}

		err := json.Unmarshal([]byte(`{"bar":"`+now.Format(DateFormat)+`"}`), &f)

		assert.NoError(t, err)

		y1, m1, d1 := now.Date()
		y2, m2, d2 := time.Time(*f.Bar).Date()

		assert.Equal(t, y1, y2)
		assert.Equal(t, m1, m2)
		assert.Equal(t, d1, d2)

	})

}

func TestDate_String(t *testing.T) {

	var (
		now     = time.Now()
		date    = Date(now)
		datePtr = &date
	)

	t.Run("Converts time.Time in correct date format", func(t *testing.T) {

		got := datePtr.String()
		assert.Equal(t, now.Format(DateFormat), got)

	})

}

func TestDate_Time(t *testing.T) {

	var (
		now     = time.Now()
		date    = Date(now)
		datePtr = &date
	)

	t.Run("Converts Date into time.Time", func(t *testing.T) {

		got := datePtr.Time()

		y1, m1, d1 := now.Date()
		y2, m2, d2 := time.Time(got).Date()

		assert.Equal(t, y1, y2)
		assert.Equal(t, m1, m2)
		assert.Equal(t, d1, d2)

	})

}
