
package blob

import (
  "crypto/sha256"
  "hash"
  "encoding/hex"
)

type blob struct {
  hashFunc hash.Hash
  hashName string
  content []byte
}

func New() *blob {
  h := sha256.New()
  name := "sha256"
  return &blob{hashFunc:h, hashName: name}
}

func (b *blob) Write(data []byte) (n int, err error) {
  n = len(data)
  err = nil

  b.content = append(b.content, data...)
  b.hashFunc.Write(data)

  return
}

func (b *blob) Sum() []byte {
  sum := b.hashFunc.Sum([]byte{})
  return sum
}

func (b *blob) FileName() string {
  return b.hashName + "-" + hex.EncodeToString(b.Sum())
}

func (b *blob) Content() []byte {
  return b.content
}

