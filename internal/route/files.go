package route

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"slices"
	"strings"

	"github.com/edutko/crypto-fails/internal/middleware"
	"github.com/edutko/crypto-fails/internal/route/responses"
	"github.com/edutko/crypto-fails/internal/store"
	"github.com/edutko/crypto-fails/internal/stores"
	"github.com/edutko/crypto-fails/internal/view"
	"github.com/edutko/crypto-fails/pkg/api"
	"github.com/edutko/crypto-fails/pkg/blob"
)

func GetMyFiles(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)
	if files, err := listFiles(s.Username); err != nil {
		responses.InternalServerError(w, err)

	} else {
		responses.RenderView(w, r.Context(), view.MyFiles(files))
	}
}

func GetFiles(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)

	if blobs, err := listFiles(s.Username); err != nil {
		responses.InternalServerError(w, err)
	} else {
		responses.JSON(w, api.FilesMetadataResponse{Files: blobs})
	}
}

func PostFiles(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)
	uploadFile(s.Username, w, r)
}

func GetFile(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)
	key := path.Join(s.Username, r.PathValue("key"))

	if !strings.HasPrefix(key, s.Username+"/") {
		responses.Forbidden(w, fmt.Errorf("permission denied for %q on %q", s.Username, r.PathValue("key")))
		return
	}

	downloadFile(s.Username, key, w)
}

func DeleteFile(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)
	key := path.Join(s.Username, r.PathValue("key"))

	if !strings.HasPrefix(key, s.Username+"/") {
		responses.Forbidden(w, fmt.Errorf("permission denied for %q on %q", s.Username, r.PathValue("key")))
		return
	}

	_, err := stores.FileStore().DeleteObject(key)
	if err == nil || errors.Is(err, store.ErrNotFound) {
		responses.NoContent(w)
	} else {
		responses.InternalServerError(w, err)
	}
}

func listFiles(namespace string) ([]blob.Metadata, error) {
	blobs, err := stores.FileStore().ListObjectsWithPrefix(namespace)
	if err != nil {
		return nil, err
	}

	for i := range blobs {
		blobs[i].Key = strings.TrimPrefix(blobs[i].Key, namespace+"/")
	}
	slices.SortFunc(blobs, func(a, b blob.Metadata) int { return strings.Compare(a.Key, b.Key) })

	return blobs, nil
}
