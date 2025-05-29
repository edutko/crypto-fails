package route

import (
	"net/http"

	"github.com/edutko/crypto-fails/internal/responses"
	"github.com/edutko/crypto-fails/internal/view"
)

func Index(w http.ResponseWriter, r *http.Request) {
	responses.RenderView(w, r.Context(), view.Index())
}
