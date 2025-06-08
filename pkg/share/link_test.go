package share

import (
	"bytes"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseLink(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	timeNow = func() time.Time { return now }

	testCases := []struct {
		name     string
		query    url.Values
		expected Link
	}{
		{"basic", values("key", "one/foo.txt", "exp", "1863795600", "sig", "5DPBGlO_GQrU-lH5DpfMdCMmHXy_2UdQvZCTsg24yBw"),
			Link{Key: "one/foo.txt", Expiration: time.Unix(1863795600, 0).UTC(), Signature: "5DPBGlO_GQrU-lH5DpfMdCMmHXy_2UdQvZCTsg24yBw"}},
		{"no expiration", values("key", "one/foo.txt"),
			NewLink("one/foo.txt", time.Time{})},
		{"unsigned", values("key", "one/foo.txt", "exp", "1863795600"),
			NewLink("one/foo.txt", time.Unix(1863795600, 0))},

		{"invalid exp", values("key", "one/foo.txt", "exp", "never"),
			NewLink("one/foo.txt", time.Unix(0, 0))},
		{"path traversal", values("key", "one/../two/foo.txt", "exp", "1863795600"),
			NewLink("one/../two/foo.txt", time.Unix(1863795600, 0))},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l := ParseLink(tc.query)

			assert.Equal(t, tc.expected, l)
		})
	}
}

func TestLink_QueryString(t *testing.T) {
	testCases := []struct {
		name          string
		link          Link
		expectedQuery string
	}{
		{"unsigned", NewLink("abc/123.txt", time.Unix(1863795600, 0)), "exp=1863795600&key=abc%2F123.txt"},
		{"no expiration", NewLink("abc/123.txt", DoesNotExpire), "key=abc%2F123.txt"},
		{"signed",
			Link{Key: "abc/123.txt", Expiration: time.Unix(1863795600, 0).UTC(), Signature: "aoHEeiag0oGJxSGBAZjLwRTaaksGJY3ZLlDbcMIsCP4"},
			"exp=1863795600&key=abc%2F123.txt&sig=aoHEeiag0oGJxSGBAZjLwRTaaksGJY3ZLlDbcMIsCP4"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			qs := tc.link.QueryString()
			assert.Equal(t, tc.expectedQuery, qs)
		})
	}
}

func TestNewSignedLink(t *testing.T) {
	secret := bytes.Repeat([]byte{0x55}, 32)
	testCases := []struct {
		name         string
		key          string
		exp          time.Time
		expectedLink Link
	}{
		{"basic", "abc/123.txt", time.Unix(1863795600, 0),
			Link{Key: "abc/123.txt", Expiration: time.Unix(1863795600, 0).UTC(), Signature: "GgoJOYB9kHcVl6ElbXP7lAFQLNdGVTCcZK-P67qkZmE"}},
		{"no expiration", "abc/123.txt", DoesNotExpire,
			Link{Key: "abc/123.txt", Signature: "g55e9vBCXDdXwHVNq3gB7jmnwq4lgJKYbw_h1_I97oU"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l := NewSignedLink(tc.key, tc.exp, secret)
			assert.Equal(t, tc.expectedLink, l)
		})
	}
}

func TestLink_Verify(t *testing.T) {
	secret := bytes.Repeat([]byte{0x55}, 32)
	testCases := []struct {
		name        string
		link        Link
		expectedErr error
	}{
		{"signed", Link{Key: "abc/123.txt", Expiration: time.Unix(1863795600, 0).UTC(), Signature: "GgoJOYB9kHcVl6ElbXP7lAFQLNdGVTCcZK-P67qkZmE"}, nil},
		{"no expiration", Link{Key: "abc/123.txt", Signature: "g55e9vBCXDdXwHVNq3gB7jmnwq4lgJKYbw_h1_I97oU"}, nil},

		{"unsigned", NewLink("abc/123.txt", time.Unix(1863795600, 0)), ErrNoSignature},
		{"expired", Link{Key: "abc/123.txt", Expiration: time.Unix(946684800, 0).UTC(), Signature: "3DBXcWH6qLV3ZAU6RHO2zfLvapVXz8E3wf1sffDv4e8"}, ErrExpired},
		{"bad signature", Link{Key: "abc/123.txt", Expiration: time.Unix(1863795600, 0).UTC(), Signature: "aoHEeiag0oGJxSGBAZjLwRTaaksGJY3ZLlDbcMIsCP4"}, ErrInvalidSignature},
		{"invalid b64", Link{Key: "abc/123.txt", Expiration: time.Unix(1863795600, 0).UTC(), Signature: "$#"}, ErrInvalidSignature},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedErr, tc.link.Verify(secret))
		})
	}
}

func values(s ...string) url.Values {
	v := make(url.Values)
	for i := 0; i < len(s); i += 2 {
		v.Set(s[i], s[i+1])
	}
	return v
}
