package route

import (
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/edutko/crypto-fails/internal/crypto"
	"github.com/edutko/crypto-fails/internal/route/requests"
	"github.com/edutko/crypto-fails/internal/route/responses"
	"github.com/edutko/crypto-fails/internal/stores"
	"github.com/edutko/crypto-fails/internal/view"
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
		createUser(u, w, r)
	}
}

func GetUserPubkeys(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	getPubkeysForUser(username, w)
}

func createUser(u user.User, w http.ResponseWriter, r *http.Request) {
	if u.Username == "" || u.Password == "" {
		msg := "username and password are required"
		if requests.IsInteractive(r) {
			responses.RenderView(w, r.Context(), view.RegistrationForm(user.User{}, msg))
		} else {
			responses.BadRequestWithMessage(w, msg)
		}
		return
	}

	if !user.UsernamePattern.MatchString(u.Username) {
		msg := "username must match /" + user.UsernamePattern.String() + "/"
		if requests.IsInteractive(r) {
			responses.RenderView(w, r.Context(), view.RegistrationForm(user.User{}, msg))
		} else {
			responses.BadRequestWithMessage(w, msg)
		}
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

	created, err := stores.UserStore().PutIfNotExists(u.Username, u)
	if err != nil {
		responses.InternalServerError(w, err)
		return
	}

	if !created {
		msg := fmt.Sprintf("user %q already exists", u.Username)
		if requests.IsInteractive(r) {
			u.Password = password
			u.PasswordHash = ""
			responses.RenderView(w, r.Context(), view.RegistrationForm(u, msg))
		} else {
			responses.ConflictWithMessage(w, msg)
		}
		return
	}

	if requests.IsInteractive(r) {
		login(u.Username, password, w, r)
	} else {
		responses.Created(w, path.Join("/api", "users", url.PathEscape(u.Username)))
	}
}
