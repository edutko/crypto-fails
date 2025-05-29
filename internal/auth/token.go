package auth

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"slices"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/edutko/crypto-fails/internal/config"
	"github.com/edutko/crypto-fails/internal/user/role"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

func NewToken(username, realName string, roles []string) (string, error) {
	now := time.Now().Truncate(time.Second)
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims{
		IsAdmin:  slices.Contains(roles, role.Admin),
		RealName: realName,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        id.String(),
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(config.SessionDuration())),
			NotBefore: jwt.NewNumericDate(now.Add(-5 * time.Minute)),
		},
	})

	return token.SignedString(defaultJWTSigningKey())
}

func ParseToken(tokenString string) (*Session, error) {
	if IsSessionRevoked(tokenString) {
		return nil, nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &customClaims{}, getKey,
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(time.Second*30),
	)
	if errors.Is(err, jwt.ErrTokenInvalidClaims) {
		return nil, nil
	}
	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*customClaims); ok {
		return &Session{
			Username: claims.Subject,
			IsAdmin:  claims.IsAdmin,
			RealName: claims.RealName,
		}, nil
	}

	return nil, ErrInvalidToken
}

func getKey(token *jwt.Token) (interface{}, error) {
	// alg: none vulnerability
	if token.Header["alg"] == "none" {
		return jwt.UnsafeAllowNoneSignatureType, nil
	}

	var keyId string
	if kid, ok := token.Header["kid"]; ok {
		keyId, ok = kid.(string)
	}

	// several vulns here:
	//  * trusting jwk and jku
	//  * preferring key from jku when there's a matching local key
	//  * allowing mismatched alg/key combinations
	if jku, ok := token.Header["jku"]; ok {
		if jwksURL, ok := jku.(string); ok {
			jwks, err := jwkset.NewDefaultHTTPClient([]string{jwksURL})
			if err == nil {
				jwKey, err := jwks.KeyRead(context.Background(), keyId)
				if err == nil {
					return jwKey.Key(), nil
				}
			}
		}
	}

	if jwk, ok := token.Header["jwk"]; ok {
		if b, err := json.Marshal(jwk); err == nil {
			if jwKey, err := jwkset.NewJWKFromRawJSON(b, jwkset.JWKMarshalOptions{Private: true}, jwkset.JWKValidateOptions{}); err == nil {
				return jwKey.Key(), nil
			}
		}
	}

	if keyBytes, ok := jwtSigningKey(keyId); ok {
		switch token.Method.Alg() {
		case "HS256", "HS384", "HS512":
			return keyBytes, nil
		default:
			if bytes.HasPrefix(keyBytes, []byte("-----BEGIN")) {
				blk, _ := pem.Decode(keyBytes)
				keyBytes = blk.Bytes
			}
			pub, err := x509.ParsePKIXPublicKey(keyBytes)
			if err != nil {
				return nil, err
			}
			return pub, nil
		}
	}

	return defaultJWTSigningKey(), nil
}

type customClaims struct {
	IsAdmin  bool   `json:"admin"`
	RealName string `json:"realName"`
	jwt.RegisteredClaims
}

var jwtSigningKey = GetJWTSigningKey
var defaultJWTSigningKey = GetDefaultJWTSigningKey
