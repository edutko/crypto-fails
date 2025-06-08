package requests

import (
	"encoding/json"
	"io"
	"net/http"
)

func ParseJSONBody(r *http.Request, v any) error {
	if b, err := io.ReadAll(r.Body); err != nil {
		return err

	} else if err = json.Unmarshal(b, v); err != nil {
		return err
	}

	return nil
}
