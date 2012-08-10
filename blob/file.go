
package blob

import (
  "os"
  "errors"
  "encoding/json"
  "path/filepath"
  "io/ioutil"
)

const (
  DefaultChunkSize = 1 << 24 // 16Mb
)

type FileMeta struct {
  RcasType string
  RcasObjectRef string
  Name string
  Notes map[string]string
  Size int64
  ContentRefs []string
}

// NewFileMeta creates a map containing meta-data for a file
// at the specified path.
func NewFileMeta() *FileMeta {
  return &FileMeta{
    RcasType: File,
    ContentRefs: make([]string, 0),
    Notes: make(map[string]string),
  }
}

// LoadFromPath fills in all meta fields (name, size, etc. by reading
// the info from the file located at path. Blobs constituting the file's bytes
// are returned. AddContentRefs is invoked for all the blobs returned.
func (m *FileMeta) LoadFromPath(path string) ([]*Blob, error) {
  f, err := os.Open(path)
  if err != nil {
    return nil, err
  }

  data, err := ioutil.ReadAll(f)
  if err != nil {
    return nil, err
  }

  chunks := SplitRaw(data, DefaultChunkSize)

  // fill in meta fields
  abs, _ := filepath.Abs(path)
  abs = filepath.Dir(abs)
  stat, err := f.Stat()
  if err != nil {
    return nil, err
  }

  m.Name = stat.Name()
  m.Size = stat.Size()
  m.ContentRefs = RefsFor(chunks)

  return chunks, nil
}

// AddNotes allows arbitrary meta-data to be attached to any file.
//
// This should be used by apps to make valueable meta-data accessible to any app
// that tries to use/find the file.
func (m *FileMeta) SetNotes(id string, v interface{}) error {
  data, err := json.Marshal(v)
  if err != nil {
    return errors.New("blob: failed to marshal notes into json")
  }
  m.Notes[id] = string(data)
  return nil
}

// AddNotes allows arbitrary meta-data to be attached to any file.
//
// This should be used by apps to make valueable meta-data accessible to any app
// that tries to use/find the file.
func (m *FileMeta) GetNotes(id string, v interface{}) error {
  if s, ok := m.Notes[id]; ok {
    err := json.Unmarshal([]byte(s), v)
    if err != nil {
      return errors.New("blob: failed to unmarshal json notes:" + err.Error())
    }
    return nil
  }
  return errors.New("blob: invalid notes id")
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
func Reconstitute(blobs ...*Blob) []byte {
  data := make([]byte, 0)

  for _, b := range blobs {
    data = append(data, b.Content()...)
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

