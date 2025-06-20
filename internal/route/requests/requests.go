package requests

import (
	"context"
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

func WithInteractiveLabel(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), ctxKeyInteractive, true))
}

func IsInteractive(r *http.Request) bool {
	return r != nil && r.Context().Value(ctxKeyInteractive) != nil && r.Context().Value(ctxKeyInteractive).(bool)
}

const (
	ctxKeyInteractive = "interactive"
)
