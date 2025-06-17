package app

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
)

var (
	ErrBadSignature   = errors.New("signature is not valid")
	ErrLicenseExpired = errors.New("license is expired or not yet valid")
)

type License struct {
	Id        string          `json:"id"`
	Licensee  string          `json:"licensee"`
	ValidFrom time.Time       `json:"validFrom"`
	ValidTo   time.Time       `json:"validTo"`
	Features  map[Feature]int `json:"features"`
	Signature string          `json:"signature,omitempty"`
}

func NewLicense(licensee string, start, end time.Time, features map[Feature]int) License {
	validFeatures := make(map[Feature]int)
	for _, k := range allFeatures {
		if v, ok := features[k]; ok {
			validFeatures[k] = v
		}
	}

	return License{
		Id:        uuid.Must(uuid.NewRandom()).String(),
		Licensee:  licensee,
		ValidFrom: start.UTC().Truncate(24 * time.Hour),
		ValidTo:   end.UTC().Truncate(24 * time.Hour),
		Features:  validFeatures,
	}
}

func ParseLicense(b []byte, verificationKey *ecdsa.PublicKey) (License, error) {
	var l License
	if err := json.Unmarshal(b, &l); err != nil {
		return License{}, err
	}

	validFeatures := make(map[Feature]int)
	for _, k := range allFeatures {
		if v, ok := l.Features[k]; ok {
			validFeatures[k] = v
		}
	}
	l.Features = validFeatures

	h := sha256.Sum256(l.CanonicalBytes())
	sig, err := base64.RawURLEncoding.DecodeString(l.Signature)
	if err != nil {
		return License{}, err
	}

	if !ecdsa.VerifyASN1(verificationKey, h[:], sig) {
		return License{}, ErrBadSignature
	}

	if l.IsExpired(time.Now()) {
		return License{}, ErrLicenseExpired
	}

	return l, nil
}

func (l License) IsExpired(when time.Time) bool {
	return when.Before(l.ValidFrom) || when.After(l.ValidTo)
}

// CanonicalBytes returns the license data in a form suitable for signing. It
// is only exposed to simplify attacks against license signatures. Regular
// consumers of license data should never need to call this function.
func (l License) CanonicalBytes() []byte {
	// increment this when the format changes
	magic := append([]byte("5DW"), 0x01)

	var buf bytes.Buffer
	appendLengthAndValue(&buf, l.Id)
	appendLengthAndValue(&buf, l.Licensee)
	appendLengthAndValue(&buf, l.ValidFrom.UTC().Format(time.RFC3339))
	appendLengthAndValue(&buf, l.ValidTo.UTC().Format(time.RFC3339))

	features := make([]string, 0, len(l.Features))
	for _, f := range allFeatures {
		if l.Features[f] != 0 {
			features = append(features, fmt.Sprintf("%s:%d", f, l.Features[f]))
		}
	}
	slices.Sort(features)
	for _, f := range features {
		appendLengthAndValue(&buf, f)
	}

	prefix := make([]byte, 8)
	copy(prefix[0:], magic)
	binary.BigEndian.PutUint32(prefix[4:], uint32(buf.Len()))

	return append(prefix, buf.Bytes()...)
}

func appendLengthAndValue(buf *bytes.Buffer, s string) {
	b := []byte(s)
	_ = binary.Write(buf, binary.BigEndian, uint16(len(b)))
	buf.Write(b)
}

type Feature string

const (
	DLP             Feature = "DLP"
	MaxFileSizeMB           = "Max file size (MB)"
	MonthlyEgressGB         = "Monthly egress limit (GB)"
	OIDC                    = "OIDC"
	StorageLimitGB          = "Storage limit (GB)"
	StorageQuotas           = "Per-user storage quotas"
	VirusScanning           = "Virus scanning"
)

var allFeatures = []Feature{
	DLP,
	MaxFileSizeMB,
	MonthlyEgressGB,
	OIDC,
	StorageLimitGB,
	StorageQuotas,
	VirusScanning,
}
