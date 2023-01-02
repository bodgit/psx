package psx_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/bodgit/psx"
	"github.com/stretchr/testify/assert"
)

func TestNewWriter(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)

	w, err := psx.NewWriter(buf)
	if err != nil {
		t.Fatal(err)
	}

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(filepath.Join("testdata", "blank.mcd"))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, b, buf.Bytes())
}

func TestCopy(t *testing.T) {
	t.Parallel()

	rc, err := psx.OpenReader(filepath.Join("testdata", "MemoryCard2-1.mcd"))
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()

	wc, err := psx.NewWriter(io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer wc.Close()

	fr, err := rc.File[0].Open()
	if err != nil {
		t.Fatal(err)
	}
	defer fr.Close()

	fw, err := wc.Create()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := io.Copy(fw, fr); err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, fw.Close())
}

func ExampleWriter() {
	buf := new(bytes.Buffer)

	w, err := psx.NewWriter(buf)
	if err != nil {
		panic(err)
	}

	if err := w.Close(); err != nil {
		panic(err)
	}

	fmt.Println(buf.Len())
	// Output: 131072
}
