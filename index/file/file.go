package file

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rwcarlsen/cas/blobdb"
	"github.com/rwcarlsen/cas/index"
)

// file meta-data attributes
const (
	Size = "file-size"
	Path = "file-path"
)

func PutPath(db blobdb.Interface, i *index.Index, path string) (ref string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	abs, err := filepath.Abs(path)
	return PutReader(db, i, filepath.ToSlash(abs), f)
}

func PutReader(db blobdb.Interface, i *index.Index, path string, r io.Reader) (ref string, err error) {
	ref, n, err := db.Put(r)
	if err != nil {
		return "", err
	}

	i.Set(blobref, Size, fmt.Sprint(n))
	i.Set(blobref, Path, path)
	return ref, nil
}
