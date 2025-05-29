package crypto

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/openpgp"
)

func GetGPGKeyId(b []byte) (string, error) {
	el, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("openpgp.ReadArmoredKeyRing: %w", err)
	}

	if len(el) != 1 {
		return "", fmt.Errorf("expected one key")
	}

	return strings.ToUpper(hex.EncodeToString(el[0].PrimaryKey.Fingerprint[:])), nil
}
