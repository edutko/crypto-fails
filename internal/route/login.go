package route

import (
	"net/http"

	"github.com/edutko/crypto-fails/internal/auth"
	"github.com/edutko/crypto-fails/internal/route/requests"
	"github.com/edutko/crypto-fails/internal/route/responses"
	"github.com/edutko/crypto-fails/pkg/api"
)

func GetLoginUI(w http.ResponseWriter, r *http.Request) {
	responses.Found(w, "/")
}

func PostLoginUI(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		responses.BadRequest(w, err)
		return
	}

	username := r.Form.Get("username")
	password := r.Form.Get("password")
	if u, err := authenticate(username, password); err != nil || u == nil {
		responses.Unauthorized(w)
	} else {
		if c, err := auth.NewCookie(u.Username, u.RealName, u.Roles); err != nil {
			responses.InternalServerError(w, err)
		} else {
			http.SetCookie(w, c)
			responses.SeeOther(w, "/files")
		}
	}
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

var authenticate = auth.AuthenticateWithPassword
