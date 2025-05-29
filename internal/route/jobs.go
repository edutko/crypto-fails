package route

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/edutko/crypto-fails/internal/middleware"
	"github.com/edutko/crypto-fails/internal/responses"
	"github.com/edutko/crypto-fails/internal/store"
	"github.com/edutko/crypto-fails/internal/stores"
)

func Job(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responses.MethodNotAllowed(w)
		return
	}

	s := middleware.GetCurrentSession(r)
	id := path.Clean(r.PathValue("id"))
	if !strings.HasPrefix(id, s.Username+"/") {
		responses.Forbidden(w, fmt.Errorf("access denied for %q on %q", s.Username, r.PathValue("id")))
		return
	}

	if j, err := stores.JobStore().Get(id); errors.Is(err, store.ErrNotFound) {
		responses.NotFound(w)
	} else if err != nil {
		responses.InternalServerError(w, err)
	} else {
		responses.JSON(w, j)
	}
}
