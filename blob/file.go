package blob

import (
  "os"
  "path/filepath"
  "io/ioutil"
)

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

func PlainFileBlobs(path string) (blobs []*Blob, err error) {
  meta, err := NewFileMeta(path)
  if err != nil {
    return nil, err
  }

  blobs, err = FileBlobs(path)
  if err != nil {
    return nil, err
  }

  meta.AttachRefs(RefsFor(blobs)...)

  m, err := meta.ToBlob()
  if err != nil {
    return nil, err
  }
  return append(blobs, m), nil
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

