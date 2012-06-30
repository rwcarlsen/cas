
package blob

import (
  "crypto"
  "crypto/rand"
  "crypto/sha256"
  "crypto/sha512"
  "encoding/hex"
  "encoding/json"
  "time"
)

const (
  NameHashSep = "-"
  DefaultHash = crypto.SHA256
  DefaultChunkSize = 1048576 // in bytes
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

// SplitRaw creates blobs by splitting data into blockSize (bytes) chunks
func SplitRaw(data []byte, blockSize int) []*Blob {
  blobs := make([]*Blob, 0)
  for i := 0; i < len(data); i += blockSize {
    end := min(len(data), i + blockSize)
    blobs = append(blobs, Raw(data[i:end]))
  }
  return blobs
}

// Object creates an immutable timestamped blob that can be used to
// simulate mutable objects that have a dynamic, pruneable revision
// history.
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

// Combine reconstitutes split data into a single byte slice
func Combine(blobs ...*Blob) []byte {
  data := make([]byte, 0)

  for _, b := range blobs {
    data = append(data, b.Content...)
  }
  return data
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
      smallest = val
    }
  }
  return smallest
}

