package weakprng

import (
	"encoding/binary"
	"math/rand/v2"
)

type deterministicSource struct {
	state  []byte
	offset int
}

func NewDeterministicSource() rand.Source {
	return NewDeterministicSourceWithSeed(defaultDeterministicSeed)
}

func NewDeterministicSourceWithSeed(seed []byte) rand.Source {
	if len(seed)%8 != 0 {
		padLen := 8 - (len(seed) % 8)
		seed = append(seed, make([]byte, padLen)...)
	}
	return &deterministicSource{seed, 0}
}

func (r *deterministicSource) Uint64() uint64 {
	result := binary.BigEndian.Uint64(r.state[r.offset : r.offset+8])
	r.offset += 8
	if r.offset >= len(r.state) {
		r.offset = 0
	}
	return result
}

var defaultDeterministicSeed = []byte{0x36, 0x8b, 0x7f, 0x45, 0xa9, 0xdd, 0xa2, 0x6d}
