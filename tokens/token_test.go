package tokens

import (
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

func TestRetrieveKID(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]interface{}
		want    string
		wantErr error
	}{
		{
			name:    "No `kid` in the header",
			args:    map[string]interface{}{},
			want:    "",
			wantErr: jwt.ErrInvalidKey,
		},
		{
			name:    "Empty `kid` in the header",
			args:    map[string]interface{}{"kid": ""},
			want:    "",
			wantErr: jwt.ErrInvalidKey,
		},
		{
			name:    "successful case",
			args:    map[string]interface{}{"kid": "mykid"},
			want:    "mykid",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		got, err := RetrieveKID(tt.args)
		assert.Equal(t, tt.want, got, tt.name)
		assert.Equal(t, tt.wantErr, err, tt.name)
	}
}
