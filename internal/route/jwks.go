package route

import (
	"context"
	"net/http"
	"path"

	"github.com/MicahParks/jwkset"

	"github.com/edutko/crypto-fails/internal/crypto/keys"
	"github.com/edutko/crypto-fails/internal/route/responses"
	"github.com/edutko/crypto-fails/internal/store/constants"
	"github.com/edutko/crypto-fails/internal/stores"
)

func JWKS(w http.ResponseWriter, r *http.Request) {
	jwks, err := buildJWKS(r.Context())
	if err != nil {
		responses.InternalServerError(w, err)
		return
	}

	if body, err := jwks.JSONPublic(r.Context()); err != nil {
		responses.InternalServerError(w, err)
	} else {
		responses.JSONBytes(w, body)
	}
}

func buildJWKS(ctx context.Context) (jwkset.Storage, error) {
	kids, err := stores.KeyStore().ListKeysWithPrefix(constants.JWTSigningKIDPrefix)
	if err != nil {
		return nil, err
	}

	jwks := jwkset.NewMemoryStorage()
	for _, kid := range kids {
		keyPEM, err := stores.KeyStore().Get(kid)
		if err != nil {
			return nil, err
		}
		k, err := keys.ParsePublicKeyPEM(keyPEM)
		if err != nil {
			return nil, err
		}
		jwk, err := jwkset.NewJWKFromKey(k, jwkset.JWKOptions{
			Metadata: jwkset.JWKMetadataOptions{KID: path.Base(kid)},
		})
		if err := jwks.KeyWrite(ctx, jwk); err != nil {
			return nil, err
		}
	}

	return jwks, nil
}
