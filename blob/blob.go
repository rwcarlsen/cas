
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
)

const (
  NameHashSep = "-"
  DefaultHash = crypto.SHA256
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
func NewMeta(kind, objectRef string) MetaData {
  m := MetaData{}
  m["blob-type"] = t
  m["object-ref"] = ref
  m["timestamp"] = time.Now().UTC().String()
  return m
}

type Blob struct {
  Hash crypto.Hash
  Content []byte
}

func Object() *Blob {
  m := NewMeta("object", "")

  r := make([]byte, 100)
  _, err := rand.Read(r)
  if err != nil {
    return nil, err
  }
  m["random"] = r

  data, err := json.Marshal(m)
  if err != nil {
    return nil, err
  }
}

// Raw creates a blob using the DefaultHash holding the passed content.
func Raw(content []byte) *Blob {
  return &Blob{Hash: DefaultHash, Content: content}
}

func File(path string) (file, metadata *Blob, err error) {
  f, err := os.Open(path)
  if err != nil {
    return nil, nil, err
  }

  data, err := ioutil.ReadAll(f)
  if err != nil {
    return nil, nil, err
  }

  stat, err := f.Stat()
  if err != nil {
    return nil, nil, err
  }

  meta := NewMetaData("file", ref)
  meta["name"] = stat.Name()
  abs, _ := filepath.Abs(path)
  meta["path"] = abs
  meta["size"] = stat.Size()
  meta["mod-time"] = stat.ModTime()

  b := Raw(data)
  p, err := Pointer(b.Ref(), meta)
  if err != nil {
    return nil, nil, err
  }
  return b, p, nil
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

