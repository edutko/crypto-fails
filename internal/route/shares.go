package route

import (
	"errors"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/edutko/crypto-fails/internal/auth"
	"github.com/edutko/crypto-fails/internal/config"
	"github.com/edutko/crypto-fails/internal/crypto/random"
	"github.com/edutko/crypto-fails/internal/middleware"
	"github.com/edutko/crypto-fails/internal/route/requests"
	"github.com/edutko/crypto-fails/internal/route/responses"
	"github.com/edutko/crypto-fails/internal/stores"
	"github.com/edutko/crypto-fails/internal/view"
	"github.com/edutko/crypto-fails/pkg/api"
	"github.com/edutko/crypto-fails/pkg/share"
)

func GetMyShares(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)
	if links, err := listShares(s.Username); err != nil {
		responses.InternalServerError(w, err)

	} else {
		var tbl [][]string
		for _, l := range links {
			tbl = append(tbl, []string{
				strings.TrimPrefix(l.Key, s.Username+"/"),
				l.Expiration.Format("2006-01-02"),
				l.Id,
				l.URL,
			})
		}
		responses.RenderView(w, r.Context(), view.MyShares(tbl, []string{"left-justified", "centered"}))
	}
}

func PostShare(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)
	key := r.PostFormValue("key")

	// normalize path and check prefix to prevent path traversals
	if !strings.HasPrefix(path.Join(s.Username, key), s.Username+"/") {
		responses.BadRequest(w, errors.New("invalid key"))
		return
	}

	if _, err := newSignedLink(s.Username, key); err != nil {
		responses.InternalServerError(w, err)
	} else {
		responses.Found(w, "/shares")
	}
}

func GetShares(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)
	if links, err := listShares(s.Username); err != nil {
		responses.InternalServerError(w, err)
	} else {
		responses.JSON(w, api.SharesResponse{Links: links})
	}
}

func PostShares(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)

	var link share.Link
	if err := requests.ParseJSONBody(r, &link); err != nil {
		responses.BadRequest(w, err)
		return
	}

	// Don't trust the username or expiration in the request.
	l, err := newSignedLink(s.Username, link.Key)
	if err != nil {
		responses.InternalServerError(w, err)
		return
	}

	responses.JSON(w, l)
}

func DeleteShare(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)
	id := r.PathValue("id")

	if _, err := stores.ShareStore().Delete(path.Join(s.Username, id)); err != nil {
		responses.InternalServerError(w, err)
	} else {
		responses.NoContent(w)
	}
}

func newSignedLink(username, relativeKey string) (share.Link, error) {
	exp := time.Now().Add(config.ShareLinkDuration())
	l := share.NewSignedLink(path.Join(username, relativeKey), exp, auth.GetShareLinkSecret())

	err := stores.ShareStore().Put(path.Join(username, random.String(6)), l)

	l.URL = buildURL(l).String()
	return l, err
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
			Id:         ids[i],
			Key:        strings.TrimPrefix(l.Key, prefix),
			Expiration: l.Expiration,
			URL:        buildURL(l).String(),
		})
	}

	return links, nil
}

func buildURL(l share.Link) *url.URL {
	u := config.BaseURL().JoinPath("/download")
	u.RawQuery = l.QueryString()
	return u
}
