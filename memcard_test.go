package psx_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bodgit/psx"
	"github.com/stretchr/testify/assert"
)

func TestDetectMemoryCard(t *testing.T) {
	t.Parallel()

	tables := []struct {
		file string
	}{
		{
			filepath.Join("testdata", "blank.mcd"),
		},
	}

	for _, table := range tables {
		table := table
		t.Run(table.file, func(t *testing.T) {
			t.Parallel()

			f, err := os.Open(table.file)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			fi, err := f.Stat()
			if err != nil {
				t.Fatal(err)
			}

			ok, err := psx.DetectMemoryCard(f, fi.Size())
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, true, ok)
		})
	}
}
