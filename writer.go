package psx

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"
)

var (
	errDuplicateName = errors.New("duplicate name")
	errInvalidLength = errors.New("invalid length")
	errNoFreeSpace   = errors.New("no free space")
)

type fileWriter struct {
	buf *bytes.Buffer
	w   *Writer
}

func (w *fileWriter) maxSize() int {
	return binary.Size(directoryFrame{}) + numBlocks*blockSize
}

func (w *fileWriter) Write(p []byte) (int, error) {
	if len(p)+w.buf.Len() > w.maxSize() {
		// Would exceed the maximum size
		return 0, errInvalidLength
	}

	return w.buf.Write(p) //nolint:wrapcheck
}

//nolint:cyclop,funlen
func (w *fileWriter) Close() error {
	w.w.mu.Lock()
	defer w.w.mu.Unlock()

	delete(w.w.fw, w)

	mc := w.w.mc

	df := new(directoryFrame)
	if err := binary.Read(w.buf, binary.LittleEndian, df); err != nil {
		return fmt.Errorf("unable to read header: %w", err)
	}

	for _, x := range mc.HeaderBlock.DirectoryFrame {
		if !x.isFirst() {
			continue
		}

		if x.filename() == df.filename() {
			return errDuplicateName
		}
	}

	if w.buf.Len()%blockSize != 0 || w.buf.Len() != int(df.Size) {
		return errInvalidLength
	}

	blocks := w.buf.Len() / blockSize

	if w.w.i+blocks > numBlocks {
		return errNoFreeSpace
	}

	for i := 0; i < blocks; i++ {
		frame := w.w.i + i

		if i == 0 {
			mc.HeaderBlock.DirectoryFrame[frame] = *df
		}

		lo := uint16(frame + 1)
		if i+1 == blocks {
			lo = lastLink
		}

		ab := blockMiddleLink
		if i == 0 {
			ab = blockFirstLink
		} else if i+1 == blocks && blocks > 1 {
			ab = blockLastLink
		}

		mc.HeaderBlock.DirectoryFrame[frame].LinkOrder = lo
		mc.HeaderBlock.DirectoryFrame[frame].AvailableBlocks = ab

		if _, err := io.ReadFull(w.buf, mc.DataBlock[frame][:]); err != nil {
			return fmt.Errorf("unable to read data block: %w", err)
		}
	}

	w.w.i += blocks

	return mc.checksum()
}

// A Writer is used for creating a new memory card image with files written to
// it.
type Writer struct {
	mu sync.Mutex
	w  io.Writer
	mc *memoryCard
	fw map[*fileWriter]struct{}
	i  int
}

// Create returns an io.WriteCloser for writing a new file on the memory card.
// The file should consist of a 128 byte header followed by one or more 8 KiB
// blocks as indicated in the header.
func (w *Writer) Create() (io.WriteCloser, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.i == numBlocks {
		return nil, errNoFreeSpace
	}

	fw := &fileWriter{new(bytes.Buffer), w}
	w.fw[fw] = struct{}{}

	return fw, nil
}

// Close writes out the memory card to the underlying io.Writer. Any in-flight
// open memory card files are closed first.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for fw := range w.fw {
		if err := fw.Close(); err != nil {
			return err
		}

		delete(w.fw, fw)
	}

	b, err := w.mc.MarshalBinary()
	if err != nil {
		return err
	}

	if n, err := w.w.Write(b); err != nil || n != w.mc.size() {
		if err != nil {
			return fmt.Errorf("unable to write memory card: %w", err)
		}

		return errInvalidLength
	}

	return nil
}

// NewWriter returns a Writer that will write a new memory card to w.
func NewWriter(w io.Writer) (*Writer, error) {
	mc, err := newMemoryCard()
	if err != nil {
		return nil, err
	}

	return &Writer{
		w:  w,
		mc: mc,
		fw: make(map[*fileWriter]struct{}),
	}, nil
}
