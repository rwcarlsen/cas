
package blob

import (
  "crypto/sha256"
  "crypto"
  "encoding/hex"
)

const (
  NameHashSep = "-"
  DefaultHash = crypto.SHA512
)

var (
  hash2Name = map[crypto.Hash]string {
    crypto.SHA512: "sha512",
  }

  name2Hash := make(map[string]crypto.Hash)
  for h, n := range hash2Name {
    name2Hash[n] = h
  }
)

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
  return &Blob{hash: DefaultHash, Content: content}
}

func (b *Blob) Sum() []byte {
  hsh := b.h.New()
  hsh.Write(b.content)
  return b.hashFunc.Sum([]byte{})
}

func (b *Blob) Ref() string {
  return hashName[b.Hash] + NameHashSep + hex.EncodeToString(b.Sum())
}

func (b *Blob) String() string {
  return b.Ref() + ":\n" +  string(b.Content)
}
