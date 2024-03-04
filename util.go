package psx

import (
	"bytes"
	"hash"
	"io"

	"github.com/bodgit/psx/internal/xor"
)

func checksum(b []byte) []byte {
	h := xor.New()

	_, _ = io.Copy(h, bytes.NewReader(b))

	return h.Sum(nil)
}

func checksumReader(r io.Reader, h hash.Hash, size int) io.Reader {
	return io.MultiReader(io.TeeReader(io.LimitReader(r, int64(size)-1), h), r)
}
