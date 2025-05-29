package route

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/edutko/crypto-fails/internal/auth"
	"github.com/edutko/crypto-fails/internal/config"
	"github.com/edutko/crypto-fails/internal/crypto/random"
	"github.com/edutko/crypto-fails/internal/middleware"
	"github.com/edutko/crypto-fails/internal/responses"
	"github.com/edutko/crypto-fails/internal/share"
	"github.com/edutko/crypto-fails/internal/stores"
	"github.com/edutko/crypto-fails/internal/view"
)

func MyShares(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)
	prefix := s.Username + "/"
	if r.Method != http.MethodGet {
		responses.MethodNotAllowed(w)
		return
	}

	if links, err := listShares(s.Username); err != nil {
		responses.InternalServerError(w, err)
	} else {
		var tbl [][]string
		for _, v := range links {
			tbl = append(tbl, []string{strings.TrimPrefix(v.Key, prefix), v.Expiration.Format("2006-01-02")})
		}

		responses.RenderView(w, r.Context(), view.MyShares(tbl, []string{"left-justified", "centered"}))
	}
}

func NewShare(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			responses.BadRequest(w, err)
			return
		}

		key := r.PostFormValue("key")
		if !strings.HasPrefix(path.Join(s.Username, key), s.Username+"/") {
			responses.BadRequest(w, errors.New("invalid key"))
			return
		}

		if l, err := newSignedLink(s.Username, key); err != nil {
			responses.InternalServerError(w, err)
		} else {
			responses.Plaintext(w, l)
		}

	} else {
		responses.MethodNotAllowed(w)
	}
}

func Shares(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)
	if r.Method == http.MethodGet {
		if links, err := listShares(s.Username); err != nil {
			responses.InternalServerError(w, err)
		} else {
			responses.JSON(w, struct {
				Links []share.Link `json:"links"`
			}{links})
		}

	} else if r.Method == http.MethodPost {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			responses.InternalServerError(w, err)
			return
		}

		var link share.Link
		if err = json.Unmarshal(b, &link); err != nil {
			responses.BadRequest(w, err)
			return
		}

		l, err := newSignedLink(s.Username, link.Key)
		if err != nil {
			responses.InternalServerError(w, err)
			return
		}

		responses.JSON(w, struct {
			Link string `json:"link"`
		}{l})

	} else {
		responses.MethodNotAllowed(w)
	}
}

func Share(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)
	id := r.PathValue("id")

	if r.Method == http.MethodDelete {
		if _, err := stores.ShareStore().Delete(path.Join(s.Username, id)); err != nil {
			responses.InternalServerError(w, err)
		} else {
			responses.NoContent(w)
		}

	} else {
		responses.MethodNotAllowed(w)
	}
}

func newSignedLink(username, relativeKey string) (string, error) {
	exp := time.Now().Add(config.ShareLinkDuration())
	l := share.NewLink(path.Join(username, relativeKey), exp)

	err := stores.ShareStore().Put(path.Join(username, random.String(6)), l)

	return l.SignedQueryString(auth.GetShareLinkSecret()), err
}

func listShares(username string) ([]share.Link, error) {
	prefix := username + "/"
	ids, err := stores.ShareStore().ListKeysWithPrefix(prefix)
	if err != nil {
		return nil, err
	}

	var links []share.Link
	for i := range ids {
		l, _ := stores.ShareStore().Get(ids[i])
		links = append(links, share.Link{
			Id:         strings.TrimPrefix(ids[i], prefix),
			Key:        strings.TrimPrefix(l.Key, prefix),
			Expiration: l.Expiration,
		})
	}

	return links, nil
}
