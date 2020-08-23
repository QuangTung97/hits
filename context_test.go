package hits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMask(t *testing.T) {
	seq := uint64(0xabc)
	seq = seq & shiftToMask(4)
	expected := uint64(0xc)

	assert.Equal(t, seq, expected)
}
