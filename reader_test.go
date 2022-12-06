package psx_test

import (
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/bodgit/psx"
)

func TestFS(t *testing.T) {
	t.Parallel()

	rc, err := psx.OpenReader(filepath.Join("testdata", "MemoryCard2-1.mcd"))
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()

	if err := fstest.TestFS(rc, "BESLES-00024TOMBRAID", "BESCES-01237TEKKEN-3"); err != nil {
		t.Fatal(err)
	}
}
