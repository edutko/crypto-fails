package route

import (
	"net/http"

	"github.com/edutko/crypto-fails/internal/route/responses"
	"github.com/edutko/crypto-fails/internal/view"
	"github.com/edutko/crypto-fails/pkg/user"
)

func GetRegister(w http.ResponseWriter, r *http.Request) {
	responses.RenderView(w, r.Context(), view.RegistrationForm())
}

func PostRegister(w http.ResponseWriter, r *http.Request) {
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
}
