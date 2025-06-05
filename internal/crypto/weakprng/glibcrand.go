package weakprng

import "math/rand/v2"

// glibcRand implements a linear congruential generator using the same parameters as the glibc
// rand() function
// https://sourceware.org/git/?p=glibc.git;a=blob;f=stdlib/random_r.c;h=b6297fe099af9af238caaebb9c21764e9360812b;hb=9d94997b5f9445afd4f2bccc5fa60ff7c4361ec1#l364
type glibcRand struct {
	state int32
}

func NewGlibcRand() rand.Source {
	return NewGlibcRandWithSeed(1)
}

func NewGlibcRandWithSeed(seed int32) rand.Source {
	return &glibcRand{seed}
}

func (r *glibcRand) next() int32 {
	val := ((r.state * 1103515245) + 12345) & 0x7fffffff
	r.state = val
	return val
}

func (r *glibcRand) Uint64() uint64 {
	// every fourth byte has its high bit cleared 🫠
	return uint64(r.next())<<32 + uint64(r.next())
}
