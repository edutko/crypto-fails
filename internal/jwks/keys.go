package jwks

import (
	"context"

	"github.com/MicahParks/jwkset"
	"github.com/google/uuid"

	"github.com/edutko/crypto-fails/internal/crypto"
	"github.com/edutko/crypto-fails/internal/crypto/keys"
)

func InitializeKeys() (jwkset.Storage, map[string]crypto.KeyPair) {
	jwks := jwkset.NewMemoryStorage()
	keyPEMs := make(map[string]crypto.KeyPair)

	ecKid, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	ecKey, ecPriv, ecPub, err := keys.GenerateECDSAKeyPair()
	if err != nil {
		panic(err)
	}
	keyPEMs[ecKid.String()] = crypto.KeyPair{PrivatePEM: ecPriv, PublicPEM: ecPub}

	jwk, err := jwkset.NewJWKFromKey(ecKey, jwkset.JWKOptions{
		Metadata: jwkset.JWKMetadataOptions{KID: ecKid.String()},
	})
	if err != nil {
		panic(err)
	}
	err = jwks.KeyWrite(context.Background(), jwk)
	if err != nil {
		panic(err)
	}

	rsaKid, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	rsaKey, rsaPriv, rsaPub, err := keys.GenerateRSAKeyPair()
	if err != nil {
		panic(err)
	}
	keyPEMs[rsaKid.String()] = crypto.KeyPair{PrivatePEM: rsaPriv, PublicPEM: rsaPub}

	jwk, err = jwkset.NewJWKFromKey(rsaKey, jwkset.JWKOptions{
		Metadata: jwkset.JWKMetadataOptions{KID: rsaKid.String()},
	})
	if err != nil {
		panic(err)
	}
	err = jwks.KeyWrite(context.Background(), jwk)
	if err != nil {
		panic(err)
	}

	return jwks, keyPEMs
}
