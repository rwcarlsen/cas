
package blob

import (
  "crypto/sha256"
  "hash"
)

type blob struct {
  h hash.Hash
}

func New() *blob {
  h := sha256.New()
  return &blob{h:h}
}

func (b *blob) Write(data []byte) {
  b.h.Write(data)
}

func (b *blob) Sum() []byte {
  sum := b.h.Sum([]byte{})
  return sum
}

