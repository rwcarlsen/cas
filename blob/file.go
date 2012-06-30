package blob

import (
  "os"
  "path/filepath"
  "io/ioutil"
)

// NewFileMeta creates a map containing meta-data for a file
// at the specified path.
func NewFileMeta(path string) (meta MetaData, err error) {
  f, err := os.Open(path)
  if err != nil {
    return nil, err
  }

  abs, _ := filepath.Abs(path)
  stat, err := f.Stat()
  if err != nil {
    return nil, err
  }

  meta = NewMeta("file")
  meta["name"] = stat.Name()
  meta["path"] = abs
  meta["size"] = stat.Size()
  meta["mod-time"] = stat.ModTime().UTC()

  return meta, nil
}

func FileBlobs(path string) (blobs []*Blob, err error) {
  f, err := os.Open(path)
  if err != nil {
    return nil, err
  }

  data, err := ioutil.ReadAll(f)
  if err != nil {
    return nil, err
  }

  chunks := SplitRaw(data, DefaultChunkSize)

  return chunks, nil
}

func FileBlobsAndMeta(path string) (meta MetaData, blobs []*Blob, err error) {
  meta, err = NewFileMeta(path)
  if err != nil {
    return nil, nil, err
  }

  blobs, err = FileBlobs(path)
  if err != nil {
    return nil, nil, err
  }

  meta.AttachRefs(RefsFor(blobs)...)
  return
}

func DirBlobsAndMeta(path string) (metas []MetaData, blobs []*Blob, err error) {
  blobs = make([]*Blob, 0)
  metas = make([]MetaData, 0)

  walkFn := func(path string, info os.FileInfo, inerr error) error {
    if info.IsDir() {
      return nil
    }
    meta, newblobs, err := FileBlobsAndMeta(path)
    if err != nil {
      return err
    }
    blobs = append(blobs, newblobs...)
    metas = append(metas, meta)
    return nil
  }

  err = filepath.Walk(path, walkFn)
  return metas, blobs, err
}

