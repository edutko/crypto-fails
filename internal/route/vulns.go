package route

import (
	"errors"
	"io"
	"net/http"

	"github.com/edutko/crypto-fails/internal/config"
	"github.com/edutko/crypto-fails/internal/route/responses"
	"github.com/edutko/crypto-fails/internal/store"
	"github.com/edutko/crypto-fails/internal/stores"
)

func TweakCiphertext(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	if err := r.ParseMultipartForm(config.MaxFileSize()); err != nil {
		responses.BadRequest(w, err)
		return
	}

	f, _, err := r.FormFile("file")
	if err != nil {
		responses.InternalServerError(w, err)
		return
	}

	if r.Method == http.MethodPut {
		// This looks a lot like uploadFile(), but it does NOT encrypt the file.
		// We want an attacker to be able to upload chosen ciphertext as-is.
		if fw, err := stores.FileStore().PutObject(key); errors.Is(err, store.ErrNotFound) {
			responses.NotFound(w)
		} else if err != nil {
			responses.InternalServerError(w, err)
		} else {
			_, _ = io.Copy(fw, f)
			responses.Accepted(w)
		}
	} else {
		responses.MethodNotAllowed(w)
	}
}
