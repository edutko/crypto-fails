package auth

import (
	"errors"
	"path"

	"github.com/google/uuid"

	"github.com/edutko/crypto-fails/internal/crypto/keys"
	"github.com/edutko/crypto-fails/internal/crypto/random"
	"github.com/edutko/crypto-fails/internal/store"
	"github.com/edutko/crypto-fails/internal/store/constants"
	"github.com/edutko/crypto-fails/internal/stores"
)

func GetCookieEncryptionKey() []byte {
	k, err := keyStore().Get(constants.CookieEncryptionKID)
	if errors.Is(err, store.ErrNotFound) {
		k = random.Bytes(32)
		if _, err = keyStore().PutIfNotExists(constants.CookieEncryptionKID, k); err != nil {
			panic(err)
		}
		k, err = keyStore().Get(constants.CookieEncryptionKID)
	}
	if err != nil {
		panic(err)
	}
	return k
}

func GetDefaultJWTSigningKey() []byte {
	if k, ok := GetJWTSigningKey(path.Join(constants.JWTSigningKIDPrefix, "default")); ok {
		return k
	} else {
		return []byte("JWTSecretKeyDontUseInProduction!")
	}
}

func GetJWTSigningKey(keyId string) ([]byte, bool) {
	if k, err := keyStore().Get(path.Join(constants.JWTSigningKIDPrefix, keyId)); err != nil {
		return nil, false
	} else {
		return k, true
	}
}

func GetShareLinkSecret() []byte {
	k, err := keyStore().Get(constants.ShareLinkSecretId)
	if errors.Is(err, store.ErrNotFound) {
		k = random.Bytes(12)
		if _, err = keyStore().PutIfNotExists(constants.ShareLinkSecretId, k); err != nil {
			panic(err)
		}
		k, err = keyStore().Get(constants.ShareLinkSecretId)
	}
	if err != nil {
		panic(err)
	}
	return k
}

func InitializeKeys() error {
	ecKid := uuid.Must(uuid.NewRandom())
	_, _, ecPub, err := keys.GenerateECDSAKeyPair()
	if err != nil {
		return err
	}
	if err = keyStore().Put(path.Join(constants.JWTSigningKIDPrefix, ecKid.String()), ecPub); err != nil {
		return err
	}

	rsaKid := uuid.Must(uuid.NewRandom())
	_, _, rsaPub, err := keys.GenerateRSAKeyPair()
	if err != nil {
		return err
	}
	if err = keyStore().Put(path.Join(constants.JWTSigningKIDPrefix, rsaKid.String()), rsaPub); err != nil {
		return err
	}

	return nil
}

var keyStore = stores.KeyStore
