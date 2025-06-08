package route

import (
	"net/http"

	"github.com/edutko/crypto-fails/internal/route/responses"
	"github.com/edutko/crypto-fails/internal/view"
)

func GetIndex(w http.ResponseWriter, r *http.Request) {
	responses.RenderView(w, r.Context(), view.Index())
}
