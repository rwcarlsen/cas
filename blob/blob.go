
package blob

import (
  "crypto"
  "crypto/sha512"
  "encoding/hex"
)

const (
  NameHashSep = "-"
  DefaultHash = crypto.SHA512
)

var (
  hash2Name = map[crypto.Hash]string {crypto.SHA512: "sha512"}
  name2Hash = map[string]crypto.Hash { }
)

func init() {
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

func New(content []byte) *Blob {
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
