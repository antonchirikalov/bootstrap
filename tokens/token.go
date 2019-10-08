package tokens

import "github.com/dgrijalva/jwt-go"

// RetrieveKID searches for valid `kid` parameter in the token header
func RetrieveKID(headers map[string]interface{}) (string, error) {
	if kid, ok := headers["kid"]; !ok {
		return "", jwt.ErrInvalidKey
	} else if kidstr, ok := kid.(string); ok && kidstr != "" {
		return kidstr, nil
	}
	return "", jwt.ErrInvalidKey
}
