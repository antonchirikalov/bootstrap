package jsontime

import (
	"reflect"
	"time"
)

// TimeValue provides an interface for conversion of DateTime
type TimeValue interface {
	Time() time.Time
}

// TimerValuer provides a value for validation
func TimerValuer(field reflect.Value) interface{} {
	if t, ok := field.Interface().(TimeValue); ok {
		return t.Time()
	}
	return nil
}
