
package blob

import (
  "crypto"
  "crypto/rand"
  "crypto/sha256"
  "crypto/sha512"
  "encoding/hex"
  "encoding/json"
  "os"
  "io/ioutil"
  "path/filepath"
  "time"
)

const (
  NameHashSep = "-"
  DefaultHash = crypto.SHA256
  DefaultChunkSize = 1048576
)

var (
  hash2Name = map[crypto.Hash]string { }
  name2Hash = map[string]crypto.Hash { }
)

func init() {
  hash2Name[crypto.SHA256] = "sha256"
  hash2Name[crypto.SHA512] = "sha512"
  crypto.RegisterHash(crypto.SHA256, sha256.New)
  crypto.RegisterHash(crypto.SHA512, sha512.New)
  for h, n := range hash2Name {
    name2Hash[n] = h
  }
}

func HashToName(h crypto.Hash) string {
  return hash2Name[h]
}

func NameToHash(n string) crypto.Hash {
  return name2Hash[n]
}

// standardized way to create key-value based blobs
type MetaData map[string] interface{}

func NewMeta(kind string) MetaData {
  m := MetaData{}
  m["blob-type"] = kind
  return m
}

func (m MetaData) SetObjectRef(ref string) {
  m["object-ref"] = ref
}

func (m MetaData) AttachRefs(refs ...string) {
  m["refs"] = refs
}

func (m MetaData) ToBlob() (b *Blob, err error) {
  m["timestamp"] = time.Now().UTC()
  data, err := json.Marshal(m)
  if err != nil {
    return nil, err
  }
  return Raw(data), nil
}

type Blob struct {
  Hash crypto.Hash
  Content []byte
}

// Raw creates a blob using the DefaultHash holding the passed content.
func Raw(content []byte) *Blob {
  return &Blob{Hash: DefaultHash, Content: content}
}

func SplitRaw(data []byte, blockSize int) []*Blob {
  blobs := make([]*Blob, 0)
  for i := 0; i < len(data); i += blockSize {
    end := min(len(data), i + blockSize)
    blobs = append(blobs, Raw(data[i:end]))
  }
  return blobs
}

func Object() (b *Blob, err error) {
  m := NewMeta("object")

  r := make([]byte, 100)
  if _, err = rand.Read(r); err != nil {
    return nil, err
  }
  m["random"] = r

  return m.ToBlob()
}

func (b *Blob) GetMeta() (meta MetaData, err error) {
  err = json.Unmarshal(b.Content, &meta)
  return
}

func (b *Blob) Sum() []byte {
  hsh := b.Hash.New()
  hsh.Write(b.Content)
  return hsh.Sum([]byte{})
}

func (b *Blob) Ref() string {
  return HashToName(b.Hash) + NameHashSep + hex.EncodeToString(b.Sum())
}

func (b *Blob) String() string {
  return b.Ref() + ":\n" +  string(b.Content)
}

func RefsFor(blobs []*Blob) []string {
  refs := make([]string, len(blobs))
  for i, b := range blobs {
    refs[i] = b.Ref()
  }
  return refs
}

func min(vals ...int) int {
  smallest := vals[0]
  for _, val := range vals[1:] {
    if val < smallest {
      val = smallest
    }
  }
  return smallest
}

//////////////////////////////////////////////////////////////
///////// for creating specific types of blobs////////////////
//////////////////////////////////////////////////////////////

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

