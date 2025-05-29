package route

import (
	"errors"
	"io"
	"net/http"
	"path"

	"github.com/edutko/crypto-fails/internal/config"
	"github.com/edutko/crypto-fails/internal/middleware"
	"github.com/edutko/crypto-fails/internal/responses"
	"github.com/edutko/crypto-fails/internal/share"
	"github.com/edutko/crypto-fails/internal/store"
	"github.com/edutko/crypto-fails/internal/stores"
)

func Upload(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)
	if r.Method == http.MethodPost {
		uploadFile(s.Username, w, r, true)
	} else {
		responses.MethodNotAllowed(w)
	}
}

func uploadFile(namespace string, w http.ResponseWriter, r *http.Request, interactive bool) {
	if err := r.ParseMultipartForm(config.MaxFileSize()); err != nil {
		responses.BadRequest(w, err)
		return
	}

	f, h, err := r.FormFile("file")
	if err != nil {
		responses.InternalServerError(w, err)
		return
	}

	fw, err := stores.FileStore().PutObject(path.Join(namespace, h.Filename))
	if errors.Is(err, store.ErrNotFound) {
		responses.NotFound(w)

	} else if err != nil {
		responses.InternalServerError(w, err)

	} else {
		defer fw.Close()
		if ew, err := stores.FileStore().NewEncryptingWriter(namespace, fw); err != nil {
			responses.InternalServerError(w, err)
		} else {
			_, _ = io.Copy(ew, f)
			if interactive {
				responses.Found(w, "/files")
			} else {
				responses.Created(w, "/download?"+share.NewLink(h.Filename, share.DoesNotExpire).QueryString())
			}
		}
	}
}
