package route

import (
	"net/http"

	"github.com/edutko/crypto-fails/internal/config"
	"github.com/edutko/crypto-fails/internal/responses"
)

func JWKS(w http.ResponseWriter, r *http.Request) {
	if body, err := config.JWKS().JSONPublic(r.Context()); err != nil {
		responses.InternalServerError(w, err)
	} else {
		responses.OkJSONBytes(w, body)
	}
}
