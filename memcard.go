package psx

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// Based on https://www.psdevwiki.com/ps3/PS1_Savedata and
// http://problemkaputt.de/psx-spx.htm#memorycarddataformat

const (
	// BlockFirstLink signifies the first block in either a single or multi-block save
	BlockFirstLink = iota + 0x51
	// BlockMiddleLink signifies any block that is neither the first nor last block in a multi-block save
	BlockMiddleLink
	// BlockLastLink signifies the last block in a multi-block save
	BlockLastLink
	// BlockAvailable is used for any free blocks on the memory card
	BlockAvailable = 0xa0
	// BlockUnavailable is used for any block that cannot be used
	BlockUnavailable = 0xff
)

const (
	// LastLink marks the last (or only) block in a save
	LastLink = 0xffff
	// NumBlocks represents the total number of usable blocks on a PlayStation 1 memory card
	NumBlocks       = 15
	numUnusedFrames = 20
	frameSize       = 128
)

var (
	headerSignature = [2]byte{'M', 'C'}
	dataSignature   = [2]byte{'S', 'C'}
)

var (
	errBadHeaderChecksum    = errors.New("bad header frame checksum")
	errBadDirectoryChecksum = errors.New("bad directory frame checksum")
)

type headerFrameFields struct {
	Signature [2]byte
	_         [125]byte
}

type headerFrame struct {
	headerFrameFields
	Checksum byte
}

func (h headerFrame) isValid() bool {
	return bytes.Equal(h.Signature[:], headerSignature[:])
}

type directoryFrameFields struct {
	AvailableBlocks byte
	_               [3]byte
	Use             [4]byte
	LinkOrder       uint16
	CountryCode     [2]byte
	ProductCode     [10]byte
	Identifier      [8]byte
	_               [97]byte
}

type directoryFrame struct {
	directoryFrameFields
	Checksum byte
}

func (d *directoryFrame) UpdateChecksum() {
	h := XOR()
	binary.Write(h, binary.LittleEndian, d.directoryFrameFields)
	d.Checksum = h.Sum(nil)[0]
}

type unusedFrame struct {
	AvailableBlocks byte
	Reserved        [3]byte
	_               [4]byte
	LinkOrder       uint16
	_               [118]byte
}

type headerBlock struct {
	HeaderFrame    headerFrame
	DirectoryFrame [NumBlocks]directoryFrame
	UnusedFrame    [numUnusedFrames]unusedFrame
	_              [27 * frameSize]byte
	TrailingFrame  headerFrame
}

type dataBlock struct {
	Signature [2]byte
	SaveData  [8190]byte
}

func (d dataBlock) isValid() bool {
	return bytes.Equal(d.Signature[:], dataSignature[:])
}

// MemoryCard represents a PlayStation 1 memory card
type MemoryCard struct {
	HeaderBlock headerBlock
	DataBlock   [NumBlocks]dataBlock
}

func (m MemoryCard) isValid() bool {
	if !m.HeaderBlock.HeaderFrame.isValid() {
		return false
	}

	for i := 0; i < NumBlocks; i++ {
		if m.HeaderBlock.DirectoryFrame[i].AvailableBlocks == BlockFirstLink {
			if !m.DataBlock[i].isValid() {
				return false
			}
		}
	}

	return true
}

// UnmarshalBinary decodes the memory card from binary form
func (m *MemoryCard) UnmarshalBinary(b []byte) error {
	r := bytes.NewReader(b)
	if err := binary.Read(r, binary.LittleEndian, m); err != nil {
		return err
	}

	h := XOR()

	binary.Write(h, binary.LittleEndian, m.HeaderBlock.HeaderFrame.headerFrameFields)
	if h.Sum(nil)[0] != m.HeaderBlock.HeaderFrame.Checksum {
		return errBadHeaderChecksum
	}

	for i := 0; i < NumBlocks; i++ {
		h.Reset()
		binary.Write(h, binary.LittleEndian, m.HeaderBlock.DirectoryFrame[i].directoryFrameFields)
		if h.Sum(nil)[0] != m.HeaderBlock.DirectoryFrame[i].Checksum {
			return errBadDirectoryChecksum
		}
	}

	return nil
}

// MarshalBinary encodes the memory card into binary form and returns the
// result
func (m *MemoryCard) MarshalBinary() ([]byte, error) {
	b := new(bytes.Buffer)

	if err := binary.Write(b, binary.LittleEndian, m); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// NewMemoryCard returns a correctly initialized blank memory card
func NewMemoryCard() (*MemoryCard, error) {
	m := new(MemoryCard)

	h := XOR()

	hf := headerFrame{
		headerFrameFields: headerFrameFields{
			Signature: headerSignature,
		},
	}
	_ = binary.Write(h, binary.LittleEndian, hf.headerFrameFields)
	hf.Checksum = h.Sum(nil)[0]

	m.HeaderBlock.HeaderFrame = hf

	h.Reset()

	df := directoryFrame{
		directoryFrameFields: directoryFrameFields{
			AvailableBlocks: BlockAvailable,
			LinkOrder:       LastLink,
		},
	}
	_ = binary.Write(h, binary.LittleEndian, df.directoryFrameFields)
	df.Checksum = h.Sum(nil)[0]

	for i := 0; i < NumBlocks; i++ {
		m.HeaderBlock.DirectoryFrame[i] = df
	}

	uf := unusedFrame{
		AvailableBlocks: BlockUnavailable,
		Reserved:        [3]byte{0xff, 0xff, 0xff},
		LinkOrder:       LastLink,
	}

	for i := 0; i < numUnusedFrames; i++ {
		m.HeaderBlock.UnusedFrame[i] = uf
	}

	m.HeaderBlock.TrailingFrame = hf

	return m, nil
}
