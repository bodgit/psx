package psx

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Based on https://www.psdevwiki.com/ps3/PS1_Savedata and
// http://problemkaputt.de/psx-spx.htm#memorycarddataformat

const (
	blockFirstLink byte = iota + 0x51
	blockMiddleLink
	blockLastLink
	blockAvailable   = 0xa0
	blockUnavailable = 0xff
)

const (
	lastLink        = 0xffff
	blockSize       = 0x2000
	numBlocks       = 15
	reservedBlocks  = 1
	numUnusedFrames = 20
	frameSize       = 128
	cardSize        = blockSize * (numBlocks + reservedBlocks)
)

var (
	dataSignature = [2]byte{'S', 'C'} //nolint:deadcode,gochecknoglobals,unused,varcheck

	errBadDataSignature = errors.New("bad data block signature") //nolint:deadcode,unused,varcheck
	errTrailingBytes    = errors.New("trailing bytes")
)

type headerBlock struct {
	HeaderFrame    headerFrame
	DirectoryFrame [numBlocks]directoryFrame
	UnusedFrame    [numUnusedFrames]unusedFrame
	_              [27 * frameSize]byte
	TrailingFrame  headerFrame
}

func (hb *headerBlock) unmarshalBinary(r io.Reader) error {
	if err := hb.HeaderFrame.unmarshalBinary(r); err != nil {
		return err
	}

	for i := 0; i < numBlocks; i++ {
		if err := hb.DirectoryFrame[i].unmarshalBinary(r); err != nil {
			return err
		}
	}

	for i := 0; i < numUnusedFrames; i++ {
		if err := binary.Read(r, binary.LittleEndian, &hb.UnusedFrame[i]); err != nil {
			return err
		}
	}

	if _, err := io.CopyN(io.Discard, r, 27*frameSize); err != nil {
		return err
	}

	return hb.TrailingFrame.unmarshalBinary(r)
}

type memoryCard struct {
	HeaderBlock headerBlock
	DataBlock   [numBlocks][blockSize]byte
}

func (mc *memoryCard) size() int {
	return cardSize
}

func (mc *memoryCard) count() int {
	count := 0

	for i := 0; i < numBlocks; i++ {
		if !mc.HeaderBlock.DirectoryFrame[i].isFirst() {
			continue
		}

		count++
	}

	return count
}

func (mc *memoryCard) checksum() error {
	if err := mc.HeaderBlock.HeaderFrame.checksum(); err != nil {
		return err
	}

	for i := range mc.HeaderBlock.DirectoryFrame {
		if err := mc.HeaderBlock.DirectoryFrame[i].checksum(); err != nil {
			return err
		}
	}

	return mc.HeaderBlock.TrailingFrame.checksum()
}

func (mc *memoryCard) unmarshalBinary(r io.Reader) error {
	if err := mc.HeaderBlock.unmarshalBinary(r); err != nil {
		return fmt.Errorf("unable to unmarshal header block: %w", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &mc.DataBlock); err != nil {
		return fmt.Errorf("unable to unmarshal data blocks: %w", err)
	}

	if n, _ := io.CopyN(io.Discard, r, 1); n > 0 {
		return errTrailingBytes
	}

	return nil
}

func (mc *memoryCard) UnmarshalBinary(b []byte) error {
	return mc.unmarshalBinary(bytes.NewReader(b))
}

func (mc *memoryCard) MarshalBinary() ([]byte, error) {
	b := new(bytes.Buffer)

	if err := binary.Write(b, binary.LittleEndian, mc); err != nil {
		return nil, fmt.Errorf("unable to marshal memory card: %w", err)
	}

	return b.Bytes(), nil
}

func newMemoryCard() (*memoryCard, error) {
	mc := new(memoryCard)

	mc.HeaderBlock.HeaderFrame = newHeaderFrame()

	for i := 0; i < numBlocks; i++ {
		mc.HeaderBlock.DirectoryFrame[i] = newDirectoryFrame()
	}

	for i := 0; i < numUnusedFrames; i++ {
		mc.HeaderBlock.UnusedFrame[i] = newUnusedFrame()
	}

	mc.HeaderBlock.TrailingFrame = newHeaderFrame()

	if err := mc.checksum(); err != nil {
		return nil, err
	}

	return mc, nil
}

// DetectMemoryCard works out if the io.ReaderAt r pointing to the data of size
// bytes looks sufficiently like a PlayStation 1 memory card image.
func DetectMemoryCard(r io.ReaderAt, size int64) (bool, error) {
	if size == cardSize {
		sr := io.NewSectionReader(r, 0, int64(len(headerSignature)))

		b, err := io.ReadAll(sr)
		if err != nil {
			return false, fmt.Errorf("unable to read header signature: %w", err)
		}

		if bytes.Equal(b, headerSignature[:]) {
			return true, nil
		}
	}

	return false, nil
}
