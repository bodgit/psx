package psx

import "hash"

type xor struct {
	b byte
}

func (x *xor) Write(p []byte) (int, error) {
	for _, i := range p {
		x.b ^= i
	}
	return len(p), nil
}

func (x *xor) Sum(in []byte) []byte {
	return append(in, x.b)
}

func (x *xor) Reset() {
	x.b = 0
}

func (*xor) Size() int {
	return 1
}

func (*xor) BlockSize() int {
	return 1
}

// XOR returns a hash.Hash implementation that computes a single byte
// checksum by XOR'ing every byte written to it
func XOR() hash.Hash {
	return new(xor)
}
