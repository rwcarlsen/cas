
package blob

import (
  "os"
  "path/filepath"
  "io/ioutil"
  "time"
)

const (
  defaultChunkSize = 1048576 // in bytes
)

type FileMeta struct {
  RcasType Type "rcasType"
  Name string
  Path string
  Size int64
  ModTime time.Time
  ContentRefs []string
}

// NewFileMeta creates a map containing meta-data for a file
// at the specified path.
func NewFileMeta() *FileMeta {
  return &FileMeta{
    RcasType: FileType,
    ContentRefs: make([]string, 0),
  }
}

// AddContentRefs adds refs for blobs that constitute this file's bytes
func (m *FileMeta) AddContentRefs(refs ...string) {
  m.ContentRefs = append(m.ContentRefs, refs...)
}

// LoadFromPath fills in all meta fields (name, size, mod time, ...) by reading
// the info from the file located at path. Blobs constituting the file's bytes
// are returned. AddContentRefs is invoked for all the blobs returned.
func (m *FileMeta) LoadFromPath(path string) (chunks []*Blob, err error) {
  f, err := os.Open(m.Path)
  if err != nil {
    return nil, err
  }

  data, err := ioutil.ReadAll(f)
  if err != nil {
    return nil, err
  }

  chunks = SplitRaw(data, defaultChunkSize)

  // fill in meta fields
  abs, _ := filepath.Abs(path)
  stat, err := f.Stat()
  if err != nil {
    return nil, err
  }

  m.Name = stat.Name()
  m.Path = abs
  m.Size = stat.Size()
  m.ModTime = stat.ModTime().UTC()
  m.AddContentRefs(RefsFor(chunks)...)

  return chunks, nil
}

// SplitFile creates blobs by splitting data into blockSize (bytes) chunks
func SplitRaw(data []byte, blockSize int) []*Blob {
  blobs := make([]*Blob, 0)
  for i := 0; i < len(data); i += blockSize {
    end := min(len(data), i + blockSize)
    blobs = append(blobs, NewRaw(data[i:end]))
  }
  return blobs
}

// CombineFile reconstitutes split data into a single byte slice
func CombineRaw(blobs ...*Blob) []byte {
  data := make([]byte, 0)

  for _, b := range blobs {
    data = append(data, b.Content...)
  }
  return data
}

func DirBlobsAndMeta(path string) (metas []*FileMeta, blobs []*Blob, err error) {
  blobs = make([]*Blob, 0)
  metas = make([]*FileMeta, 0)

  walkFn := func(path string, info os.FileInfo, inerr error) error {
    if info.IsDir() {
      return nil
    }

    meta := NewFileMeta()
    newblobs, err := meta.LoadFromPath(path)
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

func min(vals ...int) int {
  smallest := vals[0]
  for _, val := range vals[1:] {
    if val < smallest {
      smallest = val
    }
  }
  return smallest
}

