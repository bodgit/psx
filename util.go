package psx

import (
	"bytes"
	"io"

	"github.com/bodgit/psx/internal/xor"
)

func checksum(b []byte) []byte {
	h := xor.New()

	_, _ = io.Copy(h, bytes.NewReader(b))

	return h.Sum(nil)
}
