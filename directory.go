package psx

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/bodgit/psx/internal/xor"
)

var errBadDirectoryChecksum = errors.New("bad directory frame checksum")

type directoryFrame struct {
	AvailableBlocks byte
	_               [3]byte
	Size            uint32
	LinkOrder       uint16
	CountryCode     [2]byte
	ProductCode     [10]byte
	Identifier      [8]byte
	_               [97]byte
	Checksum        [1]byte
}

func (df *directoryFrame) unmarshalBinary(r io.Reader) error {
	h := xor.New()

	if err := binary.Read(checksumReader(r, h, binary.Size(df)), binary.LittleEndian, df); err != nil {
		return err
	}

	if !bytes.Equal(df.Checksum[:], h.Sum(nil)) {
		return errBadDirectoryChecksum
	}

	return nil
}

func (df *directoryFrame) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Grow(binary.Size(df))

	_ = binary.Write(buf, binary.LittleEndian, df)

	return buf.Bytes(), nil
}

func (df *directoryFrame) generateChecksum() ([]byte, error) {
	b, err := df.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return checksum(b[:frameSize-1]), nil
}

func (df *directoryFrame) checksum() error {
	xor, err := df.generateChecksum()
	if err != nil {
		return err
	}

	copy(df.Checksum[:], xor)

	return nil
}

//nolint:unused
func (df *directoryFrame) isEmpty() bool {
	return df.AvailableBlocks == blockAvailable
}

func (df *directoryFrame) isFirst() bool {
	return df.AvailableBlocks == blockFirstLink
}

func (df *directoryFrame) countryCode() string {
	return string(df.CountryCode[:])
}

func (df *directoryFrame) productCode() string {
	return string(df.ProductCode[:])
}

func (df *directoryFrame) identifier() string {
	return string(df.Identifier[:])
}

func (df *directoryFrame) filename() string {
	return df.countryCode() + df.productCode() + df.identifier()
}

func newDirectoryFrame() directoryFrame {
	return directoryFrame{
		AvailableBlocks: blockAvailable,
		LinkOrder:       lastLink,
	}
}
