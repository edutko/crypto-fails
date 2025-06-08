package route

import (
	"net/http"

	"github.com/edutko/crypto-fails/internal/route/responses"
	"github.com/edutko/crypto-fails/internal/view"
)

func GetAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responses.MethodNotAllowed(w)
		return
	}

	responses.RenderView(w, r.Context(), view.Admin())
}
