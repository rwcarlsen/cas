package file

import (
	"os"
	"io"
	"io/ioutil"
	"time"
	"bytes"
	"path/filepath"

	"github.com/rwcarlsen/cas/schema"
	"github.com/rwcarlsen/cas/blobdb"
)

const Property = "file"

type File struct {
	Created time.Time
	Size int
	Path string
	ContentRefs []string
}

func PutPath(db blobdb.Interface, path string) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return PutReader(db, path, f)
}

func PutReader(db blobdb.Interface, path string, r io.Reader) (*File, error) {
	ref, n, err := db.Put(r)
	if err != nil {
		return nil, err
	}

	obj := schema.NewObject("")
	objref, _, err := db.Put(obj)
	if err != nil {
		return nil, err
	}

	meta := schema.NewMeta(objref)
	meta.Props[Property] =  &File{
		Created: time.Now(),
		Size: n,
		Path: filepath.ToSlash(filepath.Abs(path)),
		ContentRefs: []string{ref},
	}

	data, err := schema.Marshal(meta)
	if err != nil {
		panic(err)
	}

	ref, _, err = db.Put(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	return meta.Props[Property], nil
}

func Get(db blobdb.Interface, metaref string) (data []byte, f *File, err error) {
	r, err := db.Get(metaref)
	if err != nil {
		return nil, nil, err
	}

	data, err = ioutil.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}

	fi := &File{}
	if err = schema.UnmarshalProp(data, fi); err != nil {
		return nil, nil, err
	}

	content := make([]byte, 0, fi.Size)
	for _, ref := range fi.ContentRefs {
		r, err = db.Get(ref)
		if err != nil {
			return nil, nil, err
		}

		data, err = ioutil.ReadAll(r)
		if err != nil {
			return nil, nil, err
		}

		content = append(content, data...)
	}
	return content, fi, nil
}



