package psx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemoryCard(t *testing.T) {
	blank, err := os.ReadFile(filepath.Join("testdata", "blank.mcd"))
	if err != nil {
		t.Fatal(err)
	}

	mc, err := NewMemoryCard()
	if err != nil {
		t.Fatal(err)
	}

	b, err := mc.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, blank, b)
}

func TestUnmarshalBinary(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "blank.mcd"))
	if err != nil {
		t.Fatal(err)
	}

	mc, err := NewMemoryCard()
	if err != nil {
		t.Fatal(err)
	}

	if err := mc.UnmarshalBinary(b); err != nil {
		t.Fatal(err)
	}
}
