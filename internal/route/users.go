package route

import (
	"net/http"
	"net/url"
	"path"

	"github.com/edutko/crypto-fails/internal/crypto"
	"github.com/edutko/crypto-fails/internal/route/requests"
	"github.com/edutko/crypto-fails/internal/route/responses"
	"github.com/edutko/crypto-fails/internal/stores"
	"github.com/edutko/crypto-fails/pkg/api"
	"github.com/edutko/crypto-fails/pkg/user"
)

func GetUsers(w http.ResponseWriter, r *http.Request) {
	usernames, err := stores.UserStore().ListKeys()
	if err != nil {
		responses.InternalServerError(w, err)
		return
	}
	responses.JSON(w, api.UsersResponse{Users: usernames})
}

func PostUsers(w http.ResponseWriter, r *http.Request) {
	var u user.User
	if err := requests.ParseJSONBody(r, &u); err != nil {
		responses.BadRequest(w, err)
	} else {
		createUser(u, w, false)
	}
}

func GetUserPubkeys(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	getPubkeysForUser(w, username)
}

func createUser(u user.User, w http.ResponseWriter, interactive bool) {
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

	password := u.Password
	u.Password = ""
	u.PasswordHash = ph

	err = stores.UserStore().Put(u.Username, u)
	if err != nil {
		responses.InternalServerError(w, err)
		return
	}

	if interactive {
		interactiveLogin(w, u.Username, password)
	} else {
		responses.Created(w, path.Join("/api", "users", url.PathEscape(u.Username)))
	}
}
