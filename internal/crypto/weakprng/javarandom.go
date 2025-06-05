package weakprng

import (
	"math/rand/v2"
	"time"
)

// javaRandom implements a linear congruential generator using the same parameters as Java's
// java.util.Random
//
// https://hg.openjdk.org/jdk8/jdk8/jdk/file/tip/src/share/classes/java/util/Random.java
type javaRandom struct {
	state int64
}

func NewJavaRandom() rand.Source {
	return NewJavaRandomWithSeed(8682522807148012 ^ time.Now().UnixNano())
}

func NewJavaRandomWithSeed(seed int64) rand.Source {
	seed = (seed ^ 0x5DEECE66D) & 0xFFFFFFFFFFFF
	return &javaRandom{seed}
}

func (r *javaRandom) next() int32 {
	val := ((r.state * 0x5DEECE66D) + 0xB) & 0xFFFFFFFFFFFF
	r.state = val
	return int32(val >> 16)
}

func (r *javaRandom) Uint64() uint64 {
	return (uint64(r.next()) << 32) + uint64(r.next())
}
