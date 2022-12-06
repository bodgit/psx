package psx

import (
	"bytes"
	"io"

	"github.com/bodgit/psx/internal/hash"
)

func checksum(b []byte) []byte {
	h := hash.New()

	_, _ = io.Copy(h, bytes.NewReader(b))

	return h.Sum(nil)
}
