package psx_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bodgit/psx"
	"github.com/stretchr/testify/assert"
)

func TestWriter(t *testing.T) {
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
