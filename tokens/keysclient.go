package tokens

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"bitbucket.org/teachingstrategies/go-svc-bootstrap/response"
	"github.com/dgrijalva/jwt-go"
)

// RetrievePublicKey retrieves the public key associated with the given keyID from the key server
func RetrievePublicKey(keysServerURL, keyID string) (*rsa.PublicKey, error) {
	pubKey, err := loadPublicKey(keyID, keysServerURL)
	if err != nil {
		return nil, err
	}
	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pubKey.Key))
	if err != nil {
		return nil, err
	}
	return verifyKey, nil
}

func loadPublicKey(keyID string, keyServerURL string) (*PublicKey, error) {
	serverURL, err := url.Parse(keyServerURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse keys server url: %v", err)
	}
	serverURL.Path = path.Join(serverURL.Path, "keys", keyID)
	resp, err := http.Get(serverURL.String())
	if err != nil {
		return nil, fmt.Errorf("unable to fetch key from keys server: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to retrieve license '%s', status %d", resp.Status, resp.StatusCode)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response from keys server: %v", err)
	}
	var keyResponse PublicKeyResponse
	err = json.Unmarshal(b, &keyResponse)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal result from keys server: %v", err)
	}
	if keyResponse.Data == nil || keyResponse.Data.Key == "" {
		return nil, fmt.Errorf("unexpected response body from keys server: %v", string(b))
	}
	return keyResponse.Data, nil
}

// PublicKeyResponse represents the expected response from key server
type PublicKeyResponse struct {
	response.StandardBody
	Data *PublicKey `json:"data"`
}

// PublicKey struct represents public key information
type PublicKey struct {
	KeyID     string   `json:"keyID"`
	CreatedOn jsonTime `json:"createdOn"`
	SourceID  string   `json:"sourceID"`
	Key       string   `json:"key"`
}

type jsonTime time.Time

func (t *jsonTime) UnmarshalJSON(bytes []byte) error {
	str := strings.Trim(string(bytes), `"`)
	effOn, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return err
	}
	*t = jsonTime(effOn)
	return nil
}
