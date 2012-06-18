
package blob

import (
  "crypto"
  "crypto/sha256"
  "crypto/sha512"
  "encoding/hex"
  "encoding/json"
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

type Blob struct {
  Hash crypto.Hash
  Content []byte
}

func Raw(content []byte) *Blob {
  return &Blob{Hash: DefaultHash, Content: content}
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

type MetaData map[string] interface{}

func Pointer(ref string, meta MetaData) (b *Blob, err error) {
  m := MetaData(meta)
  m["pointsTo"] = ref
  data, err := json.Marshal(m)
  if err != nil {
    return nil, err
  }
  return Raw(data), nil
}

