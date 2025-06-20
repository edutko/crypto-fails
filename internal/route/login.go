package route

import (
	"net/http"

	"github.com/edutko/crypto-fails/internal/auth"
	"github.com/edutko/crypto-fails/internal/config"
	"github.com/edutko/crypto-fails/internal/route/requests"
	"github.com/edutko/crypto-fails/internal/route/responses"
	"github.com/edutko/crypto-fails/pkg/api"
)

func GetLoginUI(w http.ResponseWriter, r *http.Request) {
	responses.Found(w, "/")
}

func PostLoginUI(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	login(username, password, w, requests.WithInteractiveLabel(r))
}

func PostLoginAPI(w http.ResponseWriter, r *http.Request) {
	var l struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := requests.ParseJSONBody(r, &l); err != nil {
		responses.BadRequest(w, err)
		return
	}

	login(l.Username, l.Password, w, r)
}

func login(username, password string, w http.ResponseWriter, r *http.Request) {
	u, err := authenticate(username, password)
	if err != nil || u == nil {
		responses.Unauthorized(w)
		return
	}

	if requests.IsInteractive(r) {
		if c, err := auth.NewCookie(u.Username, u.RealName, config.SessionDuration(), u.Roles); err != nil {
			responses.InternalServerError(w, err)
		} else {
			http.SetCookie(w, c)
			responses.SeeOther(w, "/files")
		}

	} else {
		if token, err := auth.NewToken(u.Username, u.RealName, u.Roles); err != nil {
			responses.InternalServerError(w, err)
		} else {
			responses.JSON(w, api.TokenResponse{Token: token})
		}
	}
}

var authenticate = auth.AuthenticateWithPassword
