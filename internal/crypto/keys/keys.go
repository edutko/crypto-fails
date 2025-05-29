package keys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

func GenerateECDSAKeyPair() (*ecdsa.PrivateKey, []byte, []byte, error) {
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, err
	}

	privDER, err := x509.MarshalPKCS8PrivateKey(k)
	if err != nil {
		return nil, nil, nil, err
	}
	priv := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER})

	pubDER, err := x509.MarshalPKIXPublicKey(&k.PublicKey)
	if err != nil {
		return nil, nil, nil, err
	}
	pub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})

	return k, priv, pub, nil
}

func GenerateRSAKeyPair() (*rsa.PrivateKey, []byte, []byte, error) {
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, nil, err
	}

	privDER, err := x509.MarshalPKCS8PrivateKey(k)
	if err != nil {
		return nil, nil, nil, err
	}
	priv := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER})

	pubDER, err := x509.MarshalPKIXPublicKey(&k.PublicKey)
	if err != nil {
		return nil, nil, nil, err
	}
	pub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})

	return k, priv, pub, nil
}

func ParsePublicKeyPEM(p []byte) (any, error) {
	b, _ := pem.Decode(p)
	if b.Type == "PUBLIC KEY" {
		return x509.ParsePKIXPublicKey(b.Bytes)
	}
	return nil, fmt.Errorf("unrecognized PEM block: %q", b.Type)
}
