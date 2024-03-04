package xor

import "hash"

const (
	// BlockSize is the preferred block size.
	BlockSize = 1
	// Size is the size of the checksum in bytes.
	Size = 1
)

type digest struct {
	b byte
}

func (d *digest) BlockSize() int { return BlockSize }

func (d *digest) Reset() {
	d.b = 0
}

func (d *digest) Size() int { return Size }

func (d *digest) Sum(data []byte) []byte {
	return append(data, d.b)
}

func (d *digest) Write(p []byte) (int, error) {
	for _, i := range p {
		d.b ^= i
	}

	return len(p), nil
}

// New returns a hash.Hash implementation that computes a single byte
// checksum by XOR'ing every byte written to it.
func New() hash.Hash {
	d := new(digest)
	d.Reset()

	return d
}
