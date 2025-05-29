package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/edutko/crypto-fails/internal/config"
	"github.com/edutko/crypto-fails/internal/crypto/pkcs7"
	"github.com/edutko/crypto-fails/internal/user/role"
)

const CookieName = "auth"

var (
	ErrInvalidCookie = errors.New("invalid cookie")
)

func NewCookie(username, realName string, roles []string) (*http.Cookie, error) {
	s := &Session{
		Username: username,
		IsAdmin:  slices.Contains(roles, role.Admin),
		RealName: realName,
		Expires:  int(timeNow().Add(config.SessionDuration()).Unix()),
	}
	plaintext := pkcs7.Pad([]byte(s.QueryString()), aes.BlockSize)
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}
	ciphertext := make([]byte, len(plaintext))

	b, _ := aes.NewCipher(cookieEncryptionKey())
	cipher.NewCBCEncrypter(b, iv).CryptBlocks(ciphertext, plaintext)

	return &http.Cookie{
		Name:     CookieName,
		Value:    hex.EncodeToString(append(iv, ciphertext...)),
		MaxAge:   int(config.SessionDuration()),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}, nil
}

func ParseCookie(c *http.Cookie) (*Session, error) {
	if IsSessionRevoked(c.Value) {
		return nil, nil
	}

	ciphertext, err := hex.DecodeString(c.Value)
	if err != nil {
		return nil, ErrInvalidCookie
	}

	if len(ciphertext) < aes.BlockSize*2 || len(ciphertext)%aes.BlockSize != 0 {
		return nil, ErrInvalidCookie
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	plaintext := make([]byte, len(ciphertext))

	b, _ := aes.NewCipher(cookieEncryptionKey())
	cipher.NewCBCDecrypter(b, iv).CryptBlocks(plaintext, ciphertext)

	plaintext, err = pkcs7.Unpad(plaintext, aes.BlockSize)
	if err != nil {
		return nil, ErrInvalidCookie
	}

	return ParseSession(string(plaintext)), nil
}

var timeNow = time.Now
var cookieEncryptionKey = GetCookieEncryptionKey
