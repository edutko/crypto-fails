package pkcs7

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPad(t *testing.T) {
	testCases := []struct {
		name   string
		data   string
		padded []byte
	}{
		{"empty string", "", bytes.Repeat([]byte{0x10}, 16)},
		{"short", "abc", append([]byte("abc"), bytes.Repeat([]byte{0xd}, 13)...)},
		{"long", "abcdefghijklmnop", append([]byte("abcdefghijklmnop"), bytes.Repeat([]byte{0x10}, 16)...)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := Pad([]byte(tc.data), 16)

			assert.Equal(t, tc.padded, actual)
		})
	}
}

func TestUnpad(t *testing.T) {
	testCases := []struct {
		name        string
		padded      []byte
		data        []byte
		expectedErr error
	}{
		{"empty string", bytes.Repeat([]byte{0x10}, 16), []byte(""), nil},
		{"short", append([]byte("abc"), bytes.Repeat([]byte{0xd}, 13)...), []byte("abc"), nil},
		{"long", append([]byte("abcdefghijklmnop"), bytes.Repeat([]byte{0x10}, 16)...), []byte("abcdefghijklmnop"), nil},
		{"bad padding 1", []byte("abcdefghijkl\x01\x02\x03\x04"), nil, ErrBadPadding},
		{"bad padding 2", []byte("abcdefghijklmno\xff"), nil, ErrBadPadding},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := Unpad(tc.padded, 16)

			assert.Equal(t, tc.data, actual)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}
