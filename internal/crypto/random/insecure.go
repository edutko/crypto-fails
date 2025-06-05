package random

import (
	"encoding/binary"
	"encoding/hex"
	"io"
	"math/rand/v2"

	"github.com/edutko/crypto-fails/internal/crypto/weakprng"
)

func InsecureBytes(length int) []byte {
	b := make([]byte, length)
	_, _ = insecure.Read(b)
	return b
}

func InsecureHexString(length int) string {
	b := InsecureBytes((length + 1) / 2)
	return hex.EncodeToString(b)[:length]
}

func SetWeakPRNG(alg weakprng.Algorithm) {
	var src rand.Source
	switch alg {
	case weakprng.GlibcRand:
		src = weakprng.NewGlibcRand()
	case weakprng.Java8Random:
		src = weakprng.NewJavaRandom()
	case weakprng.MT19937:
		src = weakprng.NewMT19937()
	case weakprng.XORShift128p:
		src = weakprng.NewXORShift128p()
	default:
		src = weakprng.NewDeterministicSource()
	}

	insecure = newReader(rand.New(src))
}

func newReader(rnd *rand.Rand) io.Reader {
	return &reader{rnd, make([]byte, 0, 8)}
}

func (r *reader) Read(p []byte) (n int, err error) {
	if len(p) <= len(r.buf) {
		copy(p, r.buf[:len(p)])
		r.buf = r.buf[len(p):]
		return len(p), nil
	}

	for len(p) > len(r.buf) {
		i := r.rnd.Uint64()
		r.buf = binary.BigEndian.AppendUint64(r.buf, i)
	}
	copy(p, r.buf[:len(p)])
	r.buf = r.buf[len(p):]
	return len(p), nil
}

type reader struct {
	rnd *rand.Rand
	buf []byte
}

var insecure = newReader(rand.New(weakprng.NewDeterministicSource()))
