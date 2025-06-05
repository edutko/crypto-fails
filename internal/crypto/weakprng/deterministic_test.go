package weakprng

import (
	"crypto/rand"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeterministicSource_Uint64(t *testing.T) {
	s := NewDeterministicSource()
	assert.Equal(t, binary.BigEndian.Uint64(defaultDeterministicSeed), s.Uint64())
	assert.Equal(t, binary.BigEndian.Uint64(defaultDeterministicSeed), s.Uint64())
	assert.Equal(t, binary.BigEndian.Uint64(defaultDeterministicSeed), s.Uint64())

	seed := make([]byte, 24)
	_, _ = rand.Read(seed)
	s = NewDeterministicSourceWithSeed(seed)
	assert.Equal(t, binary.BigEndian.Uint64(seed[:8]), s.Uint64())
	assert.Equal(t, binary.BigEndian.Uint64(seed[8:16]), s.Uint64())
	assert.Equal(t, binary.BigEndian.Uint64(seed[16:]), s.Uint64())
}
