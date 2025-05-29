package route

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/edutko/crypto-fails/internal/crypto"
	"github.com/edutko/crypto-fails/internal/middleware"
	"github.com/edutko/crypto-fails/internal/responses"
	"github.com/edutko/crypto-fails/internal/store"
	"github.com/edutko/crypto-fails/internal/store/constants"
	"github.com/edutko/crypto-fails/internal/stores"
)

func Pubkeys(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)
	if r.Method == http.MethodGet {
		getPubkeysForUser(w, s.Username)

	} else if r.Method == http.MethodPost {
		if err := r.ParseMultipartForm(10 * 1024 * 1024); err != nil {
			responses.BadRequest(w, err)
			return
		}

		f, _, err := r.FormFile("file")
		if err != nil {
			responses.InternalServerError(w, err)
			return
		}

		keyBytes, err := io.ReadAll(f)
		if err != nil {
			responses.InternalServerError(w, err)
			return
		}

		kid, err := crypto.GetGPGKeyId(keyBytes)
		if err != nil {
			responses.BadRequest(w, err)
			return
		}

		if err = stores.KeyStore().Put(path.Join(constants.PubkeysByIdPrefix, kid), keyBytes); err != nil {
			responses.InternalServerError(w, err)
			return
		}

		if err = stores.KeyStore().Put(path.Join(constants.PubkeysByUserPrefix, s.Username, kid), keyBytes); err != nil {
			responses.InternalServerError(w, err)
			return
		}

		responses.Created(w, path.Join("api", "keys", url.PathEscape(kid)))

	} else {
		responses.MethodNotAllowed(w)
	}
}

func Pubkey(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		kid := r.PathValue("id")
		if pk, err := stores.KeyStore().Get(path.Join(constants.PubkeysByIdPrefix, kid)); err != nil {
			responses.InternalServerError(w, err)
		} else {
			responses.DownloadFromReader(w, kid+".pub", bytes.NewReader(pk))
		}
	} else {
		responses.MethodNotAllowed(w)
	}
}

func getPubkeysForUser(w http.ResponseWriter, username string) {
	prefix := path.Join(constants.PubkeysByUserPrefix, username) + "/"
	if pks, err := stores.KeyStore().ListKeysWithPrefix(prefix); errors.Is(err, store.ErrNotFound) {
		responses.NotFound(w)
	} else if err != nil {
		responses.InternalServerError(w, err)
	} else {
		ids := make([]string, 0)
		for _, pth := range pks {
			ids = append(ids, strings.TrimPrefix(pth, prefix))
		}
		responses.JSON(w, struct {
			Keys []string `json:"keys"`
		}{ids})
	}
}
