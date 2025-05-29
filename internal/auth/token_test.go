package auth

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"math/big"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"github.com/edutko/crypto-fails/internal/user/role"
)

func TestNewToken(t *testing.T) {
	k := []byte("JWTSecretKeyDontUseInProduction!")
	defaultJWTSigningKey = func() []byte { return k }

	testCases := []struct {
		name     string
		username string
		realName string
		roles    []string
	}{
		{"minimal", "alice", "", nil},
		{"admin", "admin", "", []string{role.Admin}},
		{"maximal", "batman", "Bruce Wayne", []string{role.Admin}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tok, err := NewToken(tc.username, tc.realName, tc.roles)

			parts := strings.Split(tok, ".")

			assert.NoError(t, err)
			assert.NotEmpty(t, tok)
			assert.Len(t, parts, 3)
			assert.JSONEq(t, `{"alg": "HS256", "typ": "JWT"}`, b64Decode(t, parts[0]))
			assert.Contains(t, b64Decode(t, parts[1]), `"sub":"`+tc.username+`"`)
			assert.Contains(t, b64Decode(t, parts[1]), `"iat":`)
			assert.Contains(t, b64Decode(t, parts[1]), `"exp":`)
			assert.Contains(t, b64Decode(t, parts[1]), `"nbf":`)
			assert.NotEmpty(t, parts[2])
		})
	}
}

func TestParseToken(t *testing.T) {
	k := []byte("JWTSecretKeyDontUseInProduction!")
	defaultJWTSigningKey = func() []byte { return k }
	k1 := bytes.Repeat([]byte{0x01}, 32)
	k2 := bytes.Repeat([]byte{0x02}, 32)
	kEC1, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	ecPubBytes, _ := x509.MarshalPKIXPublicKey(kEC1.Public())
	kRSA1, _ := rsa.GenerateKey(rand.Reader, 1024)
	revokedToken := testToken(t, "foo", "bar", []string{"something", "another thing"})
	jwtSigningKey = func(keyId string) ([]byte, bool) {
		switch keyId {
		case "1":
			return k1, true
		case "2":
			return pem.EncodeToMemory(&pem.Block{
				Type:  "PUBLIC KEY",
				Bytes: ecPubBytes,
			}), true
		default:
			return nil, false
		}
	}

	testCases := []struct {
		name            string
		token           string
		expectedSession *Session
		expectedErr     error
	}{
		{"minimal", testToken(t, "alice", "", nil), &Session{Username: "alice"}, nil},
		{"maximal", testToken(t, "bob", "Bob Robertson", []string{"admin"}),
			&Session{Username: "bob", RealName: "Bob Robertson", IsAdmin: true}, nil},
		{"HS256", testTokenWithHeader(t, "eve", map[string]any{"alg": "HS256", "kid": "1"}, k1),
			&Session{Username: "eve"}, nil},
		{"ES256", testTokenWithHeader(t, "eve", map[string]any{"alg": "ES256", "kid": "2"}, kEC1),
			&Session{Username: "eve"}, nil},

		{"revoked", expiredToken(t, "eric"), nil, nil},
		{"revoked", revokedToken, nil, nil},
		{"bad signature", testTokenWithHeader(t, "eve", nil, k1), nil, ErrInvalidToken},

		{"alg:none", testTokenWithHeader(t, "eve", map[string]any{"alg": "none"}, nil), &Session{Username: "eve"}, nil},
		{"jwk:rsa",
			testTokenWithHeader(t, "eve", map[string]any{"alg": "RS256", "kid": "2", "jwk": map[string]any{
				"kid": "2",
				"e":   b64Encode(big.NewInt(int64(kRSA1.E)).Bytes()),
				"n":   b64Encode(kRSA1.N.Bytes()),
				"kty": "RSA",
			}}, kRSA1),
			&Session{Username: "eve"}, nil},
		{"jwk:hmac",
			testTokenWithHeader(t, "eve", map[string]any{"alg": "HS256", "kid": "1", "jwk": map[string]any{
				"kid": "1",
				"k":   b64Encode(k2),
				"kty": "oct",
			}}, k2),
			&Session{Username: "eve"}, nil},
		{"jku", testTokenWithHeader(t, "eve", map[string]any{"alg": "RS256", "kid": "3", "jku": "http://localhost:9999/jwks"}, kRSA1),
			&Session{Username: "eve"}, nil},
	}

	RevokeSession(revokedToken)

	ctx, stopServer := context.WithCancel(context.Background())
	startJWKSServer(ctx, `{"keys":[{"kid":"3", "e":"`+
		base64.RawURLEncoding.EncodeToString(big.NewInt(int64(kRSA1.E)).Bytes())+
		`", "n":"`+base64.RawURLEncoding.EncodeToString(kRSA1.N.Bytes())+
		`", "kty":"RSA"}]}`)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := ParseToken(tc.token)

			assert.Equal(t, tc.expectedSession, s)
			assert.Equal(t, tc.expectedErr, err)
		})
	}

	stopServer()
}

func b64Encode(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func b64Decode(t *testing.T, s string) string {
	b, err := base64.RawURLEncoding.DecodeString(s)
	assert.NoError(t, err)
	return string(b)
}

func testToken(t *testing.T, username, realName string, roles []string) string {
	tok, err := NewToken(username, realName, roles)
	assert.NoError(t, err)
	return tok
}

func expiredToken(t *testing.T, username string) string {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now.Add(-6 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(now.Add(-5 * time.Minute)),
		},
	})

	s, err := token.SignedString(defaultJWTSigningKey())
	assert.NoError(t, err)
	return s
}

func testTokenWithHeader(t *testing.T, username string, hdr map[string]any, key any) string {
	now := time.Now()
	var alg jwt.SigningMethod
	alg = jwt.SigningMethodHS256
	switch hdr["alg"] {
	case "none":
		alg = jwt.SigningMethodNone
	case "ES256":
		alg = jwt.SigningMethodES256
	case "RS256":
		alg = jwt.SigningMethodRS256
	}

	token := jwt.NewWithClaims(alg, customClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(6 * time.Hour)),
			NotBefore: jwt.NewNumericDate(now.Add(-5 * time.Minute)),
		},
	})

	for k, v := range hdr {
		token.Header[k] = v
	}

	if alg == jwt.SigningMethodNone {
		s, err := token.SigningString()
		assert.NoError(t, err)
		return s + "."
	} else {
		s, err := token.SignedString(key)
		assert.NoError(t, err)
		return s
	}
}

func startJWKSServer(ctx context.Context, jwkSet string) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /jwks", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(jwkSet))
	})
	s := &http.Server{
		Addr:    "localhost:9999",
		Handler: mux,
	}

	go func() {
		err := s.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	time.Sleep(time.Millisecond * 500)

	go func() {
		<-ctx.Done()
		_ = s.Shutdown(context.Background())
	}()
}
