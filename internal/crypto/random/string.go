package random

import (
	"crypto/rand"
	"encoding/base64"
)

func Bytes(length int) []byte {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

func String(length int) string {
	b := Bytes(length)
	return base64.RawURLEncoding.EncodeToString(b)[:length]
}
