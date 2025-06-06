package route

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/edutko/crypto-fails/internal/crypto"
	"github.com/edutko/crypto-fails/internal/responses"
	"github.com/edutko/crypto-fails/internal/stores"
	"github.com/edutko/crypto-fails/internal/user"
)

func Users(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		getUsers(w)

	} else if r.Method == http.MethodPost {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			responses.InternalServerError(w, err)
		}

		var u user.User
		if err := json.Unmarshal(b, &u); err != nil {
			responses.BadRequest(w, err)
		}

		createUser(u, w)

	} else {
		responses.MethodNotAllowed(w)
	}
}

func UserPubkeys(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("id")
	if r.Method == http.MethodGet {
		getPubkeysForUser(w, username)
	} else {
		responses.MethodNotAllowed(w)
	}
}

func getUsers(w http.ResponseWriter) {
	usernames, err := stores.UserStore().ListKeys()
	if err != nil {
		responses.InternalServerError(w, err)
		return
	}

	responses.JSON(w, struct {
		Users []string `json:"users"`
	}{usernames})
}

func createUser(u user.User, w http.ResponseWriter) {
	if u.Username == "" || u.Password == "" {
		responses.BadRequestWithMessage(w, "username and password are required")
		return
	}

	if !user.UsernamePattern.MatchString(u.Username) {
		responses.BadRequestWithMessage(w, "username must match /"+user.UsernamePattern.String()+"/")
		return
	}

	ph, err := crypto.HashPassword(u.Password)
	if err != nil {
		responses.InternalServerError(w, err)
		return
	}

	u.Password = ""
	u.PasswordHash = ph

	err = stores.UserStore().Put(u.Username, u)
	if err != nil {
		responses.InternalServerError(w, err)
		return
	}

	responses.Created(w, path.Join("/api", "users", url.PathEscape(u.Username)))
}
