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
	"github.com/edutko/crypto-fails/internal/route/responses"
	"github.com/edutko/crypto-fails/internal/store"
	"github.com/edutko/crypto-fails/internal/store/constants"
	"github.com/edutko/crypto-fails/internal/stores"
	"github.com/edutko/crypto-fails/pkg/api"
)

func GetPubkeys(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)
	getPubkeysForUser(s.Username, w)
}

func PostPubkeys(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)

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
}

func GetPubkey(w http.ResponseWriter, r *http.Request) {
	kid := r.PathValue("id")
	if pk, err := stores.KeyStore().Get(path.Join(constants.PubkeysByIdPrefix, kid)); err != nil {
		responses.InternalServerError(w, err)
	} else {
		responses.DownloadFromReader(w, kid+".pub", bytes.NewReader(pk))
	}
}

func getPubkeysForUser(username string, w http.ResponseWriter) {
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
		responses.JSON(w, api.PubkeysResponse{Keys: ids})
	}
}
