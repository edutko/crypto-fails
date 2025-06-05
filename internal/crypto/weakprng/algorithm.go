package weakprng

type Algorithm string

const (
	Deterministic Algorithm = "deterministic"
	GlibcRand               = "glibc-rand"
	Java8Random             = "java8"
	XORShift128p            = "xorshift128p"
	MT19937                 = "mt19937"
)
