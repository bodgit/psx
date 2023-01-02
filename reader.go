package psx

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"sort"
	"sync"
	"time"
)

type fileReader struct {
	io.ReadCloser
	f *File
}

func (fr *fileReader) Stat() (fs.FileInfo, error) {
	return headerFileInfo{&fr.f.FileHeader}, nil
}

// A File is a single file within a memory card.
type File struct {
	FileHeader
	CountryCode string
	ProductCode string
	Identifier  string

	r *Reader
	i int
}

// Open returns an fs.File that provides access to the File's contents. The
// file is prefixed with a 128 byte header (the directory frame) followed by
// one or more 8 KiB blocks. Multiple files may be read concurrently.
func (f *File) Open() (fs.File, error) {
	blocks := make([]int, 0, numBlocks)
	blocks = append(blocks, f.i)

	for i := f.r.mc.HeaderBlock.DirectoryFrame[f.i].LinkOrder; i != lastLink; i = f.r.mc.HeaderBlock.DirectoryFrame[i].LinkOrder { //nolint:lll
		blocks = append(blocks, int(i))
	}

	readers := make([]io.Reader, 0, len(blocks)+1)

	b, err := f.r.mc.HeaderBlock.DirectoryFrame[f.i].MarshalBinary()
	if err != nil {
		return nil, err
	}

	readers = append(readers, bytes.NewReader(b))

	for _, block := range blocks {
		readers = append(readers, bytes.NewReader(f.r.mc.DataBlock[block][:]))
	}

	return &fileReader{io.NopCloser(io.MultiReader(readers...)), f}, nil
}

// FileHeader describes a file within a memory card.
type FileHeader struct {
	Name     string
	Modified time.Time
	Size     int64
}

// FileInfo returns an fs.FileInfo for the FileHeader.
func (h *FileHeader) FileInfo() fs.FileInfo {
	return headerFileInfo{h}
}

// Mode returns the permission and mode bits for the FileHeader.
func (h *FileHeader) Mode() fs.FileMode {
	return 0o444 //nolint:gomnd
}

type headerFileInfo struct {
	fh *FileHeader
}

func (fi headerFileInfo) Name() string               { return path.Base(fi.fh.Name) }
func (fi headerFileInfo) Size() int64                { return fi.fh.Size }
func (fi headerFileInfo) IsDir() bool                { return fi.Mode().IsDir() }
func (fi headerFileInfo) ModTime() time.Time         { return fi.fh.Modified.UTC() }
func (fi headerFileInfo) Mode() fs.FileMode          { return fi.fh.Mode() }
func (fi headerFileInfo) Type() fs.FileMode          { return fi.fh.Mode().Type() }
func (fi headerFileInfo) Sys() interface{}           { return fi.fh }
func (fi headerFileInfo) Info() (fs.FileInfo, error) { return fi, nil }

type fileListEntry struct {
	name  string
	file  *File
	isDir bool
	isDup bool
}

type fileInfoDirEntry interface {
	fs.FileInfo
	fs.DirEntry
}

func (e *fileListEntry) stat() (fileInfoDirEntry, error) { //nolint:ireturn
	if e.isDup {
		return nil, errors.New(e.name + ": duplicate entries in memory card") //nolint:goerr113
	}

	if !e.isDir {
		return headerFileInfo{&e.file.FileHeader}, nil
	}

	return e, nil
}

func (e *fileListEntry) Name() string {
	_, elem := split(e.name)

	return elem
}

func (e *fileListEntry) Size() int64       { return 0 }
func (e *fileListEntry) Mode() fs.FileMode { return fs.ModeDir | 0o555 } //nolint:gomnd
func (e *fileListEntry) Type() fs.FileMode { return fs.ModeDir }
func (e *fileListEntry) IsDir() bool       { return true }
func (e *fileListEntry) Sys() interface{}  { return nil }

func (e *fileListEntry) ModTime() time.Time {
	if e.file == nil {
		return time.Time{}
	}

	return e.file.FileHeader.Modified.UTC()
}

func (e *fileListEntry) Info() (fs.FileInfo, error) { return e, nil }

// A Reader serves content from a memory card image.
type Reader struct {
	File []*File

	mc *memoryCard

	fileListOnce sync.Once
	fileList     []fileListEntry
}

func (r *Reader) init(nr io.Reader) error {
	r.mc = new(memoryCard)

	if err := r.mc.unmarshalBinary(nr); err != nil {
		return err
	}

	r.File = make([]*File, 0, r.mc.count())

	for i := range r.mc.HeaderBlock.DirectoryFrame {
		df := r.mc.HeaderBlock.DirectoryFrame[i]

		if !df.isFirst() {
			continue
		}

		f := &File{r: r, i: i}
		f.Name = df.filename()
		f.Size = int64(binary.Size(df) + int(df.Size))
		f.CountryCode = df.countryCode()
		f.ProductCode = df.productCode()
		f.Identifier = df.identifier()

		r.File = append(r.File, f)
	}

	return nil
}

func (r *Reader) initFileList() {
	r.fileListOnce.Do(func() {
		files := make(map[string]int)

		for _, file := range r.File {
			name := file.Name

			if idx, ok := files[name]; ok {
				r.fileList[idx].isDup = true

				continue
			}

			idx := len(r.fileList)
			entry := fileListEntry{
				name:  name,
				file:  file,
				isDir: false,
			}
			r.fileList = append(r.fileList, entry)
			files[name] = idx
		}

		sort.Slice(r.fileList, func(i, j int) bool { return fileEntryLess(r.fileList[i].name, r.fileList[j].name) })
	})
}

func fileEntryLess(x, y string) bool {
	xdir, xelem := split(x)
	ydir, yelem := split(y)

	return xdir < ydir || xdir == ydir && xelem < yelem
}

func split(name string) (string, string) {
	if len(name) > 0 && name[len(name)-1] == '/' {
		name = name[:len(name)-1]
	}

	i := len(name) - 1
	for i >= 0 && name[i] != '/' {
		i--
	}

	if i < 0 {
		return ".", name
	}

	return name[:i], name[i+1:]
}

var dotFile = &fileListEntry{name: "./", isDir: true}

func (r *Reader) openLookup(name string) *fileListEntry {
	if name == "." {
		return dotFile
	}

	dir, elem := split(name)

	files := r.fileList
	i := sort.Search(len(files), func(i int) bool {
		idir, ielem := split(files[i].name)

		return idir > dir || idir == dir && ielem >= elem
	})

	if i < len(files) {
		fname := files[i].name
		if fname == name || len(fname) == len(name)+1 && fname[len(name)] == '/' && fname[:len(name)] == name {
			return &files[i]
		}
	}

	return nil
}

func (r *Reader) openReadDir(dir string) []fileListEntry {
	files := r.fileList

	i := sort.Search(len(files), func(i int) bool {
		idir, _ := split(files[i].name)

		return idir >= dir
	})

	j := sort.Search(len(files), func(j int) bool {
		jdir, _ := split(files[j].name)

		return jdir > dir
	})

	return files[i:j]
}

type openDir struct {
	e      *fileListEntry
	files  []fileListEntry
	offset int
}

func (d *openDir) Close() error               { return nil }
func (d *openDir) Stat() (fs.FileInfo, error) { return d.e.stat() }

func (d *openDir) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: d.e.name, Err: errors.New("is a directory")} //nolint:goerr113
}

func (d *openDir) ReadDir(count int) ([]fs.DirEntry, error) {
	n := len(d.files) - d.offset
	if count > 0 && n > count {
		n = count
	}

	if n == 0 {
		if count <= 0 {
			return nil, nil
		}

		return nil, io.EOF
	}

	list := make([]fs.DirEntry, n)
	for i := range list {
		s, err := d.files[d.offset+i].stat()
		if err != nil {
			return nil, err
		}

		list[i] = s
	}

	d.offset += n

	return list, nil
}

// Open opens the named file in the memory card image, using the semantics of
// fs.FS.Open: paths are always slash separated, with no leading / or ../
// elements.
func (r *Reader) Open(name string) (fs.File, error) {
	r.initFileList()

	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}

	e := r.openLookup(name)
	if e == nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	if e.isDir {
		return &openDir{e, r.openReadDir(name), 0}, nil
	}

	return e.file.Open()
}

// A ReadCloser is a Reader that must be closed when no longer needed.
type ReadCloser struct {
	Reader
	f *os.File
}

// Close closes the memory card image, rendering it unusable for I/O.
func (rc *ReadCloser) Close() error {
	if err := rc.f.Close(); err != nil {
		return fmt.Errorf("unable to close: %w", err)
	}

	return nil
}

// NewReader returns a new Reader reading from r.
func NewReader(r io.Reader) (*Reader, error) {
	mcr := new(Reader)
	if err := mcr.init(r); err != nil {
		return nil, err
	}

	return mcr, nil
}

// OpenReader will open the memory card image specified by name and return a
// ReadCloser.
func OpenReader(name string) (*ReadCloser, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("unable to open: %w", err)
	}

	r := new(ReadCloser)
	if err := r.init(f); err != nil {
		f.Close()

		return nil, err
	}

	r.f = f

	return r, nil
}
