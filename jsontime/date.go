package jsontime

import (
	"fmt"
	"strings"
	"time"
)

// DateFormat matches the ISO standard for dates
const DateFormat = "2006-01-02"

// Date is platform-wide type for dates which are supplied or returned as JSON
type Date time.Time

// MarshallJSON ensures a Date is serialized in platform-specific format
func (d *Date) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(*d).Format(DateFormat))
	return []byte(stamp), nil
}

// UnmarshalJSON ensures a Date is serialized in platform-specific format
func (d *Date) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)
	date, err := time.Parse(DateFormat, str)
	if err != nil {
		return err
	}
	*d = Date(date)
	return nil
}

// String converts Date into DateFormat string
func (d *Date) String() string {
	return d.Time().UTC().Format(DateFormat)
}

// Time satisfies the TimeValue interface for validation
func (d *Date) Time() time.Time {
	return time.Time(*d)
}
