package localdisk

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path"
	"sort"

	"github.com/rwcarlsen/cas/blobdb"
)

func init() {
	blobdb.Register("localdisk", creator)
}

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
		return ref, 0, err
	}
	defer f.Close()

	n, err = io.Copy(f, &buf2)
	if err != nil {
		return ref, 0, err
	}
	return ref, n, nil
}

func (db *Dbase) Enumerate(after string, limit int) ([]string, error) {
	f, err := os.Open(db.location)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	sort.Strings(names)
	if after != "" {
		i := sort.SearchStrings(names, after)
		names = append([]string{}, names[i+1:]...)
	}
	return names, nil
}

func creator(params blobdb.Params) (blobdb.Interface, error) {
	root, ok := params["Root"]
	if !ok {
		return nil, errors.New("localdisk: missing 'Root' from Params")
	}
	return &Dbase{root}, nil
}
