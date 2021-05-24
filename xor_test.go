package psx

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXOR(t *testing.T) {
	h := XOR()

	assert.Equal(t, 1, h.Size())
	assert.Equal(t, 1, h.BlockSize())

	_, err := h.Write(append([]byte{'M', 'C'}, bytes.Repeat([]byte{0}, 125)...))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, []byte{0x0e}, h.Sum(nil))

	h.Reset()

	assert.Equal(t, []byte{0x00}, h.Sum(nil))
}
