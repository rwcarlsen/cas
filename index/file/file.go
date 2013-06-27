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

// PutPath calls PutReader with the file at the specified path as the
// io.Reader.
func (s *Store) PutPath(path string) (blobref string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	abs, err := filepath.Abs(path)
	return s.PutReader(filepath.ToSlash(abs), f)
}

// PutReader creates dumps the data from r as a blob into the store
// database and adds file-based meta-data attributes to the index.
// path is the value of the file.Path attribute
func (s *Store) PutReader(path string, r io.Reader) (blobref string, err error) {
	ref, n, err := s.Db.Put(r)
	if err != nil {
		return "", err
	}

	s.Index.Set(blobref, Size, fmt.Sprint(n))
	s.Index.Set(blobref, Path, path)

	return ref, nil
}

// GetBytes returns the most recent blobref+data that has ever had the
// specified path.  The blobref and data returned may no longer have the
// same path.
func (s *Store) GetBytes(path string) (blobref string, data []byte, err error) {
	refs, err := s.Index.FindExact(Path, path, 1)
	if err != nil {
		return nil, err
	} else if len(refs) != 1 {
		return nil, fmt.Errorf("file: path %v not found in index")
	}

	return blobdb.GetBytes(s.Db, refs[0])
}
