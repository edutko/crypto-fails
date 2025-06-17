package app

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewLicense(t *testing.T) {
	uuidv4Pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	testCases := []struct {
		name     string
		licensee string
		start    time.Time
		end      time.Time
		features map[Feature]int
	}{
		{"minimal", "ACME, Inc.", time.Now(), time.Now().Add(90 * 24 * time.Hour),
			map[Feature]int{},
		},
		{"multiple features", "ACME, Inc.", time.Now(), time.Now().Add(90 * 24 * time.Hour),
			map[Feature]int{DLP: 1, StorageLimitGB: 1000, OIDC: 1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expected := License{
				Licensee:  tc.licensee,
				ValidFrom: tc.start.UTC().Truncate(24 * time.Hour),
				ValidTo:   tc.end.UTC().Truncate(24 * time.Hour),
				Features:  tc.features,
			}

			l := NewLicense(tc.licensee, tc.start, tc.end, tc.features)

			assert.Regexp(t, uuidv4Pattern, l.Id)
			l.Id = ""
			assert.Equal(t, expected, l)
		})
	}
}

func TestParseLicense(t *testing.T) {
	testCases := []struct {
		name        string
		id          string
		licensee    string
		start       time.Time
		end         time.Time
		features    map[Feature]int
		signature   string
		expectedErr error
	}{
		{"minimal",
			"60cb8225-b819-4291-a9ca-1ae6f1739d4b", "ACME, Inc.", time.Now(), time.Now().Add(90 * 24 * time.Hour),
			map[Feature]int{}, "", nil,
		},
		{"multiple features",
			"a7f76d0-5ca5-4045-b3d7-a77b5bc971b1", "ACME, Inc.", time.Now(), time.Now().Add(90 * 24 * time.Hour),
			map[Feature]int{DLP: 1, StorageLimitGB: 1000, OIDC: 1}, "", nil,
		},
		{"not yet valid",
			"a3546a01-88c8-45d1-83b4-5afc9762b18d", "ACME, Inc.", time.Now().Add(25 * time.Hour), time.Now().Add(90 * 24 * time.Hour),
			map[Feature]int{}, "", ErrLicenseExpired,
		},
		{"expired",
			"12b8ba70-23bb-4e1d-8ec4-a89c3ccc4a33", "ACME, Inc.", time.Now(), time.Now().Add(-1 * time.Minute),
			map[Feature]int{}, "", ErrLicenseExpired,
		},
		{"invalid signature",
			"12b8ba70-23bb-4e1d-8ec4-a89c3ccc4a33", "ACME, Inc.", time.Now(), time.Now().Add(90 * 24 * time.Hour),
			map[Feature]int{DLP: 1, StorageLimitGB: 1000, OIDC: 1},
			"MEUCIHkcm3CJYMicNE7WMvKR3NI58xLyi4RLmKkNsXwPhdN5AiEA86VPlSC8gW4PxMskGg-UDk29QjK_jCHXpKMjVFPJmUE", ErrBadSignature,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expected := License{
				Id:        tc.id,
				Licensee:  tc.licensee,
				ValidFrom: tc.start.UTC().Truncate(24 * time.Hour),
				ValidTo:   tc.end.UTC().Truncate(24 * time.Hour),
				Features:  tc.features,
				Signature: tc.signature,
			}
			if tc.signature == "" {
				expected.Signature = base64.RawURLEncoding.EncodeToString(sign(expected, key))
			}
			b, _ := json.Marshal(expected)

			l, err := ParseLicense(b, pubKey)

			if tc.expectedErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, expected, l)
			} else {
				assert.NotNil(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
			}
		})
	}
}

var key = &ecdsa.PrivateKey{
	PublicKey: ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     bigintFromHex("8473a4cb906a92dcf9b3c2a6f7e0e755bd842aa65b38bc4fb03c68a111ebc187"),
		Y:     bigintFromHex("a4d2beaddccd478707c6e8fec0b4f9f11235091c8372f550940f694935de11fd"),
	},
	D: bigintFromHex("041eb7fe4ce0fab48fd397bb9d89c174ba2a4311cf25b4ccb84cab8d5bab6f24"),
}
var pubKey = &key.PublicKey

func bigintFromHex(s string) *big.Int {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return big.NewInt(0).SetBytes(b)
}

func sign(l License, priv *ecdsa.PrivateKey) []byte {
	h := sha256.Sum256(l.CanonicalBytes())
	sig, err := ecdsa.SignASN1(rand.Reader, priv, h[:])
	if err != nil {
		panic(err)
	}
	return sig
}
