package crypto

import "golang.org/x/crypto/bcrypt"

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
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
