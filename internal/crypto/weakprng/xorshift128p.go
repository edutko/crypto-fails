package weakprng

import (
	cryptorand "crypto/rand"
	"math/big"
	"math/rand/v2"
)

// xorShift128p implements Marsaglia's Xorshift128+ PRNG using the same shifts as V8
// http://vigna.di.unimi.it/ftp/papers/xorshiftplus.pdf
//
// This variant produces uint64 values rather than float64s, which means that
// it is slightly easier to attack from scratch but cannot be directly attacked
// using tools intended for predicting V8's radom numbers.
type xorShift128p struct {
	state [2]uint64
}

func NewXORShift128p() rand.Source {
	return NewXORShift128pWithSeed(randomSeed())
}

func NewXORShift128pWithSeed(seed uint64) rand.Source {
	// https://github.com/v8/v8/blob/13.9.99/src/base/utils/random-number-generator.cc#L220-L225
	s0 := murmurHash3(seed)
	return &xorShift128p{[2]uint64{
		s0, murmurHash3(^s0),
	}}
}

func (r *xorShift128p) Uint64() uint64 {
	result := r.state[0] + r.state[1]

	// https://github.com/v8/v8/blob/13.9.99/src/base/utils/random-number-generator.cc#L220-L225
	s1 := r.state[0]
	s0 := r.state[1]
	r.state[0] = s0
	s1 ^= s1 << 23
	s1 ^= s1 >> 17
	s1 ^= s0
	s1 ^= s0 >> 26
	r.state[1] = s1

	return result
}

// https://github.com/v8/v8/blob/13.9.99/src/base/utils/random-number-generator.cc#L228-L235
func murmurHash3(h uint64) uint64 {
	h ^= h >> 33
	h *= uint64(0xFF51AFD7ED558CCD)
	h ^= h >> 33
	h *= uint64(0xC4CEB9FE1A85EC53)
	h ^= h >> 33
	return h
}

func randomSeed() uint64 {
	i, err := cryptorand.Int(cryptorand.Reader, big.NewInt(0).Lsh(big.NewInt(1), 64))
	if err != nil {
		panic(err)
	}
	return i.Uint64()
}
