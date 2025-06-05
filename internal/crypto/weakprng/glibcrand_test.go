package weakprng

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlibcRand_Uint64(t *testing.T) {
	testCases := []struct {
		seed     int32
		expected []uint64
	}{
		{1, []uint64{0x41c67ea6167eb0e7, 0x2781e494446b9b3d, 0x794bdf3215fb7483, 0x59e2b6001cfbae39}},
		{12345, []uint64{0x53dc167e270427df, 0x56651c2c0daa96f5, 0x421f1c8a3ead62fb, 0x4d1dcf182f5aad71}},
		{0x7fffffff, []uint64{0x3e39e1cc11397c15, 0x26866b2a685e9d1b, 0x22094eb86e42c491, 0x23780ff67d3feff7}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tc.seed), func(t *testing.T) {
			s := NewGlibcRandWithSeed(tc.seed)
			for i := range len(tc.expected) {
				assert.Equal(t, tc.expected[i], s.Uint64())
			}
		})
	}
}

// C code for verification
//
//#include <stdio.h>
//#include <stdlib.h>
//
//int main(int argc, char* argv[]) {
//    int seed = 1;
//    if (argc > 1) {
//        seed = atoi(argv[1]);
//    }
//
//    int64_t s = 0;
//    initstate(seed, (char*)&s, 8);
//    printf("[]uint64{");
//    for (int i = 0; i < 4; i++) {
//        printf("0x%08x%08x, ", rand(), rand());
//    }
//    printf("}\n");
//    return 0;
//}
