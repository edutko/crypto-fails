package route

import (
	"crypto/aes"
	"crypto/sha256"
	"errors"
	"io"
	"net/http"
	"path"

	"github.com/edutko/crypto-fails/internal/config"
	"github.com/edutko/crypto-fails/internal/crypto"
	"github.com/edutko/crypto-fails/internal/crypto/random"
	"github.com/edutko/crypto-fails/internal/middleware"
	"github.com/edutko/crypto-fails/internal/route/responses"
	"github.com/edutko/crypto-fails/internal/store"
	"github.com/edutko/crypto-fails/internal/store/constants"
	"github.com/edutko/crypto-fails/internal/stores"
	"github.com/edutko/crypto-fails/pkg/share"
)

func PostUpload(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)
	uploadFile(s.Username, w, r, true)
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
		if ew, err := newEncryptingWriter(namespace, fw); err != nil {
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

func newEncryptingWriter(kid string, w io.WriteCloser) (io.WriteCloser, error) {
	k, err := getOrCreateKey(kid)
	iv := sha256.Sum256([]byte(kid))

	switch config.FileEncryptionMode() {
	case crypto.ModeGCM:
		return crypto.GCMEncrypter(k, iv[:crypto.GCMNonceSize], w), err
	default:
		return crypto.CTREncrypter(k, iv[:aes.BlockSize], w), err
	}
}

func getOrCreateKey(kid string) ([]byte, error) {
	kid = path.Join(constants.BlobStoreKIDPrefix, kid)

	if k, err := stores.KeyStore().Get(kid); err == nil {
		return k, nil
	}

	if _, err := stores.KeyStore().PutIfNotExists(kid, random.Bytes(32)); err != nil {
		return nil, err
	}

	if k, err := stores.KeyStore().Get(kid); err != nil {
		return nil, err
	} else {
		return k, nil
	}
}
