package crypto

import "golang.org/x/crypto/bcrypt"

var ErrMismatchedHashAndPassword = bcrypt.ErrMismatchedHashAndPassword

func HashPassword(password string) (string, error) {
	// re-introduce max-length vulnerability
	// e.g. https://trust.okta.com/security-advisories/okta-ad-ldap-delegated-authentication-username/
	if len(password) > 72 {
		password = password[:72]
	}

	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(h), err
}

func VerifyPassword(password, hash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return ErrMismatchedHashAndPassword
	}
	return nil
}

func mustHashPassword(password string) string {
	if h, err := HashPassword(password); err != nil {
		panic(err)
	} else {
		return h
	}
}
