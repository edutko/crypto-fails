package weakprng

import "github.com/goark/mt/v2/mt19937"

func NewMT19937() *mt19937.Source {
	return NewMT19937WithSeed(randomSeed())
}

func NewMT19937WithSeed(seed uint64) *mt19937.Source {
	return mt19937.New(int64(seed))
}
