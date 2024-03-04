package psx

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/bodgit/psx/internal/xor"
)

var (
	headerSignature = [2]byte{'M', 'C'} //nolint:gochecknoglobals

	errBadHeaderChecksum  = errors.New("bad header frame checksum")
	errBadHeaderSignature = errors.New("bad header frame signature")
)

type headerFrame struct {
	Signature [2]byte
	_         [125]byte
	Checksum  [1]byte
}

func (hf *headerFrame) unmarshalBinary(r io.Reader) error {
	h := xor.New()

	if err := binary.Read(checksumReader(r, h, binary.Size(hf)), binary.LittleEndian, hf); err != nil {
		return err
	}

	if !bytes.Equal(hf.Signature[:], headerSignature[:]) {
		return errBadHeaderSignature
	}

	if !bytes.Equal(hf.Checksum[:], h.Sum(nil)) {
		return errBadHeaderChecksum
	}

	return nil
}

func (hf *headerFrame) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Grow(binary.Size(hf))

	_ = binary.Write(buf, binary.LittleEndian, hf)

	return buf.Bytes(), nil
}

func (hf *headerFrame) generateChecksum() ([]byte, error) {
	b, err := hf.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return checksum(b[:frameSize-1]), nil
}

func (hf *headerFrame) checksum() error {
	xor, err := hf.generateChecksum()
	if err != nil {
		return err
	}

	copy(hf.Checksum[:], xor)

	return nil
}

func newHeaderFrame() headerFrame {
	return headerFrame{
		Signature: headerSignature,
	}
}
