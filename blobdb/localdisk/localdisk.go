package localdisk

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path"

	"github.com/rwcarlsen/cas/blobdb"
)

var (
	DupContentErr  = errors.New("blobdb: blob hash-content combo already exist")
	HashCollideErr = errors.New("blobdb: blob hash collision")
)

type Dbase struct {
	location string
}

func New(loc string) (db *Dbase, err error) {
	var mode os.FileMode = 0744
	if os.MkdirAll(loc, mode); err != nil {
		return nil, err
	}
	return &Dbase{location: loc}, nil
}

func (db *Dbase) Get(ref string) (r io.ReadCloser, err error) {
	p := path.Join(db.location, ref)
	return os.Open(p)
}

func (db *Dbase) Put(r io.Reader) (ref string, n int64, err error) {
	var buf1, buf2 bytes.Buffer
	mw := io.MultiWriter(&buf1, &buf2)
	if _, err := io.Copy(mw, r); err != nil {
		return "", 0, err
	}
	ref = blobdb.MakeBlobRef(&buf1)

	p := path.Join(db.location, ref)
	f, err := os.Create(p)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	n, err = io.Copy(f, &buf2)
	if err != nil {
		return "", 0, err
	}
	return ref, n, nil
}

func (db *Dbase) Enumerate(after string, limit int) ([]string, error) {
	f, err := os.Open(db.location)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	remain := limit
	refs := []string{}
	for remain > 0 {
		names, err := f.Readdirnames(32768)
		if err == io.EOF {
			return refs, nil
		} else if err != nil {
			return nil, err
		}

		if after != "" {
			for i, name := range names {
				if name >= after {
					break
				}
				names = names[i:]
			}
		}

		refs = append(refs, names...)
		remain -= len(names)
	}

	return refs, nil
}
