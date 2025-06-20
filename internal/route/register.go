package route

import (
	"net/http"

	"github.com/edutko/crypto-fails/internal/route/requests"
	"github.com/edutko/crypto-fails/internal/route/responses"
	"github.com/edutko/crypto-fails/internal/view"
	"github.com/edutko/crypto-fails/pkg/user"
)

func GetRegister(w http.ResponseWriter, r *http.Request) {
	responses.RenderView(w, r.Context(), view.RegistrationForm(user.User{}, ""))
}

func PostRegister(w http.ResponseWriter, r *http.Request) {
	createUser(user.User{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
		RealName: r.FormValue("realname"),
		Email:    r.FormValue("email"),
	}, w, requests.WithInteractiveLabel(r))
}
