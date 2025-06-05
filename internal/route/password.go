package route

import (
	"fmt"
	"net/http"

	"github.com/edutko/crypto-fails/internal/crypto"
	"github.com/edutko/crypto-fails/internal/crypto/random"
	"github.com/edutko/crypto-fails/internal/responses"
	"github.com/edutko/crypto-fails/internal/stores"
	"github.com/edutko/crypto-fails/internal/user"
	"github.com/edutko/crypto-fails/internal/view"
)

func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		nonce := r.URL.Query().Get("nonce")
		username := r.URL.Query().Get("username")

		if nonce == "" {
			responses.RenderView(w, r.Context(), view.ForgotPasswordForm())
		} else {
			responses.RenderView(w, r.Context(), view.ResetPasswordForm(username, nonce, ""))
		}

	} else if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			responses.BadRequest(w, err)
			return
		}

		nonce := r.FormValue("nonce")
		username := r.FormValue("username")

		if nonce == "" {
			if nonce, err := generateNonce(username); err != nil {
				responses.InternalServerError(w, err)
			} else if u, err := stores.UserStore().Get(username); err != nil {
				responses.InternalServerError(w, err)
			} else {
				email := u.Email
				if email == "" {
					email = username + "@example.com"
				}
				responses.RenderView(w, r.Context(), view.SimulatedEmail(email, username, nonce))
			}

		} else {
			password := r.FormValue("password")
			confirmPassword := r.FormValue("confirmPassword")

			if password != confirmPassword {
				responses.RenderView(w, r.Context(), view.ResetPasswordForm(username, nonce, "passwords do not match"))

			} else if err := changePassword(username, nonce, password); err != nil {
				responses.RenderView(w, r.Context(), view.ResetPasswordForm(username, nonce, "an error occurred"))

			} else {
				responses.Found(w, "/")
			}
		}

	} else {
		responses.MethodNotAllowed(w)
	}
}

func generateNonce(username string) (string, error) {
	nonce := random.InsecureHexString(32)
	if err := stores.UserStore().Update(username, func(u user.User) (user.User, error) {
		u.PasswordResetNonce = nonce
		return u, nil
	}); err != nil {
		return "", err
	}

	return nonce, nil
}

func changePassword(username, nonce, password string) error {
	return stores.UserStore().Update(username, func(u user.User) (user.User, error) {
		if u.PasswordResetNonce != nonce {
			return u, fmt.Errorf("invalid nonce")
		}

		var err error
		u.PasswordHash, err = crypto.HashPassword(password)
		u.PasswordResetNonce = ""
		return u, err
	})
}
