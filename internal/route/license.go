package route

import (
	"io"
	"net/http"

	"github.com/edutko/crypto-fails/internal/app"
	"github.com/edutko/crypto-fails/internal/route/responses"
)

func PostLicense(w http.ResponseWriter, r *http.Request) {
	f, _, err := r.FormFile("file")
	if err != nil {
		responses.InternalServerError(w, err)
		return
	}

	b, err := io.ReadAll(f)
	if err != nil {
		responses.InternalServerError(w, err)
		return
	}

	if err = app.ApplyLicense(b); err != nil {
		responses.BadRequest(w, err)
	} else {
		responses.Found(w, "/admin")
	}
}
