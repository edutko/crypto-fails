package route

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"sort"
	"strings"

	"github.com/edutko/crypto-fails/internal/middleware"
	"github.com/edutko/crypto-fails/internal/responses"
	"github.com/edutko/crypto-fails/internal/store"
	"github.com/edutko/crypto-fails/internal/store/blob"
	"github.com/edutko/crypto-fails/internal/stores"
	"github.com/edutko/crypto-fails/internal/view"
)

func MyFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responses.MethodNotAllowed(w)
		return
	}

	s := middleware.GetCurrentSession(r)
	if files, err := listFiles(s.Username); err != nil {
		responses.InternalServerError(w, err)

	} else {
		var keys []string
		m := make(map[string]blob.Metadata)
		for _, f := range files {
			keys = append(keys, f.Key)
			m[f.Key] = f
		}
		sort.Strings(keys)

		var tbl [][]string
		for _, k := range keys {
			itm := m[k]
			tbl = append(tbl, []string{itm.Key, fmt.Sprintf("%d", itm.Size), itm.Modified.Format("2006-01-02")})
		}

		responses.RenderView(w, r.Context(), view.MyFiles(tbl, []string{"left-justified", "right-justified", "centered"}))
	}
}

func Files(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)

	if r.Method == http.MethodGet {
		if blobs, err := listFiles(s.Username); err != nil {
			responses.InternalServerError(w, err)
		} else {
			responses.JSON(w, struct {
				Files []blob.Metadata `json:"files"`
			}{blobs})
		}

	} else if r.Method == http.MethodPost {
		uploadFile(s.Username, w, r, false)

	} else {
		responses.MethodNotAllowed(w)
	}
}

func File(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)
	key := path.Join(s.Username, r.PathValue("key"))

	if !strings.HasPrefix(key, s.Username+"/") {
		responses.Forbidden(w, fmt.Errorf("permission denied for %q on %q", s.Username, r.PathValue("key")))
		return
	}

	if r.Method == http.MethodGet {
		downloadFile(s.Username, key, w, r)

	} else if r.Method == http.MethodDelete {
		_, err := stores.FileStore().DeleteObject(key)
		if err == nil || errors.Is(err, store.ErrNotFound) {
			responses.NoContent(w)
		} else {
			responses.InternalServerError(w, err)
		}

	} else {
		responses.MethodNotAllowed(w)
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

	return blobs, nil
}
