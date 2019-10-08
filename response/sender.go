package response

import (
	"encoding/json"
	"errors"
	"net/http"
)

// Send writes a StandardResponse to the given ResponseWriter
func Send(w http.ResponseWriter, sr *StandardResponse) error {
	body := sr.BuildBody()

	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	for k, v := range sr.headers {
		for _, value := range v {
			w.Header().Add(k, value)
		}
	}

	if sr.StatusCode == 0 {
		return errors.New("status code must be set")
	}

	w.WriteHeader(sr.StatusCode)

	_, err = w.Write(b)
	if err != nil {
		return err
	}

	return nil
}
