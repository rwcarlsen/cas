package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rwcarlsen/cas/blobdb"
	"github.com/rwcarlsen/cas/index"
)

// file meta-data attribute keys
const (
	Size = "file-size"
	Path = "file-path"
)

type Store struct {
	Db    blobdb.Interface
	Index *index.Index
}

func (s *Store) PutPath(path string) (blobref string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	abs, err := filepath.Abs(path)
	return s.PutReader(filepath.ToSlash(abs), f)
}

func (s *Store) PutReader(path string, r io.Reader) (blobref string, err error) {
	ref, n, err := s.Db.Put(r)
	if err != nil {
		return "", err
	}

	s.Index.Set(blobref, Size, fmt.Sprint(n))
	s.Index.Set(blobref, Path, path)

	return ref, nil
}
