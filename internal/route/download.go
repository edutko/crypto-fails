package route

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/edutko/crypto-fails/internal/auth"
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
		// unsigned links are only valid for a user's own files
		if strings.HasPrefix(path.Join(s.Username, l.Key), s.Username+"/") {
			downloadFile(s.Username, l.Key, w, r)
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
		downloadFile(parts[0], path.Join(parts[1:]...), w, r)
	}
}

func downloadFile(namespace, key string, w http.ResponseWriter, _ *http.Request) {
	fr, m, err := stores.FileStore().GetObject(path.Join(namespace, key))
	if errors.Is(err, store.ErrNotFound) {
		responses.NotFound(w)

	} else if err != nil {
		responses.InternalServerError(w, err)

	} else {
		defer fr.Close()
		if dr, err := stores.FileStore().NewDecryptingReader(namespace, fr); err != nil {
			responses.InternalServerError(w, err)
		} else {
			responses.DownloadFromReader(w, path.Base(m.Key), dr)
		}
	}
}
