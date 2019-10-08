package jsontime

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var now = time.Now().UTC()

type tv string

func (t tv) Time() time.Time {
	return now
}

func TestTimerValuer(t *testing.T) {
	tests := []struct {
		name         string
		reflectValue reflect.Value
		want         interface{}
	}{
		{
			name: "Converts TimeValue reflect.Value to time.Time",
			reflectValue: func() reflect.Value {
				t := tv("test")
				return reflect.ValueOf(t)
			}(),
			want: now,
		},
		{
			name: "Does not convert string reflect.Value to time.Time",
			reflectValue: func() reflect.Value {
				t := "test"
				return reflect.ValueOf(t)
			}(),
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TimerValuer(tt.reflectValue)
			assert.Equal(t, tt.want, got)
		})
	}
}
