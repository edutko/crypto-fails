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
	interactiveLogin(w, username, password)
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

	if u, err := authenticate(l.Username, l.Password); err != nil {
		responses.Unauthorized(w)
	} else {
		if token, err := auth.NewToken(u.Username, u.RealName, u.Roles); err != nil {
			responses.InternalServerError(w, err)
		} else {
			responses.JSON(w, api.TokenResponse{Token: token})
		}
	}
}

func interactiveLogin(w http.ResponseWriter, username, password string) {
	if u, err := authenticate(username, password); err != nil || u == nil {
		responses.Unauthorized(w)
	} else {
		if c, err := auth.NewCookie(u.Username, u.RealName, config.SessionDuration(), u.Roles); err != nil {
			responses.InternalServerError(w, err)
		} else {
			http.SetCookie(w, c)
			responses.SeeOther(w, "/files")
		}
	}
}

var authenticate = auth.AuthenticateWithPassword
