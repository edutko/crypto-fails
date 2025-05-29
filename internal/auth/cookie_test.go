package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/edutko/crypto-fails/internal/config"
	"github.com/edutko/crypto-fails/internal/crypto/pkcs7"
	"github.com/edutko/crypto-fails/internal/crypto/random"
)

func TestCreateCookie(t *testing.T) {
	config.InitializeForTesting(config.Config{
		SessionDuration: 1 * time.Hour,
	})
	now := time.Now()
	exp := now.Add(config.SessionDuration()).Unix()
	timeNow = func() time.Time { return now }
	k := random.Bytes(32)
	cookieEncryptionKey = func() []byte { return k }

	testCases := []struct {
		name     string
		username string
		realName string
		roles    []string
		expected string
	}{
		{"minimal", "alice", "", nil, fmt.Sprintf("adm=false&exp=%d&uid=alice", exp)},
		{"maximal", "admin", "root", []string{"admin"}, fmt.Sprintf("adm=true&exp=%d&name=root&uid=admin", exp)},
		{"empty", "", "", nil, fmt.Sprintf("adm=false&exp=%d&uid=", exp)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewCookie(tc.username, tc.realName, tc.roles)

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, decrypt(c.Value))
			assert.True(t, c.HttpOnly)
			assert.True(t, c.Secure)
		})
	}
}

func TestParseCookie(t *testing.T) {
	k := random.Bytes(32)
	cookieEncryptionKey = func() []byte { return k }

	revoked := encrypt("adm=false&uid=eve")
	RevokeSession(revoked)

	testCases := []struct {
		name            string
		cookieValue     string
		expectedSession *Session
		expectedErr     error
	}{
		{"minimal", encrypt("adm=false&uid=alice"), &Session{Username: "alice"}, nil},
		{"maximal", encrypt("adm=true&name=root&uid=admin"), &Session{Username: "admin", IsAdmin: true, RealName: "root"}, nil},
		{"empty", encrypt(""), &Session{}, nil},
		{"revoked", revoked, nil, nil},

		{"invalid hex", "oopsie!", nil, ErrInvalidCookie},
		{"invalid length", "0123456789abcd", nil, ErrInvalidCookie},
		{"invalid padding", corrupted("some plaintext"), nil, ErrInvalidCookie},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := ParseCookie(&http.Cookie{
				Name:  CookieName,
				Value: tc.cookieValue,
			})

			if tc.expectedErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedSession, actual)
			} else {
				assert.Equal(t, ErrInvalidCookie, err)
				assert.Nil(t, actual)
			}
		})
	}
}

func encrypt(ptString string) string {
	plaintext := pkcs7.Pad([]byte(ptString), aes.BlockSize)
	iv := make([]byte, aes.BlockSize)
	ciphertext := make([]byte, len(plaintext))

	b, _ := aes.NewCipher(cookieEncryptionKey())
	cipher.NewCBCEncrypter(b, iv).CryptBlocks(ciphertext, plaintext)

	return hex.EncodeToString(append(iv, ciphertext...))
}

func decrypt(ctHex string) string {
	ciphertext, _ := hex.DecodeString(ctHex)

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	plaintext := make([]byte, len(ciphertext))

	b, _ := aes.NewCipher(cookieEncryptionKey())
	cipher.NewCBCDecrypter(b, iv).CryptBlocks(plaintext, ciphertext)

	plaintext, _ = pkcs7.Unpad(plaintext, aes.BlockSize)
	return string(plaintext)
}

func corrupted(plaintext string) string {
	padded := pkcs7.Pad([]byte(plaintext), aes.BlockSize)
	ciphertext := []byte(encrypt(plaintext))
	ciphertext[len(ciphertext)-1] = ciphertext[len(ciphertext)-1] ^ padded[len(padded)-1]
	return string(ciphertext)
}
