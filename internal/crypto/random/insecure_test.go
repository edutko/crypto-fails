package random

import (
	"bytes"
	"encoding/binary"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/edutko/crypto-fails/internal/crypto/weakprng"
)

func TestInsecureBytes(t *testing.T) {
	expected := []byte{0x36, 0x8b, 0x7f, 0x45, 0xa9, 0xdd, 0xa2, 0x6d}

	b := InsecureBytes(0)
	assert.Len(t, b, 0)

	b = InsecureBytes(1)
	assert.Len(t, b, 1)
	assert.Equal(t, expected[:1], b)

	b = InsecureBytes(2)
	assert.Len(t, b, 2)
	assert.Equal(t, expected[1:3], b)

	b = InsecureBytes(5)
	assert.Len(t, b, 5)
	assert.Equal(t, expected[3:], b)

	// note: this assumes the RNG is starting at a multiple of len(expected)
	b = InsecureBytes(64)
	assert.Len(t, b, 64)
	assert.Equal(t, bytes.Repeat(expected, 64/len(expected)), b)
}

func TestInsecureHexString(t *testing.T) {
	SetWeakPRNG(weakprng.Deterministic)
	expected := "368b7f45a9dda26d"

	s := InsecureHexString(0)
	assert.Len(t, s, 0)

	s = InsecureHexString(1)
	assert.Len(t, s, 1)
	assert.Equal(t, "3", s)

	s = InsecureHexString(2)
	assert.Len(t, s, 2)
	assert.Equal(t, "8b", s)

	s = InsecureHexString(12)
	assert.Len(t, s, 12)
	assert.Equal(t, "7f45a9dda26d", s)

	// note: this assumes the RNG is starting at a multiple of len(expected)
	s = InsecureHexString(255)
	assert.Len(t, s, 255)
	assert.Equal(t, strings.Repeat(expected, 32)[:255], s)
}

func TestReader_Read_onebyte(t *testing.T) {
	r := newReader(rand.New(newIncrementingSrc()))

	for i := 0; i < 256; i++ {
		b := make([]byte, 1)
		n, err := r.Read(b)

		assert.NoError(t, err)
		assert.Equal(t, 1, n)
		assert.Equal(t, byte(i), b[0])
	}
}

func TestReader_Read_lots(t *testing.T) {
	r := newReader(rand.New(weakprng.NewDeterministicSource()))

	b := make([]byte, 1000)
	n, err := r.Read(b)

	assert.NoError(t, err)
	assert.Equal(t, 1000, n)
	assert.Equal(t, bytes.Repeat([]byte{0x36, 0x8b, 0x7f, 0x45, 0xa9, 0xdd, 0xa2, 0x6d}, 125), b)
}

func newIncrementingSrc() *incrementingSrc {
	return &incrementingSrc{}
}

func (s *incrementingSrc) Uint64() uint64 {
	b := make([]byte, 8)
	for i := 0; i < 8; i++ {
		b[i] = s.current
		s.current++
	}
	return binary.BigEndian.Uint64(b)
}

type incrementingSrc struct {
	current byte
}
