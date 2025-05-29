package route

import (
	"net/http"

	"github.com/edutko/crypto-fails/internal/responses"
	"github.com/edutko/crypto-fails/internal/user"
	"github.com/edutko/crypto-fails/internal/view"
)

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		responses.RenderView(w, r.Context(), view.RegistrationForm())

	} else if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			responses.BadRequest(w, err)
			return
		}

		createUser(user.User{
			Username: r.Form.Get("username"),
			Password: r.Form.Get("password"),
			RealName: r.Form.Get("realname"),
			Email:    r.Form.Get("email"),
		}, w)

	} else {
		responses.MethodNotAllowed(w)
		return
	}
}
