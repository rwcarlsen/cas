package file

import (
	"os"
	"io"
	"time"
	"bytes"
	"path/filepath"

	"github.com/rwcarlsen/cas/schema"
	"github.com/rwcarlsen/cas/blobdb"
)

const Property = "file"

type Info struct {
	Created time.Time
	Size int64
	Path string
	ContentRefs []string
}

func PutPath(db blobdb.Interface, path string) (*Info, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return PutReader(db, path, f)
}

func PutReader(db blobdb.Interface, path string, r io.Reader) (*Info, error) {
	ref, n, err := db.Put(r)
	if err != nil {
		return nil, err
	}

	obj := schema.NewObject("")
	objref, _, err := db.Put(obj)
	if err != nil {
		return nil, err
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	fi := &Info{
		Created: time.Now(),
		Size: n,
		Path: filepath.ToSlash(abs),
		ContentRefs: []string{ref},
	}

	meta := schema.NewMeta(objref)
	meta.Props[Property] = fi

	data, err := schema.Marshal(meta)
	if err != nil {
		panic(err)
	}

	ref, _, err = db.Put(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	return fi, nil
}

func Get(db blobdb.Interface, metaref string) (data []byte, f *Info, err error) {
	data, err = blobdb.GetData(db, metaref)
	if err != nil {
		return nil, nil, err
	}

	fi := &Info{}
	if err = schema.UnmarshalProp(data, Property, fi); err != nil {
		return nil, nil, err
	}

	content := make([]byte, 0, fi.Size)
	for _, ref := range fi.ContentRefs {
		data, err = blobdb.GetData(db, ref)
		if err != nil {
			return nil, nil, err
		}
		content = append(content, data...)
	}
	return content, fi, nil
}



