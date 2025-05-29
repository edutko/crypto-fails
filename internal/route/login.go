package route

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/edutko/crypto-fails/internal/auth"
	"github.com/edutko/crypto-fails/internal/responses"
)

func LoginUI(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		responses.Found(w, "/")

	} else if r.Method == http.MethodPost {
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

	} else {
		responses.MethodNotAllowed(w)
	}
}

func LoginAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responses.MethodNotAllowed(w)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.InternalServerError(w, err)
		return
	}

	var l struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err = json.Unmarshal(body, &l); err != nil {
		responses.BadRequest(w, err)
		return
	}

	if u, err := authenticate(l.Username, l.Password); err != nil {
		responses.Unauthorized(w)
	} else {
		if token, err := auth.NewToken(u.Username, u.RealName, u.Roles); err != nil {
			responses.InternalServerError(w, err)
		} else {
			responses.JSON(w, struct {
				Token string `json:"token"`
			}{Token: token})
		}
	}
}

var authenticate = auth.AuthenticateWithPassword
