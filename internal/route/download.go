package route

import (
	"crypto/aes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/edutko/crypto-fails/internal/auth"
	"github.com/edutko/crypto-fails/internal/config"
	"github.com/edutko/crypto-fails/internal/crypto"
	"github.com/edutko/crypto-fails/internal/middleware"
	"github.com/edutko/crypto-fails/internal/route/responses"
	"github.com/edutko/crypto-fails/internal/store"
	"github.com/edutko/crypto-fails/internal/stores"
	"github.com/edutko/crypto-fails/pkg/share"
)

func GetDownload(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)
	l := share.ParseLink(r.URL.Query())
	err := l.Verify(auth.GetShareLinkSecret())
	if errors.Is(err, share.ErrNoSignature) && s != nil {
		// unsigned links are valid for a user's own files and are relative to the user's namespace
		if strings.HasPrefix(path.Join(s.Username, l.Key), s.Username+"/") {
			downloadFile(s.Username, l.Key, w)
		} else {
			responses.Forbidden(w, fmt.Errorf("permission denied for %q on %q", s.Username, l.Key))
		}

	} else if errors.Is(err, share.ErrInvalidSignature) || errors.Is(err, share.ErrExpired) {
		responses.Forbidden(w, fmt.Errorf("invalid share link for %q", r.PathValue("key")))

	} else if err != nil {
		responses.BadRequest(w, err)

	} else {
		// trust that the first segment of a signed link contains the namespace
		parts := strings.Split(l.Key, "/")
		downloadFile(parts[0], path.Join(parts[1:]...), w)
	}
}

func downloadFile(username, key string, w http.ResponseWriter) {
	fr, m, err := stores.FileStore().GetObject(path.Join(username, key))
	if errors.Is(err, store.ErrNotFound) {
		responses.NotFound(w)

	} else if err != nil {
		responses.InternalServerError(w, err)

	} else {
		defer fr.Close()
		if dr, err := newDecryptingReader(username, fr); err != nil {
			responses.InternalServerError(w, err)
		} else {
			responses.DownloadFromReader(w, path.Base(m.Key), dr)
		}
	}
}

func newDecryptingReader(kid string, r io.ReadCloser) (io.Reader, error) {
	k, err := getOrCreateKey(kid)
	iv := sha256.Sum256([]byte(kid))

	switch config.FileEncryptionMode() {
	case crypto.ModeGCM:
		return crypto.GCMDecrypter(k, iv[:crypto.GCMNonceSize], r), err
	default:
		return crypto.CTRDecrypter(k, iv[:aes.BlockSize], r), err
	}
}
