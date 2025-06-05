package weakprng

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJavaRandom_Uint64(t *testing.T) {
	testCases := []struct {
		seed     int64
		expected []uint64
	}{
		{1, []uint64{0xbb1ad57319b89cd8, 0x68fb0e6f684df992, 0x352cccfc0946b8f0, 0x552cf1e4a8ab85dd}},
		{12345, []uint64{0x5c9f20d58361b331, 0xeed8a921eac80778, 0xd545798e09a4ef89, 0x5393ea7f1fea4064}},
		{0x7fffffff, []uint64{0x9ebeecb2e83ea2c3, 0xd945a016d18f7f9e, 0xcc3bde68a57d3b4c, 0x1155339a0c33450c}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tc.seed), func(t *testing.T) {
			s := NewJavaRandomWithSeed(tc.seed)
			for i := range len(tc.expected) {
				assert.Equal(t, tc.expected[i], s.Uint64())
			}
		})
	}
}

// Java code for verification
//
//import java.util.Random;
//
//public class Main {
//    public static void main(String[] args) {
//        int seed = 1;
//        if (args.length > 0) {
//            seed = Integer.parseInt(args[1]);
//        }
//        java.util.Random rng = new Random(seed);
//        System.out.print("[]uint64{");
//        for (int i = 1; i <= 4; i++) {
//            System.out.printf("0x%016x, ", rng.nextLong());
//        }
//        System.out.println("}");
//    }
//}
