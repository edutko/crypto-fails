package random

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	s := String(0)
	assert.Len(t, s, 0)

	s = String(1)
	assert.Len(t, s, 1)
	assert.Regexp(t, regexp.MustCompile(`^[-_0-9A-Za-z]$`), s)

	s = String(2)
	assert.Len(t, s, 2)
	assert.Regexp(t, regexp.MustCompile(`^[-_0-9A-Za-z]+$`), s)

	s = String(255)
	assert.Len(t, s, 255)
	assert.Regexp(t, regexp.MustCompile(`^[-_0-9A-Za-z]+$`), s)
}
