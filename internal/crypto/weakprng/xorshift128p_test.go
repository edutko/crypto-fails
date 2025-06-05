package weakprng

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXorShift128p_Uint64(t *testing.T) {
	testCases := []struct {
		seed     uint64
		expected []uint64
	}{
		{1, []uint64{0x787caa13083f2034, 0xd28f42aee9855ff9, 0x4f8e769a08c71d33, 0xc46c323ff0b001cd}},
		{12345, []uint64{0x99b96397027d204a, 0xedafa41bebb2ed00, 0x9373b209985c75ed, 0x33bde1e9458f739a}},
		{0x7fffffff, []uint64{0x3233230ea2d9695a, 0xc8b03f5968f58493, 0x1e0b936795dd9e58, 0xec90203ebd3f6722}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tc.seed), func(t *testing.T) {
			s := NewXORShift128pWithSeed(tc.seed)
			for i := range len(tc.expected) {
				assert.Equal(t, tc.expected[i], s.Uint64())
			}
		})
	}
}
