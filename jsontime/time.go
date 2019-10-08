package jsontime

import (
	"fmt"
	"strings"
	"time"
)

// DateTime is platform-wide type for datetime fields which are supplied or returned as JSON
type DateTime time.Time

// DateTimeFormat matches RFC3339 for date times
const DateTimeFormat = time.RFC3339

// UnmarshalJSON converts a DateTimeFormat'ed string into a DateTime field
func (t *DateTime) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)
	dt, err := time.Parse(DateTimeFormat, str)
	if err != nil {
		return err
	}
	*t = DateTime(dt)
	return nil
}

// MarshalJSON converts a DateTime into a byte slice for serialization
func (t *DateTime) MarshalJSON() ([]byte, error) {
	ts := fmt.Sprintf("\"%s\"", t.String())
	return []byte(ts), nil
}

// String converts DateTime into DateTimeFormat string
func (t *DateTime) String() string {
	return t.Time().UTC().Format(DateTimeFormat)
}

// Time satisfies the TimeValue interface
func (t *DateTime) Time() time.Time {
	return time.Time(*t)
}
